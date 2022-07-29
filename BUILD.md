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
cobra-cli add display --license mit --author "Open Traffic Generator"
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
   1.1 Port metrics

```Shell
cat test/transform/port_metrics.json | ./otgen transform                   | diff test/transform/port_metrics_passthrough.json -
cat test/transform/port_metrics.json | ./otgen transform -m port           | diff test/transform/port_metrics_frames.json -
cat test/transform/port_metrics.json | ./otgen transform -m port -c frames | diff test/transform/port_metrics_frames.json -
cat test/transform/port_metrics.json | ./otgen transform -m port -c bytes  | diff test/transform/port_metrics_bytes.json -
cat test/transform/port_metrics.json | ./otgen transform -m port -c pps    | diff test/transform/port_metrics_frame_rate.json -
cat test/transform/port_metrics.json | ./otgen transform -m port -c tput   | diff test/transform/port_metrics_byte_rate.json -
```

   1.2 Flow metrics

```Shell
cat test/transform/flow_metrics.json | ./otgen transform                   | diff test/transform/flow_metrics_passthrough.json -
cat test/transform/flow_metrics.json | ./otgen transform -m flow           | diff test/transform/flow_metrics_frames.json -
cat test/transform/flow_metrics.json | ./otgen transform -m flow -c frames | diff test/transform/flow_metrics_frames.json -
cat test/transform/flow_metrics.json | ./otgen transform -m flow -c bytes  | diff test/transform/flow_metrics_bytes.json -
cat test/transform/flow_metrics.json | ./otgen transform -m flow -c pps    | diff test/transform/flow_metrics_frame_rate.json -
````

2. Templates - JSON
   2.1 Port metrics

```Shell
cat test/transform/port_metrics.json | ./otgen transform -f templates/transformPassThrough.tmpl   | diff test/transform/port_metrics_passthrough.json -
cat test/transform/port_metrics.json | ./otgen transform -f templates/transformPortFrames.tmpl    | diff test/transform/port_metrics_frames.json -
cat test/transform/port_metrics.json | ./otgen transform -f templates/transformPortBytes.tmpl     | diff test/transform/port_metrics_bytes.json -
cat test/transform/port_metrics.json | ./otgen transform -f templates/transformPortFrameRate.tmpl | diff test/transform/port_metrics_frame_rate.json -
cat test/transform/port_metrics.json | ./otgen transform -f templates/transformPortByteRate.tmpl  | diff test/transform/port_metrics_byte_rate.json -
````

   2.2 Flow metrics

```Shell
cat test/transform/flow_metrics.json | ./otgen transform -f templates/transformPassThrough.tmpl   | diff test/transform/flow_metrics_passthrough.json -
cat test/transform/flow_metrics.json | ./otgen transform -f templates/transformFlowFrames.tmpl    | diff test/transform/flow_metrics_frames.json -
cat test/transform/flow_metrics.json | ./otgen transform -f templates/transformFlowBytes.tmpl     | diff test/transform/flow_metrics_bytes.json -
cat test/transform/flow_metrics.json | ./otgen transform -f templates/transformFlowFrameRate.tmpl | diff test/transform/flow_metrics_frame_rate.json -
````


3. Templates - Tables
   3.1 Port metrics

```Shell
cat test/transform/port_metrics.json | ./otgen transform -f templates/transformPortFramesTable.tmpl    | diff test/transform/port_metrics_frames_table.txt -
cat test/transform/port_metrics.json | ./otgen transform -f templates/transformPortBytesTable.tmpl     | diff test/transform/port_metrics_bytes_table.txt -
cat test/transform/port_metrics.json | ./otgen transform -f templates/transformPortFrameRateTable.tmpl | diff test/transform/port_metrics_frame_rate_table.txt -
cat test/transform/port_metrics.json | ./otgen transform -f templates/transformPortByteRateTable.tmpl  | diff test/transform/port_metrics_byte_rate_table.txt -
````

   3.2 Flow metrics

```Shell
cat test/transform/flow_metrics.json | ./otgen transform -f templates/transformFlowFramesTable.tmpl    | diff test/transform/flow_metrics_frames_table.txt -
cat test/transform/flow_metrics.json | ./otgen transform -f templates/transformFlowBytesTable.tmpl     | diff test/transform/flow_metrics_bytes_table.txt -
cat test/transform/flow_metrics.json | ./otgen transform -f templates/transformFlowFrameRateTable.tmpl | diff test/transform/flow_metrics_frame_rate_table.txt -
````

4. Full pipe with port metrics

```Shell
cat ../otg.b2b.json | ./otgen run -k 2>/dev/null | ./otgen transform -m port
````

### `display`

Currently, only for visual inspection

1. Charts

```Shell
cat test/transform/port_metrics.json | ./test/transform/delay.sh 0.5 | ./otgen transform -m port -c frames | ./otgen display --mode chart --type line
cat test/transform/port_metrics.json | ./test/transform/delay.sh 0.5 | ./otgen transform -m port -c bytes  | ./otgen display --mode chart --type line
cat test/transform/port_metrics.json | ./test/transform/delay.sh 0.5 | ./otgen transform -m port -c pps    | ./otgen display --mode chart --type line
cat test/transform/port_metrics.json | ./test/transform/delay.sh 0.5 | ./otgen transform -m port -c tput   | ./otgen display --mode chart --type line
````

2. Table

```Shell
cat test/transform/port_metrics.json | ./test/transform/delay.sh 0.5 | ./otgen transform -m port -c frames | ./otgen display --mode table
cat test/transform/port_metrics.json | ./test/transform/delay.sh 0.5 | ./otgen transform -m port -c bytes  | ./otgen display --mode table
cat test/transform/port_metrics.json | ./test/transform/delay.sh 0.5 | ./otgen transform -m port -c pps    | ./otgen display --mode table
cat test/transform/port_metrics.json | ./test/transform/delay.sh 0.5 | ./otgen transform -m port -c tput   | ./otgen display --mode table
````