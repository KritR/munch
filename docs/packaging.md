# Packaging

## Purpose and scope

This document defines how `munch` is built, packaged, released, and distributed.

It covers:

* release artifact shape
* supported target platforms
* GoReleaser configuration goals
* GitHub CI and CD responsibilities
* the planned expansion to package-manager distribution

It does not cover user-facing shell setup details. Those belong in `docs/install.md`.

## Goals

The packaging system should:

* ship a single binary named `munch`
* use GoReleaser as the source of truth for release artifact generation
* publish artifacts through GitHub Releases in iteration 1
* rely on embedded shell init scripts instead of separate shell asset installation
* support automated release publication through GitHub Actions
* leave room for Homebrew, Debian, Winget, and Nix in iteration 2

## Non-goals

The initial packaging phase does not aim to provide:

* automatic shell configuration mutation
* curl-pipe-sh as the primary install model
* package-manager publication in iteration 1
* signed/notarized installers in MVP

## Iteration 1 distribution strategy

Iteration 1 uses GitHub Releases as the canonical distribution channel.

The release pipeline should:

* build `munch` for the supported OS and architecture matrix
* produce compressed release archives
* generate a checksum file
* publish all artifacts to GitHub Releases

Iteration 1 should be sufficient for:

* direct manual installation
* CLI-assisted shell setup via `munch init zsh` and `munch init fish`
* self-contained shell integration through init scripts embedded in the binary
* future package-manager integrations that consume the same artifacts

## Release artifacts

Iteration 1 release artifacts should include:

* one archive per supported target
* a checksums file

Recommended archive naming:

* `munch_<version>_<os>_<arch>.tar.gz` for Unix targets
* `.zip` for Windows targets if Windows builds are introduced

Each release archive should contain:

* the `munch` binary
* optional lightweight metadata such as a release README or example config

## Supported targets

Iteration 1 target support should focus on:

* macOS arm64
* macOS amd64
* Linux amd64
* Linux arm64

Windows can remain out of scope for iteration 1 unless artifact generation is useful ahead of shell UX support.

For each target, the project should distinguish between:

* built targets
* manually smoke-tested targets

## GoReleaser configuration

GoReleaser should become the source of truth for release packaging.

The repository includes `.goreleaser.yaml` as the source of truth for release packaging.

The configuration defines:

* build matrix
* output binary name as `munch`
* archive naming
* checksum generation
* changelog/release note behavior
* GitHub release publishing

Iteration 1 does not need to configure package-manager publishers yet, but the GoReleaser layout should leave room for them.

Local validation:

```sh
goreleaser check
GOCACHE="$PWD/.gocache" goreleaser release --snapshot --clean
```

## GitHub CI

GitHub Actions should validate the repo on pushes and pull requests.

The repository includes `.github/workflows/ci.yml`.

CI runs:

* formatting checks
* test suite execution
* build verification
* GoReleaser config validation

The goal of CI is fast signal on correctness and release readiness, not full release publication.

## GitHub CD and release automation

Release automation should be tag-driven.

The repository includes `.github/workflows/release.yml`.

Release model:

* a version tag such as `v0.1.0` triggers the release workflow
* GitHub Actions runs GoReleaser
* GoReleaser builds artifacts and publishes a GitHub Release

The release workflow should be responsible for:

* building the release matrix
* generating checksums
* publishing artifacts
* attaching release notes or changelog content

## Release process

The expected release flow should be:

1. prepare release changes
2. tag the release
3. GitHub Actions runs the release workflow
4. GoReleaser builds and publishes artifacts
5. smoke-test one or more downloaded artifacts

Recommended manual commands:

```sh
jj bookmark set main -r @-
jj git push --bookmark main
git tag v0.1.0
git push origin v0.1.0
```

## Iteration 2 expansion

Iteration 2 should add package-manager distribution.

Planned channels:

* Homebrew
* Debian package support
* Winget
* Nix

Those additions should build on the iteration 1 artifact model rather than replace it.

## Open questions

Open questions to revisit as packaging is implemented:

* whether Windows artifacts should be built before Windows shell UX exists
* whether release notes should be fully automated or lightly curated
* whether signing/notarization becomes necessary before broader adoption
