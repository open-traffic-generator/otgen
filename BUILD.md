# Building instructions

## Prerequisites

1. Go
2. Cobra-cli
3. GoReleaser https://goreleaser.com/install/, for example `go install github.com/goreleaser/goreleaser@v1.6`

## Build history

### Cobra

```Shell
go mod init github.com/open-traffic-generator/otgen
cobra-cli init --license mit --author "Open Traffic Generator"
cobra-cli add run --license mit --author "Open Traffic Generator"
cobra-cli add version --license mit --author "Open Traffic Generator"
cobra-cli add transform --license mit --author "Open Traffic Generator"
````

### GoReleaser

```Shell
goreleaser init
goreleaser build --single-target --snapshot --rm-dist
goreleaser release --snapshot --rm-dist
````

### Build

```Shell
go get
go mod tidy
go build -ldflags="-X 'github.com/open-traffic-generator/otgen/cmd.version=v0.0.0-${USER}'"
````
