package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli/v3"
	"github.com/viperadnan-git/go-gpm/internal/core"
)

func libraryAction(ctx context.Context, cmd *cli.Command) error {
	if err := loadConfig(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	apiClient, err := createAPIClient()
	if err != nil {
		return err
	}

	showTrashed := cmd.Bool("trashed")
	albumFilter := cmd.String("album")
	rawOutput := cmd.Bool("raw")

	// Handle raw JSON output
	if rawOutput {
		jsonData, err := apiClient.FetchLibraryStateRaw(ctx, "")
		if err != nil {
			return fmt.Errorf("failed to fetch library state: %w", err)
		}
		fmt.Println(string(jsonData))
		return nil
	}

	// Fetch initial library state
	fmt.Println("Fetching library state...")
	resp, err := apiClient.FetchLibraryState(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to fetch library state: %w", err)
	}

	// Build album map for lookup (by key and by name)
	albumMap := make(map[string]*core.AlbumInfo)
	albumByName := make(map[string]*core.AlbumInfo)
	for i := range resp.Albums {
		albumMap[resp.Albums[i].AlbumKey] = &resp.Albums[i]
		albumByName[strings.ToLower(resp.Albums[i].Name)] = &resp.Albums[i]
	}

	// Resolve album filter to album key
	var filterAlbumKey string
	if albumFilter != "" {
		// Try exact key match first
		if _, ok := albumMap[albumFilter]; ok {
			filterAlbumKey = albumFilter
		} else if album, ok := albumByName[strings.ToLower(albumFilter)]; ok {
			// Try name match (case-insensitive)
			filterAlbumKey = album.AlbumKey
		} else {
			return fmt.Errorf("album not found: %s", albumFilter)
		}
		fmt.Printf("Filtering by album: %s\n", albumMap[filterAlbumKey].Name)
	}

	pageNum := 1
	stateToken := resp.StateToken

	for {
		// Print albums (only on first page and if no album filter)
		if pageNum == 1 && len(resp.Albums) > 0 && albumFilter == "" {
			printAlbums(resp.Albums)
		}

		// Print media items
		printMediaItems(resp.MediaItems, albumMap, showTrashed, filterAlbumKey, pageNum)

		// Check for next page
		if resp.PageToken == "" {
			fmt.Println("\n--- End of library ---")
			break
		}

		// Ask user for next page
		fmt.Printf("\nMore items available. Load next page? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		if input != "y" && input != "yes" {
			fmt.Println("Exiting.")
			break
		}

		// Fetch next page
		pageNum++
		fmt.Printf("\nFetching page %d...\n", pageNum)
		resp, err = apiClient.FetchLibraryPage(ctx, resp.PageToken, stateToken)
		if err != nil {
			return fmt.Errorf("failed to fetch library page: %w", err)
		}
	}

	return nil
}

func printAlbums(albums []core.AlbumInfo) {
	fmt.Println("\n=== ALBUMS ===")
	fmt.Printf("%-45s  %-25s  %s\n", "ALBUM KEY", "NAME", "ITEMS")
	fmt.Println(strings.Repeat("-", 80))

	for _, album := range albums {
		name := album.Name
		if len(name) > 25 {
			name = name[:22] + "..."
		}
		fmt.Printf("%-45s  %-25s  %d\n", album.AlbumKey, name, album.ItemCount)
	}
	fmt.Println()
}

func printMediaItems(items []core.MediaItemInfo, albumMap map[string]*core.AlbumInfo, showTrashed bool, filterAlbumKey string, pageNum int) {
	fmt.Printf("\n=== MEDIA ITEMS (Page %d) ===\n", pageNum)
	fmt.Printf("%-45s  %-30s  %-10s  %-12s  %s\n", "MEDIA KEY", "FILENAME", "TYPE", "SIZE", "STATUS")
	fmt.Println(strings.Repeat("-", 120))

	displayed := 0
	skippedTrash := 0
	skippedAlbum := 0

	for _, item := range items {
		// Skip trashed items unless --trashed flag is set
		if item.IsInTrash && !showTrashed {
			skippedTrash++
			continue
		}

		// Skip items not in the filtered album
		if filterAlbumKey != "" && item.AlbumMediaKey != filterAlbumKey {
			skippedAlbum++
			continue
		}

		filename := item.Filename
		if len(filename) > 30 {
			filename = filename[:27] + "..."
		}

		mediaType := "image"
		if item.IsVideo {
			mediaType = "video"
		}

		size := formatSize(item.FileSize)

		status := ""
		if item.IsInTrash {
			status = "[TRASH]"
			if item.TrashedAt > 0 {
				trashedTime := time.UnixMilli(item.TrashedAt)
				status = fmt.Sprintf("[TRASH %s]", trashedTime.Format("2006-01-02"))
			}
		} else if filterAlbumKey == "" {
			// Only show album status if not filtering by album
			if album, ok := albumMap[item.AlbumMediaKey]; ok {
				status = fmt.Sprintf("[in: %s]", truncateName(album.Name, 15))
			}
		}

		fmt.Printf("%-45s  %-30s  %-10s  %-12s  %s\n",
			item.MediaKey, filename, mediaType, size, status)
		displayed++
	}

	fmt.Printf("\nShowing %d items", displayed)
	if skippedTrash > 0 {
		fmt.Printf(" (hiding %d trashed)", skippedTrash)
	}
	fmt.Println()
}

func formatSize(bytes int64) string {
	if bytes == 0 {
		return "-"
	}
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func truncateName(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
