# chaosblade-win

A Go CLI skeleton inspired by ChaosBlade, built with Cobra for Windows environments.

## Requirements
- Go 1.20+

## Quick start
1. Install dependencies: `go mod tidy`
2. Run the CLI: `go run .`
3. Build binaries: `go build ./...`

## Examples
- Start a bounded CPU load for 45s on two cores: `chaosblade-win create cpu load --cores 2 --percent 60 --duration 45s`
- Stop the tracked CPU experiment: `chaosblade-win destroy cpu`
- Fill disk with ~1 GB (keeps 64 MB headroom when using --percent): `chaosblade-win create disk fill --size 1024`
- Allocate ~25% of memory: `chaosblade-win create mem load --percent 25`
- Start network delay/loss/bandwidth (requires WinDivert): `chaosblade-win create net delay 120 --jitter 40 --loss 1.5 --bandwidth 500 --filter "outbound and tcp"`
- Tear down any network experiment: `chaosblade-win destroy net`

## Project layout
- cmd/: Cobra commands and entrypoints.
- exec/: Execution logic abstractions for running experiments.
- spec/: Experiment model definitions (targets/actions/flags) that drive CLI descriptions.

## Development notes
- A VS Code task `go build` is available to run the build via the Tasks interface.
- Add new commands under cmd/ by creating additional files that attach to `rootCmd`.
- Implement experiment execution in exec/ and expand models in spec/ as features grow.

## Safety notes
- CPU: `--percent` is validated to the 1-100 range; use `--duration` to auto-stop in unattended runs.
- Disk: `create disk fill --percent` keeps at least 64 MB free; verify the path is correct before running.
- Memory: the allocator enforces a minimum of 1 MB and respects the computed percent of total memory; use conservative percentages on production hosts.
- Tracking: experiments write PID markers under the system temp directory (chaosblade-win/*.json); use `chaosblade-win destroy <target>` to stop and clean markers.
- Spec: target/action metadata in spec/ drives CLI descriptions; extend it when adding new experiments.
