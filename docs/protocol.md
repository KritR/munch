# Protocol

## Purpose and scope

This document defines the structured message contract for `munch`.

It covers two related layers:

* the shell integration protocol between the shell adapter and the widget runtime
* the shared domain payloads used inside the widget runtime and suggestion pipeline

These layers are documented together because the shell-facing contract and the core payload shapes need to stay aligned. The shell adapter should remain thin, and the runtime should operate on stable, explicit objects rather than ad hoc data structures.

This document does not define:

* shell behavior or shell-specific implementation details
* widget state transitions
* provider-specific API contracts
* safety policy rules

Those belong in `shell-integration.md`, `state-machine.md`, `provider-integration.md`, and `safety-spec.md`.

## Protocol goals

The protocol should satisfy the following goals:

* remain simple enough for thin shell adapters
* be explicit and versioned
* stay logically transport-agnostic even though MVP uses bridge-mode transport
* be straightforward to validate
* support additive forward-compatible change
* keep the shell response surface minimal

The protocol is designed for MVP simplicity, but it should not block a later move to a long-lived local backend.

## Non-goals

This document does not attempt to define:

* provider API payloads
* prompt wording
* widget UI rendering
* detailed runtime state transitions
* safety heuristics or confirmation policy logic

## Protocol boundaries

The protocol has two main boundaries.

### Shell integration boundary

The shell integration boundary is the contract between:

* the shell adapter
* the widget runtime

In MVP, this boundary is implemented through shell-specific bridge modes rather than direct shell-side JSON parsing. The logical request and response objects documented here are still the source of truth, but the shell scripts communicate with the widget through:

* environment variables into the bridge
* shell-safe assignments out of the bridge

### Shared domain payload boundary

The shared domain payload boundary covers the structured objects that move between major runtime components inside the widget process, such as suggestions and safety-enriched suggestion metadata.

These payloads are documented here so that:

* the runtime, suggestion engine, and safety evaluator share the same concepts
* the shell response shape remains consistent with the runtime's internal models
* later refactors do not silently drift the meaning of core objects

## Transport model

In MVP, the shell integration protocol is adapted through shell bridge modes.

The widget still operates on the same logical message shapes defined in this document, but the shell-facing transport is:

* shell-local state passed in via environment variables
* bridge-generated shell-safe assignments returned on `stdout`

Direct JSON request/response exchange still exists inside the widget and remains useful for non-bridge session mode and tests, but shell adapters do not parse JSON directly.

The logical message shapes defined in this document should remain reusable if the implementation later moves to:

* a long-lived local backend
* a socket-based transport
* another RPC mechanism

## Versioning and compatibility

All protocol messages must include `schema_version`.

Protocol rules:

* `schema_version` is an integer
* MVP starts at `1`
* unknown fields are ignored
* additive fields are allowed within a schema version as long as required field semantics do not change
* changing the meaning of a required field requires a schema version bump
* adding a new action value requires a schema version bump

If a shell adapter or widget runtime encounters an unsupported `schema_version`, the invocation should fail cleanly rather than attempting partial interpretation.

## Top-level message families

The main message and payload families are:

* `ShellInvocationRequest`
* `ShellInvocationResponse`
* `Suggestion`
* `SafetyAssessment`
* `FinalAction`

`FinalAction` is documented as a conceptual object even though, in MVP, it is represented directly inside `ShellInvocationResponse`.

## Shell invocation request

`ShellInvocationRequest` is the message sent from the shell adapter to the widget runtime when a widget session starts.

### Fields

* `schema_version`: integer, required
* `request_id`: string, required
* `shell`: string enum, required
* `original_buffer`: string, required
* `prompt_text`: string, required
* `cursor_position`: integer, required

### Field semantics

#### `schema_version`

The protocol schema version used to interpret the payload.

#### `request_id`

A unique request correlation identifier for this invocation. In MVP, one widget invocation maps to one request and one final response, so a separate `session_id` is not required.

#### `shell`

The invoking shell.

Allowed values in MVP:

* `zsh`
* `fish`

Unsupported values should be rejected by the widget runtime.

#### `original_buffer`

The exact shell buffer contents captured at widget open. This is the shell state that must be restored on cancel or hard failure.

#### `prompt_text`

The editable text shown in the widget at session start. In MVP, it is initialized to the same value as `original_buffer`, but it remains conceptually distinct.

#### `cursor_position`

The cursor offset in the original shell buffer at widget open. In MVP, this field exists for forward compatibility and does not drive cursor-aware editing behavior.

### Validation rules

`ShellInvocationRequest` is valid only if:

* all required fields are present
* `schema_version` is supported
* `request_id` is non-empty
* `shell` is one of the supported enum values
* `cursor_position` is greater than or equal to `0`
* `original_buffer` may be empty
* `prompt_text` may be empty

## Shell invocation response

`ShellInvocationResponse` is the final message returned from the widget runtime to the shell adapter.

### Fields

* `schema_version`: integer, required
* `request_id`: string, required
* `action`: string enum, required
* `command`: string, conditionally required
* `selected_suggestion_index`: integer, optional

### Field semantics

#### `action`

Allowed values in MVP:

* `cancel`
* `insert`
* `execute`

There is no `error` action in MVP.

Recoverable errors should be handled while the widget is still active and shown inside the widget UI. Hard failures are out-of-band and are represented by process failure, malformed output, or unsupported schema version.

#### `command`

The command string to apply to the shell buffer.

Required when:

* `action = insert`
* `action = execute`

Must be omitted or ignored when:

* `action = cancel`

#### `selected_suggestion_index`

An optional zero-based index of the suggestion chosen by the user. This is diagnostic metadata only and does not affect shell behavior.

### Validation rules

`ShellInvocationResponse` is valid only if:

* all required fields are present
* `schema_version` is supported
* `request_id` is non-empty
* `action` is one of the supported enum values
* `command` is present and non-empty when `action` is `insert` or `execute`
* `command` is omitted or ignored when `action` is `cancel`
* `selected_suggestion_index`, if present, is greater than or equal to `0`

## Final action model

`FinalAction` is the conceptual result returned by the widget runtime after the user finishes the session.

In MVP, `FinalAction` is flattened into `ShellInvocationResponse`.

Possible final actions:

* `cancel`
* `insert`
* `execute`

This model exists to keep the action concept distinct from the transport-specific response object. If later transports or daemonized flows require richer response envelopes, the conceptual `FinalAction` should remain stable.

## Suggestion payload model

`Suggestion` is the canonical command suggestion object used inside the widget runtime after generation and safety enrichment.

### Fields

* `command`: string, required
* `description`: string, required
* `risk`: string enum, required
* `requires_confirmation`: boolean, required
* `assumptions`: array of strings, optional
* `uses_tools`: array of strings, optional
* `confidence`: number, optional

### Field semantics

#### `command`

The command text presented to the user and eligible for insertion or execution.

#### `description`

A short human-readable explanation of what the command does.

#### `risk`

The risk classification assigned to the suggestion.

Allowed values in MVP:

* `low`
* `medium`
* `high`

#### `requires_confirmation`

The flattened policy decision indicating whether the current safety policy requires explicit confirmation before the suggestion can be committed.

This field is explicit rather than inferred at the UI boundary so the widget runtime does not need to recompute the same policy result after safety evaluation.

#### `assumptions`

Optional list of assumptions associated with the suggestion, such as tool availability or interpretation assumptions.

#### `uses_tools`

Optional list of tools referenced by the command, such as `rg`, `fd`, or `git`.

#### `confidence`

Optional numeric confidence value if the generation pipeline chooses to provide one. MVP does not require a specific scale beyond consistent local interpretation.

### Validation rules

`Suggestion` is valid only if:

* `command` is present and non-empty
* `description` is present and non-empty
* `risk` is one of the supported enum values
* `requires_confirmation` is present
* `assumptions`, if present, is an array of strings
* `uses_tools`, if present, is an array of strings
* `confidence`, if present, is numeric

## Safety assessment model

`SafetyAssessment` is the conceptual result of evaluating a suggestion against safety policy.

This object is documented separately even though MVP flattens part of its result onto `Suggestion`.

### Fields

* `risk`: string enum, required
* `requires_confirmation`: boolean, required
* `reason`: string, optional

### Semantics

`SafetyAssessment` distinguishes between:

* classification, represented by `risk`
* policy outcome, represented by `requires_confirmation`
* explanation, represented by `reason`

This keeps safety classification conceptually separate from the final `Suggestion` object even if the runtime stores a flattened version for convenience.

## Error model

The protocol distinguishes between three categories of failure.

### In-widget recoverable errors

These are errors the widget can surface while remaining active, such as:

* provider request failure
* invalid model output that can be retried
* empty result set

These are not shell response actions. They belong to widget UI state and should be surfaced in the widget header or equivalent status area.

### Out-of-band protocol or process failures

These are failures where the shell adapter does not receive a valid terminal response, such as:

* widget process exits non-zero
* widget process times out
* widget returns malformed JSON
* widget returns an unsupported schema version

In these cases, the shell adapter should restore the original shell buffer and treat the invocation as failed.

### Validation failures

Validation failures occur when a payload is structurally present but invalid according to this protocol. They should be handled as hard failures at the boundary where validation occurs.

## Validation rules

Across all protocol messages and payloads:

* required fields must be present
* unsupported enum values are invalid
* unknown fields are ignored
* malformed JSON is invalid
* type mismatches are invalid

Coupled field rules:

* `command` is required for `insert` and `execute`
* `command` is omitted or ignored for `cancel`
* `requires_confirmation` must be explicit on `Suggestion`

Validation should occur at every boundary where payloads cross component or process lines, rather than being deferred until later runtime use.

## Compatibility rules

The protocol supports additive evolution under these rules:

* new optional fields may be added without a version bump if they do not change existing field semantics
* unknown fields must be ignored
* required fields must keep the same meaning within a schema version
* new enum values require a schema version bump
* changing validation rules for existing required fields requires a schema version bump

Because the shell adapter and widget runtime are expected to ship together in MVP, compatibility pressure is relatively low. Even so, explicit rules help prevent accidental drift.

## Examples

### Example `ShellInvocationRequest`

```json
{
  "schema_version": 1,
  "request_id": "req_01",
  "shell": "zsh",
  "original_buffer": "find all log files modified today",
  "prompt_text": "find all log files modified today",
  "cursor_position": 33
}
```

### Example `ShellInvocationResponse` for `cancel`

```json
{
  "schema_version": 1,
  "request_id": "req_01",
  "action": "cancel"
}
```

### Example `ShellInvocationResponse` for `insert`

```json
{
  "schema_version": 1,
  "request_id": "req_01",
  "action": "insert",
  "command": "find . -type f -name '*.log' -mtime -1",
  "selected_suggestion_index": 0
}
```

### Example `Suggestion`

```json
{
  "command": "rg -n TODO .",
  "description": "Search recursively for TODO comments using ripgrep",
  "risk": "low",
  "requires_confirmation": false,
  "assumptions": [
    "rg is installed"
  ],
  "uses_tools": [
    "rg"
  ],
  "confidence": 0.91
}
```

### Example `SafetyAssessment`

```json
{
  "risk": "high",
  "requires_confirmation": true,
  "reason": "Command may delete files recursively"
}
```

## Open questions

The following questions remain open for later refinement:

* whether timestamps are needed in shell-facing protocol messages
* whether `selected_suggestion_index` is worth keeping long term
* whether suggestions should later gain stable IDs within a session
* whether additional internal payload families should be documented here as implementation becomes more concrete
