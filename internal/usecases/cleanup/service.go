// internal/usecases/cleanup/service.go
package cleanup

import (
	"context"
	"fmt"
	"go-image-cleanup/internal/domain/models"
	"go-image-cleanup/internal/domain/notification"
	"go-image-cleanup/internal/domain/repositories"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

type CleanupService struct {
    repo       repositories.ImageRepository
    notifier   notification.Notifier
    logger     *zap.Logger
    timeout    time.Duration
    workerPool int
}

func NewCleanupService(repo repositories.ImageRepository, notifier notification.Notifier, logger *zap.Logger) *CleanupService {
    return &CleanupService{
        repo:       repo,
        notifier:   notifier,
        logger:     logger,
        timeout:    5 * time.Minute, // Configurable timeout
        workerPool: 5,              // Configurable worker pool size
    }
}


func (s *CleanupService) removeImagesInParallel(ctx context.Context, images []models.Image, usedImages map[string]bool) (int, int) {
    var (
        wg sync.WaitGroup
        mu sync.Mutex // Protects access to counters
        removed int
        skipped int
    )

    // Create job channel
    jobs := make(chan models.Image, len(images))

    // Start worker pool
    for i := 0; i < s.workerPool; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for img := range jobs {
                select {
                case <-ctx.Done():
                    return
                default:
                    if usedImages[img.ID] {
                        mu.Lock()
                        skipped++
                        mu.Unlock()
                        s.logger.Info("Skipping image in use",
                            zap.String("id", img.ID),
                            zap.Strings("tags", img.Tags))
                        continue
                    }

                    if err := s.repo.RemoveImage(ctx, img.ID); err != nil {
                        mu.Lock()
                        skipped++
                        mu.Unlock()
                        s.logger.Error("Failed to remove image",
                            zap.String("id", img.ID),
                            zap.Error(err))
                        continue
                    }

                    mu.Lock()
                    removed++
                    mu.Unlock()
                    s.logger.Info("Successfully removed image",
                        zap.String("id", img.ID),
                        zap.Strings("tags", img.Tags))
                }
            }
        }()
    }

    // Send jobs
    for _, img := range images {
        select {
        case jobs <- img:
        case <-ctx.Done():
            return removed, skipped
        }
    }
    close(jobs)

    // Wait for all workers to complete
    wg.Wait()

    return removed, skipped
}



// getHostInfo returns hostname and IP addresses
func (s *CleanupService) getHostInfo() (string, string, error) {
    hostname, err := os.Hostname()
    if err != nil {
        return "", "", fmt.Errorf("failed to get hostname: %w", err)
    }

    // Get all network interfaces
    interfaces, err := net.Interfaces()
    if err != nil {
        return hostname, "", fmt.Errorf("failed to get network interfaces: %w", err)
    }

    var ipAddresses []string
    for _, iface := range interfaces {
        // Skip loopback and interfaces that are down
        if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
            continue
        }

        addrs, err := iface.Addrs()
        if err != nil {
            continue
        }

        for _, addr := range addrs {
            // Convert address to string and extract IP
            ipNet, ok := addr.(*net.IPNet)
            if !ok {
                continue
            }
            ip := ipNet.IP.String()
            // Skip loopback and IPv6 addresses
            if !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
                ipAddresses = append(ipAddresses, ip)
            }
        }
    }

    return hostname, strings.Join(ipAddresses, ", "), nil
}

// internal/usecases/cleanup/service.go
func (s *CleanupService) Cleanup(ctx context.Context) error {
    // Create timeout context
    ctx, cancel := context.WithTimeout(ctx, s.timeout)
    defer cancel()

    // Create error channel
    errChan := make(chan error, 1)

    go func() {
        // Record start time in ICT+7
        startTime := time.Now().In(time.FixedZone("ICT", 7*60*60))

        images, err := s.repo.GetAllImages(ctx)
        if err != nil {
            errChan <- fmt.Errorf("failed to get images: %w", err)
            return
        }

        usedImages, err := s.repo.GetUsedImages(ctx)
        if err != nil {
            errChan <- fmt.Errorf("failed to get used images: %w", err)
            return
        }

        stats := struct {
            total    int
            removed  int
            skipped  int
        }{
            total: len(images),
        }

        // Remove images in parallel
        stats.removed, stats.skipped = s.removeImagesInParallel(ctx, images, usedImages)

        // Get host information
        hostname, ips, err := s.getHostInfo()
        hostInfo := "Unknown host"
        if err == nil {
            hostInfo = fmt.Sprintf("Host: %s\nIP(s): %s", hostname, ips)
        }

        // Calculate end time and duration
        endTime := time.Now().In(time.FixedZone("ICT", 7*60*60))
        duration := endTime.Sub(startTime)

        // Send notification with host information and ICT+7 timestamp
        message := fmt.Sprintf("Image cleanup completed on:\n%s\n\nStarted: %s\nFinished: %s\nDuration: %s\n\nResults:\n- Total: %d\n- Removed: %d\n- Skipped: %d",
            hostInfo,
            startTime.Format("2006-01-02 15:04:05 ICT"),
            endTime.Format("2006-01-02 15:04:05 ICT"),
            duration.Round(time.Second),
            stats.total,
            stats.removed,
            stats.skipped)

        if err := s.notifier.SendNotification(message); err != nil {
            s.logger.Error("Failed to send notification", zap.Error(err))
        }

        s.logger.Info("Cleanup completed",
            zap.Int("total", stats.total),
            zap.Int("removed", stats.removed),
            zap.Int("skipped", stats.skipped),
            zap.String("hostname", hostname),
            zap.String("ips", ips),
            zap.String("start_time", startTime.Format("2006-01-02 15:04:05 ICT")),
            zap.String("end_time", endTime.Format("2006-01-02 15:04:05 ICT")),
            zap.String("duration", duration.Round(time.Second).String()))

        errChan <- nil
    }()

    // Wait for completion or timeout
    select {
    case err := <-errChan:
        return err
    case <-ctx.Done():
        return fmt.Errorf("cleanup operation timed out after %v", s.timeout)
    }
}