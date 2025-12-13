# Protocol Buffers

This directory contains `.proto` files for Google Photos API communication.

## Proto Files

- **album.proto** - Album operations (create, add media, delete, rename)
- **archive.proto** - Archive/unarchive media items
- **download.proto** - Download URL retrieval for media items
- **hash_lookup.proto** - Find existing media by SHA1 hash (deduplication)
- **metadata.proto** - Media metadata operations (caption, favourite)
- **trash.proto** - Trash operations (move to trash, restore, permanent delete)
- **upload.proto** - File upload flow (tokens, commit, responses)

## Generating Go Code

### Prerequisites

```bash
# Install protoc-gen-go
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
```

### Generate All

From the project root directory:

```bash
export PATH=$PATH:$(go env GOPATH)/bin

for proto in .proto/*.proto; do
  protoc --proto_path=. --go_out=. --go_opt=module=github.com/viperadnan-git/go-gpm "$proto"
done
```

### Generate Single File

```bash
protoc --proto_path=. --go_out=. --go_opt=module=github.com/viperadnan-git/go-gpm .proto/name.proto
```

## Creating New Proto Files

Use [blackboxprotobuf](https://github.com/nccgroup/blackboxprotobuf) to reverse-engineer proto definitions from encoded messages:

```python
# pip install bbpb
import blackboxprotobuf

protobuf_hex = "hex_encoded_message"
message_name = "FindMediaByHashRequest"
proto_filename = "hash_lookup"

protobuf_bytes = bytes.fromhex(protobuf_hex)
decoded_data, message_type = blackboxprotobuf.decode_message(protobuf_bytes)

blackboxprotobuf.export_protofile({message_name: message_type}, f".proto/{proto_filename}.proto")
```

After creating the proto file, add the go_package option and use descriptive message names:

```protobuf
syntax = "proto3";

option go_package = "github.com/viperadnan-git/go-gpm/internal/pb";

// FindMediaByHashRequest checks if media with given hash already exists
message FindMediaByHashRequest {
  // fields
}
```
