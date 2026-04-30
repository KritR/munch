# Implementation Status

## Purpose

This document tracks how the current implementation of `munch` aligns with the design docs.

It replaces the earlier pre-build implementation roadmap. The project now has a working shell bridge, a live Bubble Tea UI, provider-backed suggestions, and local safety confirmation, so the useful question is no longer "what should we build first?" but:

* what is already implemented
* what is partially implemented
* where the code has intentionally diverged from the original plan
* what should be built next

This document should be updated as implementation reality changes.

## Current implementation snapshot

Current code layout:

```text
/cmd
  /munch-widget
  /bt-debug
/internal
  /app
  /bridge
    /zsh
    /fish
  /config
  /context
  /prompting
  /protocol
  /provider
    /cerebras
    /fake
  /runtime
  /safety
  /suggest
  /ui
/shell
  munch.zsh
  munch.fish
/docs
```

Key runtime pieces currently exist:

* shell bridge modes for Zsh and Fish
* protocol types and validation
* config loading and defaults
* shell-aware context collection
* provider-neutral suggestion engine
* fake provider and Cerebras provider
* local safety evaluation and confirmation
* live Bubble Tea prompt with debounce and async suggestion refresh

## Design alignment summary

Overall, the implementation is broadly aligned with the design docs, but there are a few important drifts that should be treated as explicit architectural decisions rather than accidental inconsistencies.

### Aligned areas

These design areas are implemented in a form that matches the current docs closely:

* prompt seeding from the shell buffer
* thin shell adapters
* one-shot widget process model
* provider abstraction
* local safety as the authoritative confirmation layer
* shell-aware context collection
* Bubble Tea-based widget UI
* Zsh and Fish support

### Intentional drift from the original plan

These are real implementation changes relative to the earlier planning assumptions:

* the shell adapter no longer sends JSON over `stdin` directly
* bridge modes now use shell-local env vars in and shell-safe assignments out
* the package layout uses `internal/app`, `internal/bridge`, and `internal/ui` rather than `internal/shell`
* the live UI now owns debounce and async generation rather than the runtime precomputing suggestions before opening the UI

These changes are good changes, but the docs should describe them explicitly.

### Partial implementation areas

The following design areas exist but are not complete yet:

* logging exists, but there is no dedicated telemetry package or event model yet
* packaging and install flow are not formalized
* the release-hardening phase is incomplete

## Phase status

## Phase 1: Protocol and config foundation

Status: complete

Implemented:

* `internal/protocol`
* request/response validation
* config defaults and TOML loading
* request IDs
* quiet-by-default logging with optional file logging

## Phase 2: Shell boundary proof of concept

Status: complete

Implemented:

* Zsh shell bridge
* Fish shell bridge
* `cancel`, `insert`, and `execute` handling at the shell boundary
* bridge tests for both shells

Important note:

The original plan assumed JSON request/response transport directly between shell adapters and the widget process. The real implementation now uses shell bridge modes:

* env vars carry shell-local input into the widget
* shell-safe assignments are emitted back to the shell

This should now be treated as the MVP shell transport design.

## Phase 3: Runtime skeleton and fake suggestions

Status: complete

Implemented:

* runtime session model
* fake provider path
* live editable prompt UI
* debounce and stale-result suppression
* loading, empty, and recoverable error states in the UI

Important note:

The implementation moved debounce and async generation into the UI layer rather than having the runtime precompute suggestions before the UI opens. This is the right architecture for the current UX and should be reflected in the docs.

## Phase 4: Prompt construction and provider abstraction

Status: complete

Implemented:

* provider-neutral request/response interface
* prompt construction from `prompting.md`
* fake provider implementation
* provider-backed suggestion engine

## Phase 5: Cerebras provider integration

Status: complete

Implemented:

* `internal/provider/cerebras`
* structured JSON output handling
* timeout handling
* retry behavior
* real manual provider smoke testing

## Phase 6: Safety evaluation and confirmation

Status: complete

Implemented:

* local risk classifier
* policy mapping to `requires_confirmation`
* confirmation UI flow
* confirmation reason strings

Also implemented:

* first-class interactive `execute` action in the Bubble Tea UI
* confirmation flow that preserves insert vs execute intent

## Phase 7: Hardening

Status: in progress

Implemented so far:

* protocol tests
* config tests
* bridge tests
* context parsing tests
* provider tests
* runtime tests
* safety tests
* UI interaction tests

Still needed:

* higher-level app/session integration tests
* fuller failure-path coverage around provider and UI state interactions
* release checklist
* explicit install/setup docs for real users

## Completed milestones

The original Milestone 1 is fully complete:

* protocol types and validation
* config loading
* minimal widget binary
* Zsh adapter
* Fish adapter
* correct `cancel`, `insert`, and `execute` handling at the shell boundary

The project is now well past Milestone 1.

## Current gaps

The highest-value remaining gaps are:

1. packaging and install flow
2. telemetry/event model
3. hardening and release checks

## Recommended next phases

## Phase 8: Action completion and UX polish

Status: complete

Completed:

* first-class interactive `execute` action
* live UI key help for insert vs execute
* terminal cleanup on UI exit
* selected-row and confirmation presentation improvements

Note:

There may still be minor follow-up polish in terminal cleanup and theme tuning, but the primary Phase 8 goals are now complete.

## Phase 9: Packaging and install

Recommended deliverables:

* install instructions for Zsh and Fish
* shell init snippets
* `make install` or equivalent local install workflow
* config bootstrap guidance

## Phase 10: Observability and supportability

Recommended deliverables:

* proper `internal/telemetry`
* structured event model
* better diagnostics for provider and UI issues
* debug tooling that does not leak into normal shell output

## Manual testing status

Manual testing has already validated:

* Zsh insert/cancel/execute shell behavior
* Fish insert/cancel/execute shell behavior
* real Cerebras integration
* live prompt editing
* confirmation flow
* UI color behavior in real terminals

Still worth validating explicitly as features land:

* execute from the interactive UI once implemented
* provider timeout and malformed output behavior in the current live UI
* release-level smoke tests after packaging/install work

## Open implementation assumptions

The following assumptions are still active and should be revisited as the MVP matures:

* one Go widget binary remains sufficient
* bridge-mode shell transport remains simpler than reintroducing direct shell-side JSON handling
* Bubble Tea remains the right UI runtime for the product
* Cerebras continues to behave reliably with the current prompting and structured-output strategy

If any of these change, update the relevant design docs rather than letting the code drift silently.
