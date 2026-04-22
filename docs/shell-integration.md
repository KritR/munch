# Shell Integration

## Purpose and scope

This document defines how `munch` integrates with interactive shells in MVP, specifically Zsh and Fish.

The integration layer is responsible for taking the user's current shell state, launching the widget, and applying the widget's result back into the shell safely. The main goal is to make the widget feel like a native shell affordance rather than a separate application that happens to manipulate shell text.

This document covers:

* how the widget is invoked from the shell
* which shell state is captured at invocation time
* how the shell adapter and widget process communicate
* how returned actions are applied to the shell buffer
* what guarantees hold on cancel, failure, and execute
* which shell-specific responsibilities belong in adapters versus the shared runtime

This document does not define:

* suggestion ranking
* provider or model behavior
* internal widget UI design beyond shell-facing consequences
* long-lived daemon architecture beyond compatibility requirements

## Integration goals

The shell integration layer should:

* feel native in interactive shell workflows
* keep shell-specific code thin and predictable
* delegate shared behavior to a common widget/runtime
* preserve the original shell buffer unless a valid insert or execute action is returned
* support visible prompt seeding from the current shell buffer
* use a simple transport in MVP while keeping the logical contract portable to a future long-lived backend

These goals bias the design toward a narrow adapter layer in Fish and Zsh and a shared runtime that owns nearly all product behavior.

## Non-goals

For MVP, shell integration does not attempt to support:

* Bash
* non-interactive shells
* partial-buffer transforms
* shell-native rendering of the suggestion interface
* a persistent background process
* advanced multiline editing semantics

MVP should optimize for correctness, simplicity, and reversibility rather than for maximum flexibility.

## Shared integration model

Both Zsh and Fish follow the same high-level lifecycle:

1. The user invokes the widget through a shell keybinding.
2. The shell adapter captures shell-local state.
3. The shell adapter launches a fresh widget process.
4. The shell adapter launches the widget in a shell-specific bridge mode.
5. Shell-local state is passed into the bridge through environment variables.
6. The widget runs its UI and internal suggestion flow.
7. The bridge returns shell-safe assignments to the shell adapter.
8. The shell adapter interprets the returned action.
9. The shell adapter either restores the original buffer, replaces the full buffer, or replaces the full buffer and executes it.

This model is intentionally simple for MVP. It makes failure handling easier, keeps shell adapters small, and keeps JSON handling out of shell code. The contract between shell adapter and widget should still be treated as a stable logical interface even if the physical transport changes later.

## Responsibility split

### Shell adapter responsibilities

The shell adapter is responsible for:

* binding the shell key
* reading shell-native buffer state
* launching the widget process
* sending the request payload
* receiving the response payload
* applying the resulting action to the shell buffer
* redrawing or refreshing the prompt when necessary

The shell adapter should not contain product logic. It should behave like a translation layer between shell APIs and the shared widget contract.

### Shared widget/runtime responsibilities

The shared widget/runtime is responsible for:

* rendering the widget UI
* initializing visible prompt text
* collecting non-shell-specific context
* generating suggestions
* handling selection and confirmation
* surfacing recoverable errors in the widget UI
* returning a final action payload

The shell adapter should remain thin. It should not implement suggestion logic, safety classification, provider-specific behavior, or internal UI state handling.

## Captured shell state

At widget open, the shell adapter captures:

* `shell`: shell identifier such as `zsh` or `fish`
* `original_buffer`: exact shell buffer contents at invocation time
* `cursor_position`: current cursor offset in the shell buffer

For MVP, `cursor_position` is included for forward compatibility only. The integration does not use it to implement cursor-aware editing behavior.

The widget/runtime may collect additional shared context after launch, including:

* current working directory
* recent shell history
* active configuration
* repo summary
* installed tools

This split keeps shell adapters focused on shell-native state while centralizing shared context collection in one place.

## Prompt seeding behavior

The widget prompt is seeded visibly from the current shell buffer.

The request model distinguishes between:

* `original_buffer`: the exact shell contents captured when the widget opened
* `prompt_text`: the editable text shown inside the widget

In MVP:

* `prompt_text` is initialized from `original_buffer`
* both fields are equal at invocation time
* they may diverge once the user edits the prompt inside the widget

This distinction matters even in MVP. The original buffer is the shell state that must be restored on cancel, while prompt text is the editable user input inside the widget. Treating them separately now avoids conflating shell restoration semantics with widget editing semantics later.

If the original shell buffer is empty, the widget prompt starts empty.

If the original shell buffer is multiline, it is passed through as-is and displayed as plain text seed input.

## Transport and protocol

MVP uses shell bridge modes rather than direct shell-side JSON transport.

Requirements:

* the shell adapter launches the widget with a shell-specific bridge mode such as `zsh-bridge` or `fish-bridge`
* shell-local state is passed via environment variables
* the bridge returns shell-safe assignments
* malformed bridge output is treated as failure
* non-zero exit status is treated as failure

Inside the widget process, the logical request and response shapes still correspond to the protocol objects documented in `protocol.md`. The shell bridge is an adaptation layer around those objects rather than a separate product model.

The protocol should still be treated as transport-agnostic. A future daemon or RPC layer should be able to reuse the same logical message shape even if the physical transport changes.

## Request model

The shell integration request should include, at minimum:

* `schema_version`
* `shell`
* `original_buffer`
* `prompt_text`
* `cursor_position`

In MVP, `prompt_text` is initialized to the same value as `original_buffer`.

Additional context such as cwd, history, tools, repo state, and config may be collected by the widget/runtime after launch rather than by the shell adapter itself. This keeps the shell-side implementation small and avoids duplicating shared context logic across shells.

## Response model

The shell integration response should return one final action.

MVP actions:

* `cancel`
* `insert`
* `execute`

Action semantics:

* `cancel`: restore the original shell buffer exactly
* `insert`: replace the entire shell buffer with the returned command and place the cursor at end-of-line
* `execute`: replace the entire shell buffer with the returned command and trigger the shell-native execution path

Recoverable problems should be surfaced while the widget is still running, as part of the widget UI, not as a separate response action. In particular, transient request failures, parsing issues, or empty-result cases should appear in the widget header or equivalent status area if the widget is still alive and interactive.

Hard failures are out-of-band. If the widget exits non-zero, times out, or returns malformed bridge output, the shell adapter treats that as process failure rather than as a normal protocol result.

## Buffer replacement and execution semantics

The accepted command always replaces the full shell buffer in MVP.

MVP does not support:

* partial insertion
* cursor-relative transforms
* preserving only selected fragments of the original buffer

Cursor behavior:

* after `insert`, place the cursor at the end of the inserted command
* after `execute`, execution is delegated to the shell-native accept or execute mechanism
* multiline commands follow the same rule: the cursor lands at the end of the final buffer before execution or after insertion

This behavior is intentionally blunt. It is easier to reason about, easier to test, and consistent with the current product model in which the widget returns a replacement command rather than a text patch.

## Zsh integration

Zsh integration should use a custom ZLE widget.

The Zsh adapter is expected to:

* capture `BUFFER`
* capture `CURSOR`
* launch the widget subprocess
* send the integration request
* parse the response
* update `BUFFER`
* update `CURSOR` as needed
* call `zle accept-line` only for the `execute` action

Behavioral guarantees:

* if the widget is canceled, `BUFFER` is restored exactly
* if the widget fails, `BUFFER` is restored exactly
* no shell buffer mutation occurs unless a valid `insert` or `execute` action is returned

The adapter should rely on normal ZLE mechanisms for prompt redraw and execution. It should not attempt to simulate execution in a shell-specific custom way when `zle accept-line` already provides the intended behavior.

## Fish integration

Fish integration should use a custom binding or function.

The Fish adapter is expected to:

* capture the current commandline buffer
* capture the current cursor position
* launch the widget subprocess
* send the integration request
* parse the response
* replace the commandline buffer when required
* trigger Fish-native execution only for the `execute` action

Behavioral guarantees:

* if the widget is canceled, the original commandline buffer is restored exactly
* if the widget fails, the original commandline buffer is restored exactly
* no shell buffer mutation occurs unless a valid `insert` or `execute` action is returned

As in Zsh, the adapter should use native Fish mechanisms for final command execution rather than inventing separate execution logic in the adapter layer.

## Failure handling and restoration guarantees

The integration must handle these cases safely:

* widget launch failure
* widget timeout
* malformed JSON response
* non-zero process exit
* interrupted subprocess
* explicit user cancel

In all failure cases:

* preserve the original shell buffer
* do not partially apply shell changes
* leave the shell in a usable state

The default failure mode should be non-destructive.

Error presentation follows two paths:

* recoverable, in-widget issues should be shown in the widget header while the widget remains active
* unrecoverable process or protocol failures should result in silent buffer restoration and return to the shell, with no mutation applied

This split keeps shell adapters simple and avoids forcing them to interpret product-level error states.

## Multiline and edge-case handling

MVP uses the simplest correct behavior for multiline input:

* multiline shell buffers are captured verbatim
* multiline prompt seed text is displayed verbatim
* accepted output still replaces the full buffer
* cursor placement after insert is always end-of-buffer

Other edge cases:

* syntactically invalid or partially typed shell input is still treated as plain seed text
* unclosed quotes or incomplete pipelines do not block widget launch
* no attempt is made to preserve fine-grained cursor intent within multiline or invalid shell input

This approach favors predictable behavior over shell-aware editing sophistication. The widget acts on the current shell line as user text, not as a partially parsed command structure.

## Future-compatible extension points

This integration design should not block later additions such as:

* a long-lived local backend service
* alternative transports beyond one-shot `stdin` and `stdout`
* richer request and response protocols
* partial-buffer transforms
* Bash support

These are explicitly out of scope for MVP, but the shell integration contract should avoid closing them off unnecessarily. The main way it does that is by keeping the shell adapter narrow and making the request and response schema the stable boundary rather than the process model itself.

## Open questions

The following details should be finalized in the protocol and implementation docs:

* exact JSON schema fields
* timeout values
* shell-specific history access details
* whether any shell-visible fallback is needed for unrecoverable launch failures
