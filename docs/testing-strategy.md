# Testing Strategy

## Purpose and scope

This document defines the MVP testing strategy for `munch`.

It covers:

* what should be tested
* which test layers provide confidence
* what should be automated versus tested manually
* which failure classes are release-blocking

This document is about practical verification for MVP. It is not a full QA program, security review plan, or performance benchmarking document.

## Testing goals

The MVP testing strategy should provide confidence that:

* shell buffer integrity is preserved
* protocol boundaries are validated correctly
* async generation behavior does not allow stale results to overwrite current state
* safety confirmation behavior is predictable
* provider failures degrade cleanly
* core flows work in both supported shells

The testing plan should focus on the product's most failure-prone seams rather than aiming for exhaustive coverage everywhere.

## Non-goals

For MVP, testing does not attempt to provide:

* exhaustive compatibility coverage across all terminal environments
* formal security testing
* full performance benchmarking coverage
* pixel-perfect UI snapshot coverage
* live-provider correctness testing as the main confidence layer

Those areas may matter later, but they are not the primary testing investment for the MVP.

## Test pyramid for MVP

The MVP test strategy should use a simple pyramid:

* many unit tests
* a smaller number of contract and integration tests
* a focused manual QA matrix for real shell behavior

Unit and contract tests should provide most of the confidence. Manual testing should focus on the parts that are hard to simulate faithfully, especially interactive shell behavior.

## Unit test areas

Unit tests should cover pure or mostly pure logic wherever possible.

Priority unit test areas include:

* configuration resolution and defaulting
* protocol validation
* safety classification
* confirmation decision mapping
* provider response parsing
* stale-response suppression logic
* final action preparation
* context normalization, where implemented as deterministic logic

These tests should be fast and numerous because they protect the logic most likely to drift during implementation.

## Contract and schema tests

Contract tests should validate the structured boundaries defined in `protocol.md`.

Priority cases include:

* valid `ShellInvocationRequest`
* invalid `ShellInvocationRequest`
* valid `ShellInvocationResponse`
* invalid `ShellInvocationResponse`
* required versus optional field handling
* unsupported enum handling
* unknown field tolerance
* malformed JSON rejection

These tests are important because they protect the shell adapter and widget runtime boundary. A contract failure here can easily become a user-visible failure or shell integration bug.

## Integration tests

Integration tests should verify behavior across component boundaries, but still avoid unnecessary dependence on live external systems.

### Shell adapter boundary tests

Automate as much of the shell adapter behavior as possible at the subprocess and buffer-application boundary.

Priority cases:

* request creation from shell-local state
* subprocess invocation
* cancel response preserves original buffer
* insert response replaces full buffer
* execute response prepares the shell-native execution path
* malformed widget response is handled safely

Real interactive keybinding behavior does not need to be exhaustively automated in MVP.

### Runtime flow tests

The widget runtime should have integration tests for:

* session initialization
* prompt change to generation flow
* empty results flow
* recoverable provider failure flow
* confirmation-required flow
* cancel flow
* completing and closed-state behavior

These tests should verify the runtime state machine at a behavior level rather than relying only on unit tests of isolated functions.

### Provider integration tests

Provider integration should be tested primarily with mocks or fakes.

Priority cases:

* successful structured response
* timeout
* transport failure
* malformed structured output
* retry once on transient failure
* retry once on malformed structured output
* failure categorization for observability

Live provider calls should not be the main confidence layer for MVP because they introduce noise and nondeterminism.

## Manual test matrix

The MVP manual QA matrix should be concrete and small.

Required manual cases:

* Zsh insert flow
* Zsh execute flow
* Fish insert flow
* Fish execute flow
* cancel from empty prompt
* cancel from seeded prompt
* confirmation flow for a medium or high risk suggestion
* provider timeout visible as recoverable widget error
* malformed provider output visible as recoverable widget error
* multiline prompt seeding
* telemetry disabled configuration

Manual QA should verify the user-facing behavior of real shell sessions, especially places where keybindings, shell redraw, and execution behavior are hard to simulate faithfully in automated tests.

## Failure-path testing

Failure-path tests are mandatory because this product crosses several brittle boundaries.

Priority failure cases:

* widget launch failure
* invalid configuration values
* unsupported schema version
* malformed widget response
* provider timeout
* provider authentication failure
* malformed structured provider output
* safety evaluator failure fallback
* stale async result arriving after a newer request

For each of these, tests should verify:

* the shell buffer is not corrupted
* the runtime remains deterministic
* failure is classified correctly
* confirmation is not silently bypassed

## Release criteria

The following should be considered release-blocking for MVP:

* shell buffer corruption
* stale result overwrite of newer session state
* missing required confirmation for a suggestion that should require it

Before a release candidate is accepted:

* all unit tests must pass
* all contract tests must pass
* shell adapter smoke tests must pass for Fish and Zsh
* the manual QA matrix must be completed
* no known release-blocking bugs may remain open

This keeps the release bar focused on the failures most likely to damage trust in the product.

## Tooling and observability support

Good testing depends on the runtime being observable and mockable.

The implementation should support:

* provider client mocking or faking
* request ID correlation in logs and test output
* deterministic simulation of provider failure cases
* deterministic simulation of stale async responses
* shell adapter tests that do not require live remote inference

Observability should help explain failures without making tests depend on fragile incidental log output.

## Relationship to other docs

This document depends on:

* `shell-integration.md` for shell-facing behavior
* `protocol.md` for contract tests
* `state-machine.md` for runtime flow validation
* `safety-spec.md` for confirmation-related cases
* `provider-integration.md` for provider mock and retry behavior
* `configuration.md` for config resolution tests

Together, these documents define what must be true for the product to behave correctly and what should be validated before release.

## Open questions

The following questions remain open for later refinement:

* how much real shell interaction can be automated reliably in CI
* whether golden tests should be used for structured suggestion payload examples
* whether live-provider smoke tests should be added later as a non-blocking signal
