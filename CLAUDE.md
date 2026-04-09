# twofer — Claude Code Context

## Dependency licenses (depot)

A GitHub Actions job checks every dependency has a recognized license on pushes to `master` that touch `go.mod`. When adding, upgrading, or downgrading dependencies, run `/depot-accept twofer` on your feature branch before merging to master — the CI only fires on merge, so this keeps your PR self-contained. Also use it reactively after a CI failure on master.
