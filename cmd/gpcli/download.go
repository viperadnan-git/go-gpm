package main

import (
	"context"
	"fmt"

	gpm "github.com/viperadnan-git/go-gpm"

	"github.com/urfave/cli/v3"
)

func downloadAction(ctx context.Context, cmd *cli.Command) error {
	if err := loadConfig(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	input := cmd.StringArg("input")
	urlOnly := cmd.Bool("url")
	outputPath := cmd.String("output")

	apiClient, err := createAPIClient()
	if err != nil {
		return err
	}

	mediaKey, err := apiClient.ResolveMediaKey(ctx, input)
	if err != nil {
		return err
	}

	if !urlOnly {
		logger.Info("fetching download URL", "media_key", mediaKey)
	}

	downloadURL, isEdited, err := apiClient.GetDownloadUrl(ctx, mediaKey)
	if err != nil {
		return fmt.Errorf("failed to get download URL: %w", err)
	}

	if downloadURL == "" {
		return fmt.Errorf("no download URL available")
	}

	// If --url flag is set, just print the URL and exit
	if urlOnly {
		fmt.Println(downloadURL)
		return nil
	}

	// Download the file
	logger.Info("downloading", "is_edited", isEdited)
	savedPath, err := gpm.DownloadFile(downloadURL, outputPath)
	if err != nil {
		return err
	}
	logger.Info("download complete", "path", savedPath)
	return nil
}

func thumbnailAction(ctx context.Context, cmd *cli.Command) error {
	if err := loadConfig(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	input := cmd.StringArg("input")
	outputPath := cmd.String("output")
	width := int(cmd.Int("width"))
	height := int(cmd.Int("height"))
	forceJpeg := cmd.Bool("jpeg")
	noOverlay := !cmd.Bool("overlay")

	apiClient, err := createAPIClient()
	if err != nil {
		return err
	}

	mediaKey, err := apiClient.ResolveMediaKey(ctx, input)
	if err != nil {
		return err
	}

	logger.Info("downloading thumbnail", "media_key", mediaKey)

	savedPath, err := apiClient.DownloadThumbnail(ctx, mediaKey, width, height, forceJpeg, noOverlay, outputPath)
	if err != nil {
		return err
	}
	logger.Info("thumbnail downloaded", "path", savedPath)
	return nil
}
