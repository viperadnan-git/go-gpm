package main

import (
	"context"
	"fmt"

	gpm "github.com/viperadnan-git/go-gpm"

	"github.com/urfave/cli/v3"
)

func deleteAction(ctx context.Context, cmd *cli.Command) error {
	if err := loadConfig(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	cfg := cfgManager.GetConfig()

	restore := cmd.Bool("restore")
	forceDelete := cmd.Bool("force")

	// Ensure restore and permanent are mutually exclusive
	if restore && forceDelete {
		return fmt.Errorf("--restore and --force flags are mutually exclusive")
	}

	// Collect inputs from both command-line args and file
	var inputs []string

	// Get all arguments
	allArgs := cmd.Args().Slice()
	if len(allArgs) > 0 {
		inputs = append(inputs, allArgs...)
	}

	// Get items from file if --from-file is provided
	if fromFile := cmd.String("from-file"); fromFile != "" {
		fileInputs, err := readLinesFromFile(fromFile)
		if err != nil {
			return err
		}
		inputs = append(inputs, fileInputs...)
	}

	if len(inputs) == 0 {
		return fmt.Errorf("at least one item is required (provide via command-line or --from-file)")
	}

	authData := getAuthData(cfg)
	if authData == "" {
		return fmt.Errorf("no authentication configured. Use 'gpcli auth add' to add credentials")
	}

	apiClient, err := gpm.NewGooglePhotosAPI(gpm.ApiConfig{
		AuthData: authData,
		Proxy:    cfg.Proxy,
	})
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	logger.Info("resolving items", "count", len(inputs))
	itemKeys := make([]string, 0, len(inputs))
	for _, input := range inputs {
		itemKey, err := apiClient.ResolveItemKey(ctx, input)
		if err != nil {
			return fmt.Errorf("failed to resolve item key for %s: %w", input, err)
		}
		itemKeys = append(itemKeys, itemKey)
	}

	if restore {
		logger.Info("restoring from trash", "count", len(itemKeys))
		if err := apiClient.RestoreFromTrash(itemKeys); err != nil {
			return fmt.Errorf("failed to restore from trash: %w", err)
		}
		logger.Info("successfully restored from trash", "count", len(itemKeys))
	} else if forceDelete {
		logger.Info("permanently deleting", "count", len(itemKeys))
		if err := apiClient.PermanentDelete(itemKeys); err != nil {
			return fmt.Errorf("failed to permanently delete: %w", err)
		}
		logger.Info("successfully permanently deleted", "count", len(itemKeys))
	} else {
		logger.Info("moving to trash", "count", len(itemKeys))
		if err := apiClient.MoveToTrash(itemKeys); err != nil {
			return fmt.Errorf("failed to move to trash: %w", err)
		}
		logger.Info("successfully moved to trash", "count", len(itemKeys))
	}

	return nil
}

func archiveAction(ctx context.Context, cmd *cli.Command) error {
	if err := loadConfig(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	cfg := cfgManager.GetConfig()

	unarchive := cmd.Bool("unarchive")

	// Collect inputs from both command-line args and file
	var inputs []string

	// Get all arguments
	allArgs := cmd.Args().Slice()
	if len(allArgs) > 0 {
		inputs = append(inputs, allArgs...)
	}

	// Get items from file if --from-file is provided
	if fromFile := cmd.String("from-file"); fromFile != "" {
		fileInputs, err := readLinesFromFile(fromFile)
		if err != nil {
			return err
		}
		inputs = append(inputs, fileInputs...)
	}

	if len(inputs) == 0 {
		return fmt.Errorf("at least one item is required (provide via command-line or --from-file)")
	}

	authData := getAuthData(cfg)
	if authData == "" {
		return fmt.Errorf("no authentication configured. Use 'gpcli auth add' to add credentials")
	}

	apiClient, err := gpm.NewGooglePhotosAPI(gpm.ApiConfig{
		AuthData: authData,
		Proxy:    cfg.Proxy,
	})
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	logger.Info("resolving items", "count", len(inputs))
	itemKeys := make([]string, 0, len(inputs))
	for _, input := range inputs {
		itemKey, err := apiClient.ResolveItemKey(ctx, input)
		if err != nil {
			return fmt.Errorf("failed to resolve item key for %s: %w", input, err)
		}
		itemKeys = append(itemKeys, itemKey)
	}

	isArchived := !unarchive
	if isArchived {
		logger.Info("archiving items", "count", len(itemKeys))
	} else {
		logger.Info("unarchiving items", "count", len(itemKeys))
	}

	if err := apiClient.SetArchived(itemKeys, isArchived); err != nil {
		return fmt.Errorf("failed to set archived status: %w", err)
	}

	logger.Info("successfully updated archive status", "count", len(itemKeys))
	return nil
}

func favouriteAction(ctx context.Context, cmd *cli.Command) error {
	if err := loadConfig(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	cfg := cfgManager.GetConfig()

	input := cmd.StringArg("input")
	remove := cmd.Bool("remove")

	authData := getAuthData(cfg)
	if authData == "" {
		return fmt.Errorf("no authentication configured. Use 'gpcli auth add' to add credentials")
	}

	apiClient, err := gpm.NewGooglePhotosAPI(gpm.ApiConfig{
		AuthData: authData,
		Proxy:    cfg.Proxy,
	})
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	itemKey, err := apiClient.ResolveItemKey(ctx, input)
	if err != nil {
		return err
	}

	isFavourite := !remove
	if isFavourite {
		logger.Info("adding to favourites", "item_key", itemKey)
	} else {
		logger.Info("removing from favourites", "item_key", itemKey)
	}

	if err := apiClient.SetFavourite(itemKey, isFavourite); err != nil {
		return fmt.Errorf("failed to set favourite status: %w", err)
	}

	return nil
}

func captionAction(ctx context.Context, cmd *cli.Command) error {
	if err := loadConfig(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	cfg := cfgManager.GetConfig()

	input := cmd.StringArg("input")
	caption := cmd.StringArg("caption")

	authData := getAuthData(cfg)
	if authData == "" {
		return fmt.Errorf("no authentication configured. Use 'gpcli auth add' to add credentials")
	}

	apiClient, err := gpm.NewGooglePhotosAPI(gpm.ApiConfig{
		AuthData: authData,
		Proxy:    cfg.Proxy,
	})
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	itemKey, err := apiClient.ResolveItemKey(ctx, input)
	if err != nil {
		return err
	}

	logger.Info("setting caption", "item_key", itemKey, "caption", caption)

	if err := apiClient.SetCaption(itemKey, caption); err != nil {
		return fmt.Errorf("failed to set caption: %w", err)
	}

	return nil
}
