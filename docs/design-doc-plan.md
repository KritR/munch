# Design Doc Plan

## Purpose

The current `overview.md` is a good product-level design doc. It should stay concise and answer:

* what the product is
* who it is for
* how it behaves from the user's perspective
* which product decisions are already locked in

It should not become the implementation spec for every subsystem. The docs below split out the details that are most likely to create ambiguity, regressions, or rework during implementation.

## What stays in `overview.md`

Keep these sections in `overview.md`:

* overview
* goals and non-goals
* user experience
* interaction model
* high-level context model
* high-level suggestion policy
* high-level safety model
* high-level UI design
* rollout plan

Keep `overview.md` decision-oriented. If a section starts specifying transport contracts, shell hooks, classifier rules, retries, or test matrices, it should move to a focused doc.

## New docs to add

### `architecture.md`

Purpose:
Define the system decomposition and ownership boundaries across the shell adapters, shared widget runtime, suggestion engine, provider client, config, and cache.

Why it is needed:
The current docs describe product behavior, but not where major responsibilities live. That creates avoidable ambiguity around process boundaries and shared-vs-shell-specific logic.

Suggested outline:

* goals of the architecture
* component diagram
* process boundaries
* module responsibilities
* data flow from shell trigger to shell buffer update
* dependency direction rules
* extension points for new shells/providers

### `shell-integration.md`

Purpose:
Specify how Fish and Zsh integrate with the widget and how shell state is handed off and restored.

Why it is needed:
This is one of the highest-risk implementation areas and is only described at the UX level today.

Suggested outline:

* supported shells and MVP assumptions
* Zsh integration via ZLE
* Fish integration via bindings/functions
* how the current buffer is read
* how the selected command replaces the buffer
* accept-and-execute flow
* subprocess launch model
* shell-specific edge cases
* failure behavior and cleanup guarantees

### `protocol.md`

Purpose:
Define the versioned request/response contract between the UI/runtime and the suggestion engine.

Why it is needed:
The current structured response example is useful, but it is not a real contract. This needs explicit validation and compatibility rules.

Suggested outline:

* schema versioning strategy
* request payload
* response payload
* suggestion object fields
* action/result object for insert vs execute
* error payloads
* validation rules
* backward-compatibility policy

### `state-machine.md`

Purpose:
Specify widget states, transitions, and async behavior precisely.

Why it is needed:
The current docs list states, but they do not define legal transitions or stale-response behavior well enough for implementation and testing.

Suggested outline:

* state list
* state transition diagram
* events
* debounce lifecycle
* stale request suppression
* retry flow
* cancellation and close behavior
* invariants that must always hold

### `safety-spec.md`

Purpose:
Define the actual safety system as an enforceable spec rather than a high-level policy description.

Why it is needed:
Safety is central to trust in this product. The existing doc explains intent, but not exact rules, precedence, or edge-case handling.

Suggested outline:

* goals and non-goals of the safety layer
* risk taxonomy
* safety levels and thresholds
* local heuristic patterns
* model-provided vs local risk precedence
* command features that elevate risk
* repo-destructive operations
* filesystem-destructive operations
* network/install operations
* confirmation copy requirements
* known limitations and false-positive tradeoffs

### `privacy-security.md`

Purpose:
Define what local context can leave the machine, what must not, and what redaction or user controls exist.

Why it is needed:
The product sends command history and cwd to a remote model/provider boundary. That deserves a sharper treatment than a short list of sent/not-sent fields.

Suggested outline:

* trust boundaries
* data classification
* data sent to provider
* data retained locally
* redaction strategy
* history handling
* secrets exposure risks
* opt-out and local-only controls
* logging/telemetry constraints

### `configuration.md`

Purpose:
Define the user-facing config model and precedence rules.

Why it is needed:
Safety levels, provider settings, shell behavior, and ranking preferences will otherwise drift into ad hoc flags and undocumented defaults.

Suggested outline:

* config goals
* config file locations
* precedence order
* supported keys
* defaults
* shell-specific overrides
* validation and migration rules

### `provider-integration.md`

Purpose:
Specify the boundary to the LLM provider layer without coupling the product design to a single vendor forever.

Why it is needed:
`overview.md` mentions OpenRouter, but provider behavior, retries, model constraints, and structured-output enforcement deserve their own design.

Suggested outline:

* provider abstraction goals
* prompt construction inputs
* model requirements
* structured-output enforcement
* timeout and retry policy
* fallback behavior
* provider portability constraints

### `caching-performance.md`

Purpose:
Define latency budgets and the cache design needed to meet them.

Why it is needed:
The current docs state performance targets but do not explain how the system will consistently hit them.

Suggested outline:

* latency budget by stage
* cache layers
* cache key composition
* invalidation policy
* interaction with repo/history context
* cold-start vs warm-path expectations
* measurement plan

### `testing-strategy.md`

Purpose:
Define how correctness is verified across schema handling, shell UX, safety rules, and async behavior.

Why it is needed:
This product has a lot of edge-driven behavior. Without an explicit test strategy, regressions will cluster around shell adapters and safety classification.

Suggested outline:

* testing goals
* unit-test scope
* contract/schema tests
* shell integration tests
* safety classifier tests
* async/state-machine tests
* manual QA matrix
* release gating

## Suggested authoring order

### Tier 1: unblock implementation

Write these first:

1. `safety-spec.md`
2. `shell-integration.md`
3. `protocol.md`
4. `architecture.md`
5. `privacy-security.md`

These docs remove the largest sources of product and implementation ambiguity.

### Tier 2: stabilize the runtime

Write these next:

1. `state-machine.md`
2. `provider-integration.md`
3. `configuration.md`
4. `caching-performance.md`

These docs tighten runtime behavior and operability once the core shape is locked in.

### Tier 3: harden delivery

Write this once the first implementation path is clear:

1. `testing-strategy.md`

This doc is still important early, but it depends somewhat on the architecture, protocol, and shell integration decisions.

## Recommended next edits

Short term:

* keep `overview.md` as the entry doc
* add links from `overview.md` to the focused docs once they exist
* treat `outline.md` as working notes unless you want to convert it into one of the new docs

If you want `docs/` to read cleanly for collaborators, the next concrete move should be to draft `safety-spec.md` and `shell-integration.md` first.
