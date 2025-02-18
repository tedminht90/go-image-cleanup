// internal/usecases/cleanup/service.go
package cleanup

import (
	"context"
	"fmt"
	"go-image-cleanup/internal/domain/metrics"
	"go-image-cleanup/internal/domain/models"
	"go-image-cleanup/internal/domain/notification"
	"go-image-cleanup/internal/domain/repositories"
	"go-image-cleanup/pkg/helper"
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
	metrics    metrics.MetricsCollector
	logger     *zap.Logger
	timeout    time.Duration
	workerPool int
}

func NewCleanupService(
	repo repositories.ImageRepository,
	notifier notification.Notifier,
	metrics metrics.MetricsCollector,
	logger *zap.Logger,
) *CleanupService {
	return &CleanupService{
		repo:       repo,
		notifier:   notifier,
		metrics:    metrics,
		logger:     logger,
		timeout:    5 * time.Minute, // Configurable timeout
		workerPool: 5,               // Configurable worker pool size
	}
}

func (s *CleanupService) removeImagesInParallel(ctx context.Context, images []models.Image, usedImages map[string]bool) (int, int) {
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex // Protects access to counters
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

func (s *CleanupService) Cleanup(ctx context.Context) error {
	startTime := helper.TimeInICT(time.Now())

	images, err := s.repo.GetAllImages(ctx)
	if err != nil {
		s.metrics.IncCleanupErrors()
		return fmt.Errorf("failed to get images: %w", err)
	}

	usedImages, err := s.repo.GetUsedImages(ctx)
	if err != nil {
		s.metrics.IncCleanupErrors()
		return fmt.Errorf("failed to get used images: %w", err)
	}

	stats := struct {
		total   int
		removed int
		skipped int
	}{
		total: len(images),
	}

	// Remove images in parallel
	stats.removed, stats.skipped = s.removeImagesInParallel(ctx, images, usedImages)

	// Update metrics
	for i := 0; i < stats.removed; i++ {
		s.metrics.IncImagesRemoved()
	}
	for i := 0; i < stats.skipped; i++ {
		s.metrics.IncImagesSkipped()
	}

	// Get host information
	hostname, ips, err := s.getHostInfo()
	hostInfo := "Unknown host"
	if err == nil {
		hostInfo = fmt.Sprintf("Host: %s\nIP(s): %s", hostname, ips)
	}

	endTime := helper.TimeInICT(time.Now())
	duration := endTime.Sub(startTime)

	// Update timing metrics
	s.metrics.SetLastCleanupTime(endTime)
	s.metrics.ObserveCleanupDuration(duration)

	// Send notification
	message := helper.FormatCleanupMessage(
		hostInfo,
		startTime,
		endTime,
		duration,
		stats.total,
		stats.removed,
		stats.skipped,
	)

	if err := s.notifier.SendNotification(message); err != nil {
		s.logger.Error("Failed to send notification", zap.Error(err))
		s.metrics.IncCleanupErrors()
	}

	s.logger.Info("Cleanup completed",
		zap.Int("total", stats.total),
		zap.Int("removed", stats.removed),
		zap.Int("skipped", stats.skipped),
		zap.String("hostname", hostname),
		zap.String("ips", ips),
		zap.String("start_time", helper.FormatICT(startTime)),
		zap.String("end_time", helper.FormatICT(endTime)),
		zap.String("duration", duration.String()))

	return nil
}
