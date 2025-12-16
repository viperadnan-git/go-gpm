package gpm

import (
	"context"
	"sync"

	"github.com/viperadnan-git/go-gpm/internal/core"
)

// ApiConfig holds configuration for the Google Photos API client
type ApiConfig = core.ApiConfig

// TokenCache defines the interface for token storage
type TokenCache = core.TokenCache

// MemoryTokenCache stores tokens in memory (thread-safe)
type MemoryTokenCache = core.MemoryTokenCache

// DownloadInfo contains download information for a media item
type DownloadInfo = core.DownloadInfo

// NewMemoryTokenCache creates a new in-memory token cache
func NewMemoryTokenCache() *MemoryTokenCache {
	return core.NewMemoryTokenCache()
}

// GooglePhotosAPI is the main API client for Google Photos operations
type GooglePhotosAPI struct {
	*core.Api
	uploadMu sync.Mutex // Serializes upload batches
}

// NewGooglePhotosAPI creates a new Google Photos API client
func NewGooglePhotosAPI(cfg ApiConfig) (*GooglePhotosAPI, error) {
	coreApi, err := core.NewApi(cfg)
	if err != nil {
		return nil, err
	}
	return &GooglePhotosAPI{Api: coreApi}, nil
}

// DownloadThumbnail downloads a thumbnail to the specified output path
// Returns the final output path
func (g *GooglePhotosAPI) DownloadThumbnail(ctx context.Context, mediaKey string, width, height int, forceJpeg, noOverlay bool, outputPath string) (string, error) {
	body, err := g.GetThumbnail(ctx, mediaKey, width, height, forceJpeg, noOverlay)
	if err != nil {
		return "", err
	}
	defer body.Close()

	filename := mediaKey + ".jpg"
	return DownloadFromReader(body, outputPath, filename)
}

// DownloadMedia downloads a media item to the specified output path
// Returns the final output path
func (g *GooglePhotosAPI) DownloadMedia(ctx context.Context, mediaKey string, outputPath string) (string, error) {
	info, err := g.GetDownloadInfo(ctx, mediaKey)
	if err != nil {
		return "", err
	}
	return DownloadFile(info.DownloadURL, outputPath, info.Filename)
}
