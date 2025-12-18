package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	gpm "github.com/viperadnan-git/go-gpm"

	"github.com/urfave/cli/v3"
)

func uploadAction(ctx context.Context, cmd *cli.Command) error {
	filePath := cmd.StringArg("filepath")

	// Validate that filepath exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file or directory does not exist: %s", filePath)
	}

	// Load config
	if err := loadConfig(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get per-account settings
	account := cfgManager.GetSelectedAccount()
	var accountThreads int
	var accountQuality string
	var accountUseQuota bool
	if account != nil {
		accountThreads = account.UploadThreads
		accountQuality = account.Quality
		accountUseQuota = account.UseQuota
	}

	// Get CLI flags with fallback to account settings
	threads := int(cmd.Int("threads"))
	if threads == 0 {
		threads = accountThreads
	}
	if threads == 0 {
		threads = 3 // default
	}

	quality := cmd.String("quality")
	if quality == "" {
		quality = accountQuality
	}
	if quality == "" {
		quality = "original" // default
	}
	if quality != "original" && quality != "storage-saver" {
		return fmt.Errorf("invalid quality: %s (use 'original' or 'storage-saver')", quality)
	}

	albumName := cmd.String("album")

	// Parse datetime flag
	var timestamp *time.Time
	if datetimeStr := cmd.String("datetime"); datetimeStr != "" {
		var ts time.Time
		var err error
		if strings.ToLower(datetimeStr) == "now" {
			ts = time.Now()
		} else {
			ts, err = time.Parse(time.RFC3339, datetimeStr)
			if err != nil {
				return fmt.Errorf("invalid datetime format: %w", err)
			}
		}
		timestamp = &ts
	}

	// Build upload options from CLI flags
	uploadOpts := gpm.UploadOptions{
		Workers:         threads,
		Recursive:       cmd.Bool("recursive"),
		ForceUpload:     cmd.Bool("force"),
		DeleteFromHost:  cmd.Bool("delete"),
		DisableFilter:   cmd.Bool("disable-filter"),
		Caption:         cmd.String("caption"),
		ShouldFavourite: cmd.Bool("favourite"),
		ShouldArchive:   cmd.Bool("archive"),
		Quality:         quality,
		UseQuota:        cmd.Bool("use-quota") || accountUseQuota,
	}

	// Create API client
	api, err := createAPIClient()
	if err != nil {
		return err
	}

	// Handle --check mode (dry run)
	if cmd.Bool("check") {
		return checkFiles(ctx, api, filePath, threads, uploadOpts.Recursive, uploadOpts.DisableFilter)
	}

	// Log start
	logger.Info("scanning files", "path", filePath)

	// Track results
	var totalFiles, uploaded, existing, failed int
	var successfulMediaKeys []string

	// Process upload events
	for event := range api.Upload(ctx, filePath, uploadOpts) {
		if event.Total > 0 {
			totalFiles = event.Total
			logger.Info("starting upload", "files", totalFiles, "threads", threads)
		}

		switch event.Status {
		case gpm.StatusHashing, gpm.StatusUploading:
			logger.Debug(string(event.Status), "file", event.Path)
		case gpm.StatusCompleted:
			uploaded++
			progress := fmt.Sprintf("[%d/%d]", uploaded+existing+failed, totalFiles)
			logger.Info(progress+" uploaded", "mediaKey", event.MediaKey, "file", event.Path)
			if event.MediaKey != "" {
				successfulMediaKeys = append(successfulMediaKeys, event.MediaKey)
			}
		case gpm.StatusSkipped:
			existing++
			progress := fmt.Sprintf("[%d/%d]", uploaded+existing+failed, totalFiles)
			logger.Info(progress+" skipped", "mediaKey", event.MediaKey, "file", event.Path, "exists", true)
			if event.MediaKey != "" {
				successfulMediaKeys = append(successfulMediaKeys, event.MediaKey)
			}
		case gpm.StatusFailed:
			failed++
			progress := fmt.Sprintf("[%d/%d]", uploaded+existing+failed, totalFiles)
			logger.Error(progress+" failed", "file", event.Path, "error", event.Error)
		default:
			logger.Debug(string(event.Status), "file", event.Path, "mediaKey", event.MediaKey, "dedupKey", event.DedupKey, "error", event.Error)
		}
	}

	// Print summary
	logger.Info("upload complete", "uploaded", uploaded, "skipped", existing, "failed", failed)

	// Handle album creation if album name was specified
	if albumName != "" && len(successfulMediaKeys) > 0 {
		const batchSize = 500
		var albumKey string

		// Check if album already exists in mappings
		if existingKey := cfgManager.GetAlbumKey(albumName); existingKey != "" {
			albumKey = existingKey
			logger.Info("using existing album", "album", albumName, "key", albumKey)
			// Add all media to existing album in batches
			for i := 0; i < len(successfulMediaKeys); i += batchSize {
				end := min(i+batchSize, len(successfulMediaKeys))
				if err := api.AddMediaToAlbum(ctx, albumKey, successfulMediaKeys[i:end]); err != nil {
					logger.Warn("failed to add batch to album", "error", err)
				}
			}
		} else {
			logger.Info("creating album", "album", albumName)
			firstBatchEnd := min(batchSize, len(successfulMediaKeys))

			var err error
			albumKey, err = api.CreateAlbum(ctx, albumName, successfulMediaKeys[:firstBatchEnd])
			if err != nil {
				return fmt.Errorf("failed to create album: %w", err)
			}

			// Auto-store the mapping for future use
			if err := cfgManager.SetAlbumMapping(albumName, albumKey); err != nil {
				logger.Warn("failed to store album mapping", "error", err)
			}

			for i := batchSize; i < len(successfulMediaKeys); i += batchSize {
				end := min(i+batchSize, len(successfulMediaKeys))
				if err = api.AddMediaToAlbum(ctx, albumKey, successfulMediaKeys[i:end]); err != nil {
					logger.Warn("failed to add batch to album", "error", err)
				}
			}
		}

		logger.Info("album ready", "album", albumName, "items", len(successfulMediaKeys))
	}

	// Handle datetime setting if timestamp was specified
	if timestamp != nil && len(successfulMediaKeys) > 0 {
		logger.Info("setting datetime", "count", len(successfulMediaKeys), "datetime", timestamp.Format(time.RFC3339))

		const batchSize = 500
		for i := 0; i < len(successfulMediaKeys); i += batchSize {
			end := min(i+batchSize, len(successfulMediaKeys))
			if err := api.SetDateTime(ctx, successfulMediaKeys[i:end], *timestamp); err != nil {
				logger.Warn("failed to set datetime for batch", "error", err)
			}
		}

		logger.Info("datetime set successfully", "count", len(successfulMediaKeys))
	}

	return nil
}

func checkFiles(ctx context.Context, api *gpm.GooglePhotosAPI, path string, threads int, recursive, disableFilter bool) error {
	logger.Info("scanning files", "path", path)

	files, err := gpm.GetGooglePhotosSupportedFiles(path, recursive, disableFilter)
	if err != nil {
		return fmt.Errorf("failed to scan files: %w", err)
	}
	if len(files) == 0 {
		logger.Info("no supported files found")
		return nil
	}

	totalFiles := len(files)
	workers := min(threads, totalFiles)
	logger.Info("starting check", "files", totalFiles, "threads", workers)

	var wouldUpload, exists, failed atomic.Int32
	var processed atomic.Int32

	workChan := make(chan string, len(files))
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filePath := range workChan {
				select {
				case <-ctx.Done():
					return
				default:
				}

				hash, err := gpm.CalculateSHA1(ctx, filePath)
				if err != nil {
					failed.Add(1)
					count := processed.Add(1)
					logger.Error(fmt.Sprintf("[%d/%d] failed", count, totalFiles), "file", filePath, "error", err)
					continue
				}

				mediaKey, _ := api.FindMediaKeyByHash(ctx, hash)
				count := processed.Add(1)
				if mediaKey != "" {
					exists.Add(1)
					logger.Info(fmt.Sprintf("[%d/%d] exists", count, totalFiles), "mediaKey", mediaKey, "file", filePath)
				} else {
					wouldUpload.Add(1)
					logger.Info(fmt.Sprintf("[%d/%d] would upload", count, totalFiles), "file", filePath)
				}
			}
		}()
	}

	for _, f := range files {
		select {
		case <-ctx.Done():
			close(workChan)
			wg.Wait()
			return ctx.Err()
		case workChan <- f:
		}
	}
	close(workChan)
	wg.Wait()

	logger.Info("check complete", "would_upload", wouldUpload.Load(), "exists", exists.Load(), "failed", failed.Load())
	return nil
}
