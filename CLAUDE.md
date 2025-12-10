# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

```bash
make build      # Build to ./gpcli
make clean      # Remove built binary
go build -o gpcli .   # Direct build command
```

## Protobuf Generation

See `.proto/README.md` for detailed instructions on generating protobuf files.

Quick reference from project root:
```bash
export PATH=$PATH:$(go env GOPATH)/bin

# Generate single file
protoc --proto_path=. --go_out=. --go_opt=M.proto/MessageName.proto=/pb .proto/MessageName.proto

# Generate all files
for proto in .proto/*.proto; do
  name=$(basename "$proto" .proto)
  protoc --proto_path=. --go_out=. --go_opt=M.proto/${name}.proto=/pb .proto/${name}.proto
done
```

## Architecture

This is a CLI tool for managing Google Photos using an unofficial API. It uses protobuf for API communication.

### Key Components

- **cli/** - CLI command definitions using urfave/cli/v3
  - `cli.go` - Command definitions and actions for upload, download, thumbnail, auth, delete, archive, favourite, caption
  - `config.go` - YAML config file management. Stores credentials and settings.
- **gogpm/** - Core library package
  - `api.go` - GooglePhotosAPI struct embedding core.Api with upload management
  - `manager.go` - Upload orchestration with worker pool. Emits progress events.
  - `sha1.go` - File hash calculation
- **gogpm/core/** - Low-level API operations
  - `api.go` - Api struct with auth token management and common headers
  - `upload.go` - Upload token, file upload, commit operations
  - `download.go` - Download URL retrieval
  - `trash.go` - MoveToTrash, RestoreFromTrash operations
  - `archive.go` - SetArchived operation
  - `metadata.go` - SetCaption, SetFavourite operations
  - `album.go` - CreateAlbum, AddMediaToAlbum operations
  - `thumbnail.go` - Thumbnail download
  - `utils.go` - SHA1ToDedupeKey, ToURLSafeBase64
- **pb/** - Protobuf-generated Go code for API request/response structures
- **.proto/** - Protobuf definitions for Google Photos API messages

### Event-Based Progress System

The upload system uses an event callback pattern:
1. `GooglePhotosAPI.Upload()` starts worker goroutines
2. Workers emit events via callback function
3. `cli.go` receives events and prints progress to stdout

Event types: `uploadStart`, `ThreadStatus`, `FileStatus`, `uploadStop`

### Config File

Config is stored in `./gpcli.config` (YAML) or custom path via `--config` flag. Contains credentials array and upload settings.

## Rules

- always implement root fixes and never add patch fixes
