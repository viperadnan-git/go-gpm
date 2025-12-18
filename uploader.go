package gpm

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/viperadnan-git/go-gpm/internal/core"
)

// UploadStatus represents the state of a file upload
type UploadStatus string

const (
	StatusHashing    UploadStatus = "hashing"
	StatusChecking   UploadStatus = "checking"
	StatusUploading  UploadStatus = "uploading"
	StatusFinalizing UploadStatus = "finalizing"
	StatusCompleted  UploadStatus = "completed"
	StatusSkipped    UploadStatus = "skipped" // Already in library
	StatusFailed     UploadStatus = "failed"
)

// UploadEvent represents a status update for a file upload
type UploadEvent struct {
	Path     string
	Status   UploadStatus
	MediaKey string
	DedupKey string
	Error    error
	WorkerID int
	Total    int // Total files in batch (set on first event)
}

// UploadOptions contains runtime options for upload operations
type UploadOptions struct {
	Workers         int
	Recursive       bool
	ForceUpload     bool
	DeleteFromHost  bool
	DisableFilter   bool
	Caption         string
	ShouldFavourite bool
	ShouldArchive   bool
	Quality         string // "original" or "storage-saver"
	UseQuota        bool
}

// Upload uploads files to Google Photos and returns a channel for status events.
// The channel is closed when upload completes. Multiple calls are queued automatically.
func (g *GooglePhotosAPI) Upload(ctx context.Context, path string, opts UploadOptions) <-chan UploadEvent {
	events := make(chan UploadEvent)

	go func() {
		// Serialize upload batches
		g.uploadMu.Lock()
		defer g.uploadMu.Unlock()
		defer close(events)

		// Filter files
		files, err := GetGooglePhotosSupportedFiles(path, opts.Recursive, opts.DisableFilter)
		if err != nil {
			events <- UploadEvent{Status: StatusFailed, Error: err}
			return
		}
		if len(files) == 0 {
			return
		}

		// Send total count with first event
		workers := max(1, opts.Workers)
		workers = min(workers, len(files))

		workChan := make(chan string, len(files))
		var wg sync.WaitGroup

		// Start workers
		for i := range workers {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				for path := range workChan {
					select {
					case <-ctx.Done():
						return
					default:
					}
					uploadFile(ctx, g.Api, path, workerID, opts, events)
				}
			}(i)
		}

		// Send work (with total on first)
		first := true
		for _, path := range files {
			select {
			case <-ctx.Done():
				close(workChan)
				wg.Wait()
				return
			default:
			}
			if first {
				events <- UploadEvent{Total: len(files)}
				first = false
			}
			workChan <- path
		}
		close(workChan)
		wg.Wait()
	}()

	return events
}

func uploadFile(ctx context.Context, api *core.Api, filePath string, workerID int, opts UploadOptions, events chan<- UploadEvent) {
	send := func(status UploadStatus, mediaKey, dedupKey string, err error) {
		events <- UploadEvent{
			Path: filePath, Status: status, MediaKey: mediaKey, DedupKey: dedupKey, Error: err, WorkerID: workerID,
		}
	}

	// Hash file
	send(StatusHashing, "", "", nil)
	sha1Hash, err := CalculateSHA1(ctx, filePath)
	if err != nil {
		send(StatusFailed, "", "", fmt.Errorf("hash error: %w", err))
		return
	}
	dedupKey := core.SHA1ToDedupeKey(sha1Hash)

	// Check if exists
	if !opts.ForceUpload {
		send(StatusChecking, "", dedupKey, nil)
		if mediaKey, _ := api.FindMediaKeyByHash(ctx, sha1Hash); mediaKey != "" {
			if opts.DeleteFromHost {
				os.Remove(filePath)
			}
			send(StatusSkipped, mediaKey, dedupKey, nil)
			return
		}
	}

	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		send(StatusFailed, "", dedupKey, fmt.Errorf("stat error: %w", err))
		return
	}

	// Upload
	send(StatusUploading, "", dedupKey, nil)
	sha1Base64 := base64.StdEncoding.EncodeToString([]byte(sha1Hash))
	token, err := api.GetUploadToken(ctx, sha1Base64, fileInfo.Size())
	if err != nil {
		send(StatusFailed, "", dedupKey, fmt.Errorf("upload token error: %w", err))
		return
	}

	commitToken, err := api.UploadFile(ctx, filePath, token)
	if err != nil {
		send(StatusFailed, "", dedupKey, fmt.Errorf("upload error: %w", err))
		return
	}

	// Finalize
	send(StatusFinalizing, "", dedupKey, nil)
	mediaKey, err := api.CommitUpload(ctx, commitToken, fileInfo.Name(), sha1Hash, fileInfo.ModTime().Unix(), opts.Quality, opts.UseQuota)
	if err != nil {
		send(StatusFailed, "", dedupKey, fmt.Errorf("commit error: %w", err))
		return
	}
	if mediaKey == "" {
		send(StatusFailed, "", dedupKey, fmt.Errorf("no media key returned"))
		return
	}

	// Post-upload ops
	if opts.Caption != "" {
		if err := api.SetCaption(ctx, mediaKey, opts.Caption); err != nil {
			slog.Error("caption failed", "path", filePath, "error", err)
		}
	}
	if opts.ShouldFavourite {
		if err := api.SetFavourite(ctx, mediaKey, true); err != nil {
			slog.Error("favourite failed", "path", filePath, "error", err)
		}
	}
	if opts.ShouldArchive {
		if err := api.SetArchived(ctx, []string{mediaKey}, true); err != nil {
			slog.Error("archive failed", "path", filePath, "error", err)
		}
	}
	if opts.DeleteFromHost {
		os.Remove(filePath)
	}

	send(StatusCompleted, mediaKey, dedupKey, nil)
}
