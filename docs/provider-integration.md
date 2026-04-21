# Provider Integration

## Purpose and scope

This document defines the integration boundary between the `munch` runtime and the remote inference provider.

It covers:

* the role of the provider client
* the contract between the suggestion engine and provider client
* structured output expectations
* retry and timeout behavior
* failure categories and handling
* portability requirements for future provider changes

This document does not define:

* shell behavior
* widget state transitions
* detailed safety policy
* exact prompt wording
* provider configuration values

Those concerns belong in `shell-integration.md`, `state-machine.md`, `safety-spec.md`, and `configuration.md`.

## Provider integration goals

The provider integration layer should:

* isolate provider-specific APIs behind a dedicated client
* keep the runtime-to-provider boundary narrow
* return machine-parseable structured output
* make provider failures observable and classifiable
* support future provider replacement without forcing a redesign of the runtime
* stay simple enough for MVP

The provider layer exists to translate between runtime intent and provider-specific mechanics, not to own product behavior.

## Non-goals

For MVP, provider integration does not attempt to provide:

* streaming suggestions
* multi-provider routing
* cost-aware provider selection
* complex fallback trees
* provider-specific UI behavior

MVP should optimize for predictable structured output and simple failure handling.

## Provider boundary

Only the provider client may communicate with the remote inference backend.

The interaction model is:

* the suggestion engine produces a provider-independent generation request
* the provider client converts that request into a provider-specific API call
* the provider client receives a provider-specific response
* the provider client converts that response into a provider-independent structured payload
* the suggestion engine validates that payload against the domain model

This keeps provider-specific request formats, response formats, and enforcement mechanisms inside one layer.

## Generation request model

The suggestion engine should pass a structured generation request object to the provider client.

In MVP, this request should conceptually include:

* `prompt_text`
* normalized context
* desired suggestion count
* expected structured output shape

The provider client is responsible for translating that request into the provider's request format, such as messages, prompt text, or any provider-specific structured-output configuration.

The provider client should not be responsible for business-level prompt design. It is an adapter layer, not the owner of generation policy.

## Structured output contract

MVP standardizes on one canonical structured-output mechanism:

* the provider must return a machine-parseable JSON object

The provider client is responsible for using whatever provider-specific mechanism best enforces or strongly encourages that output format.

The rest of the runtime depends only on the canonical structured response shape and not on the provider-specific enforcement mechanism.

This document does not restate the full domain object schemas. `protocol.md` is the source of truth for payload shapes such as:

* `Suggestion`
* `SafetyAssessment`

The provider boundary should produce data that can be mapped into those domain objects cleanly.

## Provider response handling

The provider client returns a provider-independent structured payload to the suggestion engine.

The suggestion engine remains responsible for:

* domain validation
* local reranking or filtering
* safety enrichment

This keeps transport concerns and business validation separate.

Provider-supplied risk labels or related metadata may be preserved as advisory input, but they do not override local safety evaluation.

Malformed structured output is treated as a recoverable generation failure and triggers one retry in MVP.

## Retry and timeout behavior

MVP retry behavior is intentionally narrow.

Rules:

* retry once on transient transport or provider failures
* retry once on malformed structured output
* do not implement broader retry policy in MVP

Timeout behavior should remain explicit but simple:

* provider requests must have a finite timeout
* timeout values should be chosen to preserve an interactive user experience
* timeout specifics can be finalized in implementation or configuration docs

The retry model is intentionally small so it improves resilience without making behavior unpredictable or inflating latency excessively.

## Failure categories and handling

The provider client should classify failures into a small set of operational categories.

MVP categories should include:

* timeout
* transport or network failure
* provider service failure
* authentication or configuration failure
* malformed structured output

Handling rules:

* failures should surface to the widget as recoverable generation failures when the widget remains active
* user-facing messages should remain concise and sanitized
* raw provider-specific failure detail should not be surfaced directly to the user by default
* logging and telemetry should preserve a more specific failure category for debugging

This keeps the UI simple while still preserving debuggability.

## Provider abstraction and portability

The provider client boundary must remain provider-neutral from the rest of the runtime's perspective.

Requirements:

* no provider-specific request types outside the provider client
* no provider-specific response types outside the provider client
* no vendor-specific structured-output mechanics leaked into the suggestion engine
* swapping providers should not require changing the runtime's core suggestion flow

This is especially important because the architecture assumes the provider layer is replaceable even if MVP only supports one provider path initially.

## Observability

Provider integration should emit enough operational information to debug user-facing issues and provider-side problems.

At minimum, logging and telemetry should capture:

* `request_id`
* provider or model identifier
* request latency
* retry count
* failure category
* structured parse success or failure

Observability is not optional here. The provider boundary is one of the least deterministic parts of the system, so it must be inspectable out-of-band.

## Relationship to other docs

This document depends on:

* `architecture.md` for component ownership
* `protocol.md` for domain payload shapes
* `state-machine.md` for recoverable failure behavior in the widget runtime
* `safety-spec.md` for local risk authority and confirmation behavior

Together, these docs define the full runtime path from prompt input to provider request to validated suggestions to final action.

## Open questions

The following questions remain open for later refinement:

* exact timeout values
* whether retries should later differentiate more finely by failure type
* whether streaming should ever change the provider client contract
* whether the canonical structured response shape should expand in future versions
