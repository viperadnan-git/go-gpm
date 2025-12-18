package main

import (
	"context"
	"fmt"
	"regexp"

	"github.com/urfave/cli/v3"
)

// albumKeyPattern matches album keys like AF1QipOTAHAvdvLHVyvBNXPZy_93ArwuxfW9dATmqi8T
var albumKeyPattern = regexp.MustCompile(`^AF1Qip[A-Za-z0-9_-]{32,44}$`)

// isAlbumKey returns true if the input matches the album key pattern
func isAlbumKey(s string) bool {
	return albumKeyPattern.MatchString(s)
}

// resolveAlbumKey resolves an album name or key to an album key
func resolveAlbumKey(input string) (string, error) {
	if isAlbumKey(input) {
		return input, nil
	}
	if key := cfgManager.GetAlbumKey(input); key != "" {
		return key, nil
	}
	return "", fmt.Errorf("album '%s' not found in stored mappings (use album key or store the mapping first)", input)
}

func albumCreateAction(ctx context.Context, cmd *cli.Command) error {
	if err := loadConfig(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	albumName := cmd.StringArg("name")
	if albumName == "" {
		return fmt.Errorf("album name is required")
	}

	apiClient, err := createAPIClient()
	if err != nil {
		return err
	}

	// Get media items from command-line args
	// Note: cmd.Args().Slice() returns unconsumed args (album name is already consumed by StringArg)
	mediaInputs := cmd.Args().Slice()

	// Resolve all media keys (if any provided)
	var mediaKeys []string
	if len(mediaInputs) > 0 {
		logger.Info("resolving media items", "count", len(mediaInputs))
		mediaKeys = make([]string, 0, len(mediaInputs))
		for _, input := range mediaInputs {
			mediaKey, err := apiClient.ResolveMediaKey(ctx, input)
			if err != nil {
				return fmt.Errorf("failed to resolve media key for %s: %w", input, err)
			}
			mediaKeys = append(mediaKeys, mediaKey)
		}
	}

	logger.Info("creating album", "name", albumName, "media_count", len(mediaKeys))

	albumKey, err := apiClient.CreateAlbum(ctx, albumName, mediaKeys)
	if err != nil {
		return fmt.Errorf("failed to create album: %w", err)
	}

	// Auto-store the album mapping
	if err := cfgManager.SetAlbumMapping(albumName, albumKey); err != nil {
		logger.Warn("failed to store album mapping", "error", err)
	}

	logger.Info("album created successfully", "name", albumName, "album_key", albumKey, "media_count", len(mediaKeys))
	return nil
}

func albumAddAction(ctx context.Context, cmd *cli.Command) error {
	if err := loadConfig(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	albumInput := cmd.StringArg("album-key")
	if albumInput == "" {
		return fmt.Errorf("album key or name is required")
	}

	albumKey, err := resolveAlbumKey(albumInput)
	if err != nil {
		return err
	}

	// Collect media inputs from both command-line args and file
	var mediaInputs []string

	// Get all unconsumed arguments (album-key is already consumed by StringArg)
	allArgs := cmd.Args().Slice()
	if len(allArgs) > 0 {
		mediaInputs = append(mediaInputs, allArgs...)
	}

	// Get media items from file if --from-file is provided
	if fromFile := cmd.String("from-file"); fromFile != "" {
		fileInputs, err := readLinesFromFile(fromFile)
		if err != nil {
			return err
		}
		mediaInputs = append(mediaInputs, fileInputs...)
	}

	if len(mediaInputs) == 0 {
		return fmt.Errorf("at least one media item is required (provide via command-line or --from-file)")
	}

	apiClient, err := createAPIClient()
	if err != nil {
		return err
	}

	logger.Info("resolving media items", "count", len(mediaInputs))
	mediaKeys := make([]string, 0, len(mediaInputs))
	for _, input := range mediaInputs {
		mediaKey, err := apiClient.ResolveMediaKey(ctx, input)
		if err != nil {
			return fmt.Errorf("failed to resolve media key for %s: %w", input, err)
		}
		mediaKeys = append(mediaKeys, mediaKey)
	}

	logger.Info("adding media to album", "album_key", albumKey, "media_count", len(mediaKeys))

	if err := apiClient.AddMediaToAlbum(ctx, albumKey, mediaKeys); err != nil {
		return fmt.Errorf("failed to add media to album: %w", err)
	}

	logger.Info("successfully added media to album", "album_key", albumKey, "media_count", len(mediaKeys))
	return nil
}

func albumDeleteAction(ctx context.Context, cmd *cli.Command) error {
	if err := loadConfig(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	albumInput := cmd.StringArg("album-key")
	if albumInput == "" {
		return fmt.Errorf("album key or name is required")
	}

	albumKey, err := resolveAlbumKey(albumInput)
	if err != nil {
		return err
	}

	apiClient, err := createAPIClient()
	if err != nil {
		return err
	}

	logger.Info("deleting album", "album_key", albumKey)

	if err := apiClient.DeleteAlbum(ctx, albumKey); err != nil {
		return fmt.Errorf("failed to delete album: %w", err)
	}

	logger.Info("album deleted successfully", "album_key", albumKey)
	return nil
}

func albumRenameAction(ctx context.Context, cmd *cli.Command) error {
	if err := loadConfig(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	albumInput := cmd.StringArg("album-key")
	if albumInput == "" {
		return fmt.Errorf("album key or name is required")
	}

	albumKey, err := resolveAlbumKey(albumInput)
	if err != nil {
		return err
	}

	newName := cmd.StringArg("new-name")
	if newName == "" {
		return fmt.Errorf("new album name is required")
	}

	apiClient, err := createAPIClient()
	if err != nil {
		return err
	}

	logger.Info("renaming album", "album_key", albumKey, "new_name", newName)

	if err := apiClient.RenameAlbum(ctx, albumKey, newName); err != nil {
		return fmt.Errorf("failed to rename album: %w", err)
	}

	logger.Info("album renamed successfully", "album_key", albumKey, "new_name", newName)
	return nil
}

func albumStoreAddAction(ctx context.Context, cmd *cli.Command) error {
	if err := loadConfig(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	key := cmd.StringArg("key")
	if key == "" {
		return fmt.Errorf("album key is required")
	}

	name := cmd.String("name")
	if name == "" {
		return fmt.Errorf("album name is required (use --name)")
	}

	if err := cfgManager.SetAlbumMapping(name, key); err != nil {
		return fmt.Errorf("failed to store album mapping: %w", err)
	}

	logger.Info("album mapping stored", "name", name, "key", key)
	return nil
}

func albumStoreRemoveAction(ctx context.Context, cmd *cli.Command) error {
	if err := loadConfig(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	name := cmd.StringArg("name")
	if name == "" {
		return fmt.Errorf("album name is required")
	}

	if err := cfgManager.RemoveAlbumMapping(name); err != nil {
		return fmt.Errorf("failed to remove album mapping: %w", err)
	}

	logger.Info("album mapping removed", "name", name)
	return nil
}

func albumStoreListAction(ctx context.Context, cmd *cli.Command) error {
	if err := loadConfig(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	mappings := cfgManager.GetAlbumMappings()
	if len(mappings) == 0 {
		logger.Info("no album mappings stored")
		return nil
	}

	for name, key := range mappings {
		fmt.Printf("%s -> %s\n", name, key)
	}
	return nil
}
