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


## Test

### `transform`

1. Parameters

```Shell
cat test/transform/port_metrics.json | ./otgen transform | diff test/transform/port_metrics_passthrough.json -
cat test/transform/port_metrics.json | ./otgen transform -m port | diff test/transform/port_metrics_frames.json -
````

2. Templates - JSON

```Shell
cat test/transform/port_metrics.json | ./otgen transform -f templates/transformPassThrough.tmpl   | diff test/transform/port_metrics_passthrough.json -
cat test/transform/port_metrics.json | ./otgen transform -f templates/transformPortFrames.tmpl    | diff test/transform/port_metrics_frames.json -
cat test/transform/port_metrics.json | ./otgen transform -f templates/transformPortBytes.tmpl     | diff test/transform/port_metrics_bytes.json -
cat test/transform/port_metrics.json | ./otgen transform -f templates/transformPortFrameRate.tmpl | diff test/transform/port_metrics_rate.json -
````

3. Templates - Tables

```Shell
cat test/transform/port_metrics.json | ./otgen transform -f templates/transformPortFramesTable.tmpl | diff test/transform/port_metrics_frames_table.txt -
cat test/transform/port_metrics.json | ./otgen transform -f templates/transformPortBytesTable.tmpl | diff test/transform/port_metrics_bytes_table.txt -
cat test/transform/port_metrics.json | ./otgen transform -f templates/transformPortFrameRateTable.tmpl | diff test/transform/port_metrics_rate_table.txt -
````

4. Full pipe with port metrics

```Shell
cat ../otg.b2b.json | ./otgen run -k 2>/dev/null | ./otgen transform -m port
````