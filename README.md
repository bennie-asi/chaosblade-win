# chaosblade-win

A Go CLI skeleton inspired by ChaosBlade, built with Cobra for Windows environments.

## Requirements
- Go 1.20+

## Quick start
1. Install dependencies: `go mod tidy`
2. Run the CLI: `go run .`
3. Build binaries: `go build ./...`

Note: When using `go run .`, place the subcommand immediately after the `.`. Do not insert `--` between `.` and the subcommand (for example, prefer `go run . create cpu ...` over `go run . -- create cpu ...`). Some shells or `go run` argument forwarding can cause the latter form to drop args.

## Examples
- Start a bounded CPU load for 45s on two cores (foreground): `chaosblade-win create cpu load --cores 2 --percent 60 --duration 45s`
- Start the same CPU experiment detached (returns immediately with id and pid): `chaosblade-win create cpu load --cores 2 --percent 60 --duration 45s --detach`
- Stop the tracked CPU experiment (all or last): `chaosblade-win destroy cpu`
- Stop a specific experiment by id: `chaosblade-win destroy cpu <experiment-id>`
- List tracked experiments for a target: `chaosblade-win list cpu`
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
 - Tracking: experiments write per-experiment state under the system temp directory in a per-target subfolder, e.g. `%TMP%/chaosblade-win/<target>/<id>.json`. `create` prints the created experiment id and `destroy <target> <id>` can be used to stop a specific experiment. Omitting the id will attempt to stop all tracked experiments for the target.
- Spec: target/action metadata in spec/ drives CLI descriptions; extend it when adding new experiments.
