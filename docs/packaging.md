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
* publish a Homebrew cask to `KritR/homebrew-munch` automatically on tagged releases
* rely on embedded shell init scripts instead of separate shell asset installation
* support automated release publication through GitHub Actions
* keep Linux distribution simple through direct release artifacts for now

## Non-goals

The initial packaging phase does not aim to provide:

* automatic shell configuration mutation
* curl-pipe-sh as the primary install model
* Linux package-manager integration through Homebrew

## Distribution strategy

GitHub Releases remain the canonical source of release artifacts.

The release pipeline should:

* build `munch` for the supported OS and architecture matrix
* produce Linux release archives
* produce a signed and notarized universal macOS binary archive
* generate a checksum file
* publish all artifacts to GitHub Releases
* update the Homebrew tap cask from the same release metadata

The current distribution setup is sufficient for:

* direct manual installation on Linux
* Homebrew cask installation on macOS through `KritR/homebrew-munch`
* CLI-assisted shell setup via `munch init zsh` and `munch init fish`
* self-contained shell integration through init scripts embedded in the binary
* future package-manager integrations that consume the same artifacts

## Release artifacts

Release artifacts should include:

* one tarball per supported Linux target
* one universal binary archive for macOS
* a checksums file

Recommended archive naming:

* `munch_<version>_linux_<arch>.tar.gz` for Linux
* `munch_<version>_darwin_universal.zip` for macOS
* `.zip` for Windows targets if Windows builds are introduced

Each Linux release archive should contain:

* the `munch` binary
* optional lightweight metadata such as a release README or example config

The macOS release archive should contain a universal `munch` binary that is signed and notarized through GoReleaser's cross-platform notarization flow.

The Homebrew cask should reference that same tagged universal archive and checksum rather than rebuilding anything during install.

## Supported targets

Current target support focuses on:

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

The GoReleaser configuration remains the source of truth for the release artifacts and checksum file. It also defines the macOS universal-binary build, cross-platform notarization, and Homebrew cask publishing flow.

Local validation:

```sh
goreleaser check
GOCACHE="$PWD/.gocache" goreleaser release --snapshot --clean
```

If the local environment does not already have module dependencies available, also set a workspace-local module cache:

```sh
GOCACHE="$PWD/.gocache" GOMODCACHE="$PWD/.gomodcache" goreleaser release --snapshot --clean --skip=publish
```

This packaging model depends on GoReleaser capabilities for:

* `universal_binaries` to merge the macOS `amd64` and `arm64` builds
* `notarize.macos` to sign and notarize the universal binary
* `homebrew_casks` to publish the tap recipe

## GitHub CI

GitHub Actions should validate the repo on pushes and pull requests.

The repository includes `.github/workflows/ci.yml`.

CI runs:

* formatting checks
* test suite execution
* build verification
* GoReleaser config validation
* snapshot release packaging validation

The goal of CI is fast signal on correctness and release readiness, not full release publication.

## GitHub CD and release automation

Release automation should be tag-driven.

The repository includes `.github/workflows/release.yml`.

Release model:

* a version tag such as `v0.1.0` triggers the release workflow
* GitHub Actions runs GoReleaser
* GoReleaser builds artifacts, signs and notarizes the universal macOS binary, publishes a GitHub Release, and pushes the Homebrew cask

The release workflow should be responsible for:

* building the release matrix
* generating checksums
* publishing artifacts
* updating the Homebrew tap cask in `KritR/homebrew-munch`
* attaching release notes or changelog content

The release workflow requires:

* the default `GITHUB_TOKEN` for publishing the GitHub Release in `KritR/munch`
* a repository secret named `HOMEBREW_TAP_GITHUB_TOKEN` with permission to push to `KritR/homebrew-munch`
* `MACOS_SIGN_P12`, containing the base64-encoded Apple signing certificate export
* `MACOS_SIGN_PASSWORD`, the password for that certificate export
* `MACOS_NOTARY_KEY`, containing the base64-encoded App Store Connect `.p8` key
* `MACOS_NOTARY_KEY_ID`, the App Store Connect key ID
* `MACOS_NOTARY_ISSUER_ID`, the App Store Connect issuer UUID

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

## Future expansion

Future work can add more package-manager distribution on top of the same artifact model.

Planned channels:

* Debian package support
* Winget
* Nix

Those additions should build on the iteration 1 artifact model rather than replace it.

## Open questions

Open questions to revisit as packaging is implemented:

* whether the non-stapled universal-binary notarization path is sufficient for the intended macOS UX
* whether Linux should move to a native package channel after the release-tarball phase
* whether release notes should be fully automated or lightly curated
