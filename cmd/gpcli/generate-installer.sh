#!/bin/sh

$(go env GOPATH)/bin/binst init --source=goreleaser --file=.goreleaser.yml -o binstaller.yml --name $(basename $(pwd)) --repo viperadnan-git/gogpm
$(go env GOPATH)/bin/binst gen --config=binstaller.yml -o install.sh