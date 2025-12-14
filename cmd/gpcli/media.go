package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/urfave/cli/v3"
)

func deleteAction(ctx context.Context, cmd *cli.Command) error {
	if err := loadConfig(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

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

	apiClient, err := createAPIClient()
	if err != nil {
		return err
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
		if err := apiClient.RestoreFromTrash(ctx, itemKeys); err != nil {
			return fmt.Errorf("failed to restore from trash: %w", err)
		}
		logger.Info("successfully restored from trash", "count", len(itemKeys))
	} else if forceDelete {
		logger.Info("permanently deleting", "count", len(itemKeys))
		if err := apiClient.PermanentDelete(ctx, itemKeys); err != nil {
			return fmt.Errorf("failed to permanently delete: %w", err)
		}
		logger.Info("successfully permanently deleted", "count", len(itemKeys))
	} else {
		logger.Info("moving to trash", "count", len(itemKeys))
		if err := apiClient.MoveToTrash(ctx, itemKeys); err != nil {
			return fmt.Errorf("failed to move to trash: %w", err)
		}
		logger.Info("successfully moved to trash", "count", len(itemKeys))
	}

	logger.Debug("successfully deleted", "count", len(itemKeys))
	return nil
}

func archiveAction(ctx context.Context, cmd *cli.Command) error {
	if err := loadConfig(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

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

	apiClient, err := createAPIClient()
	if err != nil {
		return err
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

	if err := apiClient.SetArchived(ctx, itemKeys, isArchived); err != nil {
		return fmt.Errorf("failed to set archived status: %w", err)
	}

	logger.Debug("archive status updated successfully", "count", len(itemKeys))
	return nil
}

func favouriteAction(ctx context.Context, cmd *cli.Command) error {
	if err := loadConfig(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	input := cmd.StringArg("input")
	remove := cmd.Bool("remove")

	apiClient, err := createAPIClient()
	if err != nil {
		return err
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

	if err := apiClient.SetFavourite(ctx, itemKey, isFavourite); err != nil {
		return fmt.Errorf("failed to set favourite status: %w", err)
	}

	logger.Debug("favourite status updated successfully")
	return nil
}

func captionAction(ctx context.Context, cmd *cli.Command) error {
	if err := loadConfig(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	input := cmd.StringArg("input")
	caption := cmd.StringArg("caption")

	apiClient, err := createAPIClient()
	if err != nil {
		return err
	}

	itemKey, err := apiClient.ResolveItemKey(ctx, input)
	if err != nil {
		return err
	}

	logger.Info("setting caption", "item_key", itemKey, "caption", caption)

	if err := apiClient.SetCaption(ctx, itemKey, caption); err != nil {
		return fmt.Errorf("failed to set caption: %w", err)
	}

	logger.Debug("caption updated successfully")
	return nil
}

func resolveAction(ctx context.Context, cmd *cli.Command) error {
	if err := loadConfig(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	input := cmd.StringArg("input")

	apiClient, err := createAPIClient()
	if err != nil {
		return err
	}

	mediaKey, err := apiClient.ResolveMediaKey(ctx, input)
	if err != nil {
		return err
	}

	fmt.Println(mediaKey)
	return nil
}

func locationAction(ctx context.Context, cmd *cli.Command) error {
	if err := loadConfig(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	input := cmd.StringArg("input")
	latitude := float32(cmd.Float("latitude"))
	longitude := float32(cmd.Float("longitude"))

	// Validate coordinates
	if latitude < -90 || latitude > 90 {
		return fmt.Errorf("latitude must be between -90 and 90")
	}
	if longitude < -180 || longitude > 180 {
		return fmt.Errorf("longitude must be between -180 and 180")
	}

	apiClient, err := createAPIClient()
	if err != nil {
		return err
	}

	itemKey, err := apiClient.ResolveItemKey(ctx, input)
	if err != nil {
		return err
	}

	logger.Info("setting location",
		"item_key", itemKey,
		"latitude", latitude,
		"longitude", longitude)

	if err := apiClient.SetLocation(ctx, itemKey, latitude, longitude); err != nil {
		return fmt.Errorf("failed to set location: %w", err)
	}

	logger.Debug("location updated successfully")
	return nil
}

func datetimeAction(ctx context.Context, cmd *cli.Command) error {
	if err := loadConfig(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	datetimeStr := cmd.StringArg("datetime")

	// Parse datetime
	var timestamp time.Time
	var err error

	if strings.ToLower(datetimeStr) == "now" {
		timestamp = time.Now()
	} else {
		// Try ISO 8601 format
		timestamp, err = time.Parse(time.RFC3339, datetimeStr)
		if err != nil {
			return fmt.Errorf("invalid datetime format. Use ISO 8601 format (e.g., '2024-12-24T15:30:00+05:30') or 'now': %w", err)
		}
	}

	// Collect inputs from both command-line args and file
	var inputs []string

	// Get all remaining arguments after datetime
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

	apiClient, err := createAPIClient()
	if err != nil {
		return err
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

	logger.Info("setting datetime", "count", len(itemKeys), "datetime", timestamp.Format(time.RFC3339))

	if err := apiClient.SetDateTime(ctx, itemKeys, timestamp); err != nil {
		return fmt.Errorf("failed to set datetime: %w", err)
	}

	logger.Debug("datetime updated successfully", "count", len(itemKeys))
	return nil
}
