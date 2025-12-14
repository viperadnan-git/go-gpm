package main

import (
	"context"
	"fmt"
	"path"
	"runtime"

	"github.com/creativeprojects/go-selfupdate"
	"github.com/urfave/cli/v3"
	gpm "github.com/viperadnan-git/go-gpm"
)

const repoSlug = "viperadnan-git/go-gpm"

func upgradeAction(ctx context.Context, cmd *cli.Command) error {
	// Check if --url flag is provided
	if assetURL := cmd.String("url"); assetURL != "" {
		return upgradeFromURL(ctx, assetURL)
	}

	// Check if --nightly flag is provided
	if cmd.Bool("nightly") {
		return upgradeFromNightly(ctx)
	}

	// Get target version (empty string = latest)
	targetVersion := cmd.StringArg("version")
	checkOnly := cmd.Bool("check")

	// Configure updater for GitHub
	source, err := selfupdate.NewGitHubSource(selfupdate.GitHubConfig{})
	if err != nil {
		return fmt.Errorf("failed to create GitHub source: %w", err)
	}

	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Source:    source,
		Validator: &selfupdate.ChecksumValidator{UniqueFilename: "checksums.txt"},
	})
	if err != nil {
		return fmt.Errorf("failed to create updater: %w", err)
	}

	repo := selfupdate.ParseSlug(repoSlug)

	var release *selfupdate.Release
	var found bool

	if targetVersion != "" {
		// Find specific version
		logger.Info("checking for version", "version", targetVersion)
		release, found, err = updater.DetectVersion(ctx, repo, targetVersion)
	} else {
		// Find latest version
		logger.Info("checking for latest version")
		release, found, err = updater.DetectLatest(ctx, repo)
	}

	if err != nil {
		return fmt.Errorf("failed to detect version: %w", err)
	}
	if !found {
		if targetVersion != "" {
			return fmt.Errorf("version %s not found", targetVersion)
		}
		logger.Info("no release found")
		return nil
	}

	currentVersion := gpm.Version

	// Compare versions (skip if same and no specific version requested)
	if targetVersion == "" && release.Version() == currentVersion {
		logger.Info("already at latest version", "version", currentVersion)
		return nil
	}

	// Check-only mode: display info and exit
	if checkOnly {
		logger.Info("update available", "current", currentVersion, "available", release.Version())
		return nil
	}

	logger.Info("updating", "from", currentVersion, "to", release.Version())

	// Get executable path
	exe, err := selfupdate.ExecutablePath()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Perform update
	if err := updater.UpdateTo(ctx, release, exe); err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}

	logger.Info("successfully updated", "version", release.Version())
	return nil
}

// upgradeFromURL downloads and installs a binary from a direct URL
func upgradeFromURL(ctx context.Context, assetURL string) error {
	logger.Info("downloading from URL", "url", assetURL)

	exe, err := selfupdate.ExecutablePath()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	if err := selfupdate.UpdateTo(ctx, assetURL, path.Base(assetURL), exe); err != nil {
		return fmt.Errorf("failed to update from URL: %w", err)
	}

	logger.Info("successfully updated from URL")
	return nil
}

// upgradeFromNightly downloads and installs the latest nightly build from a branch
func upgradeFromNightly(ctx context.Context) error {
	nightlyBranch := "main"
	assetName := fmt.Sprintf("gpcli-%s-%s.zip", runtime.GOOS, runtime.GOARCH)
	assetURL := fmt.Sprintf("https://nightly.link/%s/workflows/build/%s/%s", repoSlug, nightlyBranch, assetName)

	logger.Info("downloading nightly build", "branch", nightlyBranch, "os", runtime.GOOS, "arch", runtime.GOARCH)

	exe, err := selfupdate.ExecutablePath()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	if err := selfupdate.UpdateTo(ctx, assetURL, assetName, exe); err != nil {
		return fmt.Errorf("failed to update from nightly: %w", err)
	}

	logger.Info("successfully updated to nightly build", "branch", nightlyBranch)
	return nil
}
