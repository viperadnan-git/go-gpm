package main

import (
	"context"
	"log/slog"
	"os"

	gpm "github.com/viperadnan-git/go-gpm"

	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:                   "gpcli",
		Usage:                  "Google Photos unofficial CLI client",
		Version:                gpm.Version,
		UseShortOptionHandling: true,
		EnableShellCompletion:  true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				Aliases:     []string{"c"},
				Usage:       "Path to config file",
				Sources:     cli.EnvVars("GPCLI_CONFIG"),
				DefaultText: "~/.config/gpcli/gpcli.toml",
				Config:      cli.StringConfig{TrimSpace: true},
			},
			&cli.StringFlag{
				Name:    "log-level",
				Value:   "info",
				Usage:   "Set log level: debug, info, warn, error",
				Sources: cli.EnvVars("GPCLI_LOG_LEVEL"),
			},
			&cli.BoolFlag{
				Name:    "quiet",
				Aliases: []string{"q"},
				Usage:   "Suppress all log output (overrides --log-level)",
				Sources: cli.EnvVars("GPCLI_QUIET"),
			},
			&cli.StringFlag{
				Name:    "auth",
				Usage:   "Authentication string (overrides config file)",
				Sources: cli.EnvVars("GPCLI_AUTH"),
				Config:  cli.StringConfig{TrimSpace: true},
			},
			&cli.StringFlag{
				Name:    "log-format",
				Value:   "human",
				Usage:   "Log format: human, slog, or json",
				Sources: cli.EnvVars("GPCLI_LOG_FORMAT"),
			},
		},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			// Set log format before initializing logger
			logFormat = cmd.String("log-format")

			// Initialize logger - quiet mode overrides log level
			if cmd.Bool("quiet") {
				initQuietLogger()
			} else {
				currentLogLevel = parseLogLevel(cmd.String("log-level"))
				initLogger(currentLogLevel)
			}

			// Set config path from flag
			configPath = cmd.String("config")

			// Set auth override from flag
			if auth := cmd.String("auth"); auth != "" {
				authOverride = auth
			}
			return ctx, nil
		},
		Commands: []*cli.Command{
			{
				Name:   "auth",
				Usage:  "Manage Google Photos authentication",
				Action: authInfoAction,
				Commands: []*cli.Command{
					{
						Name:  "add",
						Usage: "Add a new authentication",
						Arguments: []cli.Argument{
							&cli.StringArg{
								Name:      "auth-string",
								UsageText: "<auth-string>",
							},
						},
						Action: credentialsAddAction,
					},
					{
						Name:    "list",
						Aliases: []string{"ls"},
						Usage:   "List all authentications",
						Action:  authInfoAction,
					},
					{
						Name:    "remove",
						Aliases: []string{"rm"},
						Usage:   "Remove an authentication by number or email",
						Arguments: []cli.Argument{
							&cli.StringArg{
								Name:      "identifier",
								UsageText: "<number|email>",
							},
						},
						Action: credentialsRemoveAction,
					},
					{
						Name:  "set",
						Usage: "Set active authentication by number or email",
						Arguments: []cli.Argument{
							&cli.StringArg{
								Name:      "identifier",
								UsageText: "<number|email>",
							},
						},
						Action: credentialsSetAction,
					},
					{
						Name:   "file",
						Usage:  "Print config file path",
						Action: authFileAction,
					},
				},
			},
			{
				Name:  "upload",
				Usage: "Upload a file or directory to Google Photos",
				Arguments: []cli.Argument{
					&cli.StringArg{
						Name:      "filepath",
						UsageText: "<filepath>",
					},
				},
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "recursive",
						Aliases: []string{"r"},
						Usage:   "Include subdirectories",
					},
					&cli.IntFlag{
						Name:    "threads",
						Aliases: []string{"t"},
						Value:   3,
						Usage:   "Number of upload threads",
					},
					&cli.BoolFlag{
						Name:    "force",
						Aliases: []string{"f"},
						Usage:   "Force upload even if file exists",
					},
					&cli.BoolFlag{
						Name:    "delete",
						Aliases: []string{"d"},
						Usage:   "Delete from host after upload",
					},
					&cli.BoolFlag{
						Name:  "disable-filter",
						Usage: "Disable file type filtering",
					},
					&cli.StringFlag{
						Name:   "album",
						Usage:  "Add uploaded files to album with this name (creates if not exists)",
						Config: cli.StringConfig{TrimSpace: true},
					},
					&cli.StringFlag{
						Name:    "quality",
						Aliases: []string{"q"},
						Value:   "original",
						Usage:   "Upload quality: 'original' or 'storage-saver'",
					},
					&cli.BoolFlag{
						Name:  "use-quota",
						Usage: "Uploaded files will count against your Google Photos storage quota",
					},
					&cli.BoolFlag{
						Name:    "archive",
						Aliases: []string{"a"},
						Usage:   "Archive uploaded files after upload",
					},
					&cli.StringFlag{
						Name:   "caption",
						Usage:  "Set caption for uploaded files",
						Config: cli.StringConfig{TrimSpace: true},
					},
					&cli.BoolFlag{
						Name:  "favourite",
						Usage: "Mark uploaded files as favourites",
					},
					&cli.StringFlag{
						Name:   "datetime",
						Usage:  "Override datetime for uploaded files (ISO 8601 format or 'now')",
						Config: cli.StringConfig{TrimSpace: true},
					},
					&cli.BoolFlag{
						Name:    "check",
						Aliases: []string{"c"},
						Usage:   "Dry run: check which files would be uploaded vs already exist",
					},
				},
				Action: uploadAction,
			},
			{
				Name:  "download",
				Usage: "Download a media item",
				Arguments: []cli.Argument{
					&cli.StringArg{
						Name:      "input",
						UsageText: "<item-key|filepath>",
					},
				},
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "url",
						Usage: "Only print download URL without downloading",
					},
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Usage:   "Output path (file path or directory)",
						Config:  cli.StringConfig{TrimSpace: true},
					},
				},
				Action: downloadAction,
			},
			{
				Name:  "thumbnail",
				Usage: "Download thumbnail for a media item",
				Arguments: []cli.Argument{
					&cli.StringArg{
						Name:      "input",
						UsageText: "<item-key|filepath>",
					},
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Usage:   "Output path (file path or directory)",
						Config:  cli.StringConfig{TrimSpace: true},
					},
					&cli.IntFlag{
						Name:    "width",
						Aliases: []string{"w"},
						Usage:   "Thumbnail width in pixels",
					},
					&cli.IntFlag{
						Name:  "height",
						Usage: "Thumbnail height in pixels",
					},
					&cli.BoolFlag{
						Name:  "jpeg",
						Usage: "Force JPEG format output",
					},
					&cli.BoolFlag{
						Name:  "overlay",
						Usage: "Show video overlay icon (hidden by default)",
					},
				},
				Action: thumbnailAction,
			},
			{
				Name:      "delete",
				Usage:     "Move items to trash, restore from trash, or permanently delete",
				UsageText: "gpcli delete <input> [input...] [--from-file FILE]",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "restore",
						Usage:   "Restore from trash instead of delete",
						Aliases: []string{"r"},
					},
					&cli.BoolFlag{
						Name:    "force",
						Usage:   "Permanently delete (can't be undone)",
						Aliases: []string{"f"},
					},
					&cli.StringFlag{
						Name:    "from-file",
						Aliases: []string{"i"},
						Usage:   "Read item keys from file (one per line)",
						Config:  cli.StringConfig{TrimSpace: true},
					},
				},
				Action: deleteAction,
			},
			{
				Name:      "archive",
				Usage:     "Archive or unarchive items",
				UsageText: "gpcli archive <input> [input...] [--from-file FILE]",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "unarchive",
						Aliases: []string{"r", "u"},
						Usage:   "Unarchive instead of archive",
					},
					&cli.StringFlag{
						Name:    "from-file",
						Aliases: []string{"i"},
						Usage:   "Read item keys from file (one per line)",
						Config:  cli.StringConfig{TrimSpace: true},
					},
				},
				Action: archiveAction,
			},
			{
				Name:  "favourite",
				Usage: "Add or remove favourite status",
				Arguments: []cli.Argument{
					&cli.StringArg{
						Name:      "input",
						UsageText: "<item-key|filepath>",
					},
				},
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "remove",
						Aliases: []string{"r"},
						Usage:   "Remove favourite status",
					},
				},
				Action: favouriteAction,
			},
			{
				Name:  "caption",
				Usage: "Set item caption",
				Arguments: []cli.Argument{
					&cli.StringArg{
						Name:      "input",
						UsageText: "<item-key|filepath>",
					},
					&cli.StringArg{
						Name:      "caption",
						UsageText: "<caption>",
						Config:    cli.StringConfig{TrimSpace: true},
					},
				},
				Action: captionAction,
			},
			{
				Name:  "resolve",
				Usage: "Resolve hash, dedup key, or file path to media key",
				Arguments: []cli.Argument{
					&cli.StringArg{
						Name:      "input",
						UsageText: "<hash|dedup-key|media-key|filepath>",
					},
				},
				Action: resolveAction,
			},
			{
				Name:        "location",
				Usage:       "Set geographic location for a media item",
				Description: "Sets the location coordinates for a media item. Note: The place name will be generic and not specific to the coordinates.",
				Arguments: []cli.Argument{
					&cli.StringArg{
						Name:      "input",
						UsageText: "<item-key|filepath>",
					},
				},
				Flags: []cli.Flag{
					&cli.Float64Flag{
						Name:     "latitude",
						Aliases:  []string{"lat"},
						Usage:    "Latitude in decimal degrees (required)",
						Required: true,
					},
					&cli.Float64Flag{
						Name:     "longitude",
						Aliases:  []string{"lon", "lng"},
						Usage:    "Longitude in decimal degrees (required)",
						Required: true,
					},
				},
				Action: locationAction,
			},
			{
				Name:      "datetime",
				Usage:     "Set date and time for one or more media items",
				UsageText: "gpcli datetime <datetime> <input> [input...] [--from-file FILE]",
				Arguments: []cli.Argument{
					&cli.StringArg{
						Name:      "datetime",
						UsageText: "Date and time in ISO 8601 format (e.g., '2024-12-24T15:30:00+05:30') or 'now'",
						Config:    cli.StringConfig{TrimSpace: true},
					},
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "from-file",
						Aliases: []string{"i"},
						Usage:   "Read item keys from file (one per line)",
						Config:  cli.StringConfig{TrimSpace: true},
					},
				},
				Action: datetimeAction,
			},
			{
				Name:  "album",
				Usage: "Manage albums",
				Commands: []*cli.Command{
					{
						Name:      "create",
						Usage:     "Create a new album with media items",
						UsageText: "gpcli album create <name> <media-key> [media-key...]",
						Arguments: []cli.Argument{
							&cli.StringArg{
								Name:      "name",
								UsageText: "Album name",
								Config:    cli.StringConfig{TrimSpace: true},
							},
						},
						Action: albumCreateAction,
					},
					{
						Name:      "add",
						Usage:     "Add media items to an existing album",
						UsageText: "gpcli album add <album-key> <media-key> [media-key...] [--from-file FILE]",
						Arguments: []cli.Argument{
							&cli.StringArg{
								Name:      "album-key",
								UsageText: "Album media key",
							},
						},
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "from-file",
								Aliases: []string{"i"},
								Usage:   "Read media keys from file (one per line)",
								Config:  cli.StringConfig{TrimSpace: true},
							},
						},
						Action: albumAddAction,
					},
					{
						Name:      "rename",
						Aliases:   []string{"mv"},
						Usage:     "Rename an album",
						UsageText: "gpcli album rename <album-key> <new-name>",
						Arguments: []cli.Argument{
							&cli.StringArg{
								Name:      "album-key",
								UsageText: "Album media key",
							},
							&cli.StringArg{
								Name:      "new-name",
								UsageText: "New album name",
							},
						},
						Action: albumRenameAction,
					},
					{
						Name:      "delete",
						Aliases:   []string{"rm"},
						Usage:     "Delete an album",
						UsageText: "gpcli album delete <album-key>",
						Arguments: []cli.Argument{
							&cli.StringArg{
								Name:      "album-key",
								UsageText: "Album media key",
							},
						},
						Action: albumDeleteAction,
					},
				},
			},
			{
				Name:  "upgrade",
				Usage: "Upgrade gpcli to latest or specific version",
				Arguments: []cli.Argument{
					&cli.StringArg{
						Name:      "version",
						UsageText: "<version> (optional, defaults to latest)",
					},
				},
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "check",
						Aliases: []string{"C"},
						Usage:   "Only check for updates without installing",
					},
					&cli.StringFlag{
						Name:    "url",
						Aliases: []string{"u"},
						Usage:   "Download and install from a specific URL",
					},
					&cli.BoolFlag{
						Name:    "nightly",
						Aliases: []string{"n"},
						Usage:   "Update to the latest nightly build",
					},
				},
				Action: upgradeAction,
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		slog.Error("command failed", "error", err)
		os.Exit(1)
	}
}
