# Implementation Plan

## Purpose

This document translates the current design set into a practical implementation path for MVP.

It covers:

* recommended implementation phases
* suggested Go package and binary boundaries
* environment setup prerequisites
* where manual testing is required
* open assumptions that should be validated during implementation

This plan assumes:

* the application is implemented in Go
* the first remote provider is Cerebras

## Implementation goals

The implementation plan should optimize for:

* proving the shell boundary early
* keeping component boundaries aligned with the design docs
* enabling local testing before real provider integration
* introducing remote/provider complexity only after the runtime works locally
* preserving a path to iterate safely through tests and manual shell verification

## Recommended execution strategy

Build the MVP as a sequence of vertical slices rather than trying to implement every subsystem at once.

Recommended order:

1. prove shell launch and result application
2. add the widget runtime skeleton and protocol types
3. add local config loading and logging
4. add a fake suggestion pipeline
5. add real provider integration
6. add safety evaluation and confirmation
7. harden with automated and manual testing

The highest-risk surface is still shell integration. That should be validated before investing heavily in prompt/provider logic.

## Suggested Go project shape

The repo currently has no code, so the initial layout can be kept clean.

Suggested structure:

```text
/cmd
  /munch-widget
  /munch-dev
/internal
  /config
  /protocol
  /runtime
  /shell
    /zsh
    /fish
  /suggest
  /provider
    /cerebras
  /safety
  /context
  /telemetry
/testdata
/docs
```

Suggested ownership:

* `cmd/munch-widget`
  The main widget process binary invoked by shell adapters.
* `cmd/munch-dev`
  Optional helper binary for local development, fixture playback, or protocol testing.
* `internal/protocol`
  Request and response types, validation, JSON encode/decode helpers.
* `internal/config`
  TOML loading, defaults, validation, effective settings.
* `internal/runtime`
  Session orchestration and state machine implementation.
* `internal/shell/zsh`
  Zsh-specific integration helpers or emitted scripts/templates.
* `internal/shell/fish`
  Fish-specific integration helpers or emitted scripts/templates.
* `internal/context`
  Local context collection and normalization.
* `internal/suggest`
  Suggestion engine and prompt construction orchestration.
* `internal/provider`
  Provider-neutral interfaces.
* `internal/provider/cerebras`
  First concrete provider client implementation.
* `internal/safety`
  Local risk heuristics and confirmation policy mapping.
* `internal/telemetry`
  Structured logs and telemetry event emission.

## Environment setup

Before implementation, set up the basic local development environment.

### Required tools

* Go toolchain
* Zsh
* Fish
* a terminal environment where both shells can be launched interactively

### Provider setup

For Cerebras integration, plan for:

* an API key stored in an environment variable
* a config entry naming that environment variable
* a way to switch between fake provider mode and real provider mode during development

### Local shell setup

You will need:

* a way to source a generated or hand-written Zsh widget binding during development
* a way to source a generated or hand-written Fish binding during development
* a repeatable shell smoke-test setup so you can quickly rerun manual flows after changes

### Suggested local config setup

Create a developer config file early with:

* `safety.level = "balanced"`
* `provider.model` set for Cerebras
* `provider.api_key_env` pointing at your local secret env var
* `provider.timeout_ms = 4000`
* `provider.max_retries = 1`
* `telemetry.enabled = false`

## Phase 1: Protocol and config foundation

### Goal

Get the non-UI boundaries stable first.

### Deliverables

* Go types for `ShellInvocationRequest` and `ShellInvocationResponse`
* JSON encode/decode and validation
* config loader with defaults and warnings
* request ID generation
* structured logging baseline

### Automated testing

Add tests for:

* valid and invalid protocol payloads
* enum validation
* conditional `command` requirements
* unknown field tolerance
* config defaulting and fallback behavior

### Manual testing

No shell testing required yet if this phase stays library-only.

## Phase 2: Shell boundary proof of concept

### Goal

Prove that shell adapters can launch the Go widget binary and safely apply `cancel`, `insert`, and `execute`.

### Deliverables

* minimal `munch-widget` binary
* Zsh adapter prototype
* Fish adapter prototype
* hardcoded or fixture-driven final actions

At this stage, the widget does not need real suggestion generation.

### Success criteria

* shell launches widget process
* shell sends JSON request
* widget returns JSON response
* `cancel` restores original buffer
* `insert` replaces full buffer
* `execute` replaces and runs

### Automated testing

Add tests for:

* request creation from shell-local state
* JSON request/response process boundary
* response handling for `cancel`, `insert`, and `execute`
* malformed widget response fallback

### Manual testing

Required:

* Zsh cancel flow
* Zsh insert flow
* Zsh execute flow
* Fish cancel flow
* Fish insert flow
* Fish execute flow
* multiline seeded prompt passthrough

This is the first mandatory manual test phase. Real shell behavior should be checked before moving on.

## Phase 3: Runtime skeleton and fake suggestions

### Goal

Build the widget runtime and state machine without depending on a real provider.

### Deliverables

* runtime session object
* state machine implementation
* prompt seeding behavior
* debounce handling
* fixture or fake suggestion engine
* suggestion selection flow
* terminal final action creation

### Automated testing

Add tests for:

* state transitions
* stale response suppression
* empty results flow
* recoverable error metadata behavior
* `Completing` and `Closed` behavior

### Manual testing

Required:

* seeded prompt editing
* loading indicator behavior
* empty result state
* recoverable error shown in header
* cancel from different runtime states

## Phase 4: Prompt construction and provider abstraction

### Goal

Introduce the real provider boundary without tying the rest of the runtime to Cerebras-specific details.

### Deliverables

* provider-neutral request interface
* provider-neutral response payload
* prompt assembly from `prompting.md`
* provider client interface
* fake provider implementation for tests

### Automated testing

Add tests for:

* prompt assembly from curated context
* schema/output expectations
* provider client interface behavior with fakes
* advisory risk preservation

### Manual testing

Minimal manual testing here unless prompt rendering needs direct inspection during development.

## Phase 5: Cerebras provider integration

### Goal

Add the first live provider implementation behind the provider client boundary.

### Deliverables

* `internal/provider/cerebras`
* request translation for Cerebras
* JSON structured-output handling
* timeout handling
* single retry on transient failures
* single retry on malformed structured output

### Automated testing

Use mocks/fakes for most tests:

* success path
* timeout
* auth/config failure
* malformed structured output
* retry once on transient failure
* retry once on malformed structured output

### Manual testing

Required:

* real-provider happy path with Cerebras
* provider timeout surfaced as recoverable widget error
* malformed output path if reproducible in a controlled dev mode
* invalid or missing API key behavior

### Environment assumptions to validate

This is the first place where implementation assumptions need confirmation:

* which Cerebras API shape best supports strict JSON output
* whether Cerebras behavior is reliable enough without few-shot examples
* whether one retry is sufficient in practice

If these assumptions fail, update `provider-integration.md` and `prompting.md` rather than patching around the issue silently in code.

## Phase 6: Safety evaluation and confirmation

### Goal

Apply local safety after provider suggestions and before final action.

### Deliverables

* heuristic risk classifier
* policy mapping from risk to `requires_confirmation`
* confirmation UI/state
* concise reason strings for confirmation

### Automated testing

Add tests for:

* `low` / `medium` / `high` classification
* `low` / `balanced` / `strict` policy mapping
* `sudo` classification
* redirection bump behavior
* confirmation-required state transitions

### Manual testing

Required:

* read-only command with no confirmation
* medium-risk command requiring confirmation under `balanced`
* high-risk command requiring confirmation
* cancel from confirmation
* execute after confirmation

## Phase 7: Harden the implementation

### Goal

Bring the implementation up to the release bar defined in `testing-strategy.md`.

### Deliverables

* full contract test suite
* shell smoke tests
* runtime integration tests
* provider mock coverage
* release checklist

### Manual testing

Run the manual matrix from `testing-strategy.md` as a full pass.

## Recommended first milestone

The first implementation milestone should be intentionally narrow.

### Milestone 1

Deliver:

* protocol types and validation
* config loading
* a minimal widget binary
* Zsh adapter
* Fish adapter
* correct handling of `cancel`, `insert`, and `execute`

Do not include yet:

* real provider integration
* real suggestion generation
* safety confirmation

The purpose of Milestone 1 is to prove that the shell/process contract is solid.

## Manual testing checkpoints

Manual testing is unavoidable at these points:

* after first Zsh adapter integration
* after first Fish adapter integration
* after runtime UI state flow exists
* after real Cerebras integration lands
* after safety confirmation lands
* before any release candidate

If a change affects shell redraw, buffer restoration, or execute behavior, re-run shell manual tests even if unit tests still pass.

## Open implementation assumptions

The following assumptions are reasonable now, but should be revisited once code exists:

* a single Go widget binary is sufficient for MVP
* shell bindings can remain thin and mostly static
* the widget UI can be implemented cleanly in Go without introducing a second runtime
* Cerebras can reliably return the canonical JSON shape with the prompting strategy in `prompting.md`
* request/response JSON over `stdin` and `stdout` remains simple enough not to require a helper subprocess layer

If one of these turns out false during implementation, update the relevant design docs before expanding the workaround into more code.

## Recommended next step

Start with Phase 1 and Phase 2 together:

* scaffold the Go module
* implement `internal/protocol`
* implement `internal/config`
* create a minimal `cmd/munch-widget`
* wire one shell adapter path end to end

Zsh is the best first adapter because it gives you a fast way to prove the contract, then Fish can follow against the same widget binary.
