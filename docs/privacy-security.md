# Privacy and Security

## Purpose and scope

This document defines the MVP privacy and security posture for `munch`.

It covers:

* trust boundaries
* local data collection
* which data may be sent to the provider
* which data should not be sent
* secret handling expectations
* logging and telemetry posture

This is a practical MVP document. It is not a formal threat model, compliance document, or sandboxing specification.

## Privacy and security goals

The MVP privacy and security posture should:

* minimize unnecessary data sent to the provider
* make remote data flow explicit
* avoid sending obvious secrets through configuration or telemetry paths
* keep safety and confirmation decisions local
* give users a realistic understanding of the trust boundary

The system should be honest about what it does not protect in MVP.

## Non-goals

For MVP, this document does not attempt to define:

* a local-only mode
* enterprise compliance controls
* filesystem sandboxing
* guaranteed secret detection or redaction
* formal attack-tree or threat-model analysis

Those concerns may matter later, but they are intentionally out of scope for the MVP design.

## Trust boundaries

`munch` spans several trust boundaries:

* the interactive shell process
* the widget process
* local machine context sources
* the remote provider
* local logging and telemetry

The most important trust boundary is the provider boundary.

Before provider use:

* shell-local state is captured locally
* context collection happens locally
* safety evaluation happens locally
* configuration resolution happens locally

At the provider boundary:

* selected prompt and context data may leave the machine

This means the system should assume that any context included in a provider request must be treated as remotely disclosed for practical purposes.

## Data collected locally

The runtime may collect the following local data during a widget session:

* shell type
* original shell buffer
* current prompt text
* current working directory
* recent command history
* installed tool availability
* normalized repo summary
* effective configuration settings

This collection happens locally so the runtime can generate better command suggestions and apply local policy.

## Data sent to the provider

In MVP, the following categories of data may be sent to the provider:

* prompt text
* current working directory
* recent command history
* installed tool availability
* normalized repo summary
* other normalized context needed for suggestion generation

This should be understood as an explicit product behavior, not an implementation detail.

Users should assume that prompt text and included context may be disclosed to the remote provider whenever a suggestion request is made.

## Data not sent to the provider

The following data should not be sent to the provider in MVP:

* file contents
* environment variable values
* provider secret values
* full shell history
* arbitrary filesystem contents
* raw repo diffs
* raw command output

If future features require broader context, that should be documented explicitly rather than assumed under the current design.

## Secret handling

Provider credentials should be supplied through environment variables rather than embedded directly in the config file.

MVP expectations:

* configuration stores the name of the environment variable, not the secret value
* configuration is read from the user's XDG config path, not from the project workspace
* secret values are read locally by the runtime
* secret values are not included in provider payloads
* secret values are not intentionally recorded in logs or telemetry

This does not guarantee that users will never type secrets into prompt text or shell history. It only defines how the application itself should handle configured provider credentials.

## History and prompt sensitivity

Prompt text and recent command history may contain sensitive information.

Examples include:

* file paths that reveal project names or usernames
* copied commands containing tokens or hostnames
* repository context that reveals internal branch or environment names

MVP does not guarantee robust secret redaction for prompt text, history, cwd, or other context before provider calls.

That means:

* users should treat the included prompt and context as potentially visible to the provider
* the product should not imply stronger protection than it actually implements

Future versions may add redaction or context-minimization features, but MVP should be explicit about the lack of robust automatic redaction.

## Logging and telemetry posture

Logging and telemetry are first-class for debuggability, but they must remain privacy-conscious.

MVP posture:

* telemetry is local-first
* telemetry is disabled by default
* logs and telemetry should avoid storing raw secret values
* logs and telemetry should avoid storing raw prompt text or recent history by default

Operational data that is appropriate to capture includes:

* request identifiers
* provider latency
* retry counts
* failure categories
* parse success or failure

The system should prefer structured operational signals over raw user content in logs.

## User controls

The MVP user-facing privacy and observability controls are intentionally limited.

Available controls:

* `telemetry.enabled`
* provider credential indirection through `provider.api_key_env`

Not available in MVP:

* per-context sharing toggles
* history opt-out toggles
* cwd opt-out toggles
* local-only inference mode

If those controls are added later, they should be documented explicitly rather than implied by the current MVP design.

## Failure and recovery considerations

Failure handling should not expand data exposure unnecessarily.

Rules:

* provider failures should not dump raw provider payloads into the UI
* recoverable user-facing errors should stay concise and sanitized
* debug logging should still avoid raw secret capture
* malformed provider output should be handled without exposing unrelated local context

This keeps operational debugging possible without turning failures into secondary privacy leaks.

## Relationship to other docs

This document depends on:

* `architecture.md` for trust boundaries and component ownership
* `provider-integration.md` for provider request and response handling
* `configuration.md` for credential and telemetry settings
* `safety-spec.md` for the decision to keep confirmation behavior local

Together, these documents define what data is collected, what may leave the machine, and what remains local in MVP.

## Open questions

The following questions remain open for later refinement:

* whether users should be able to disable specific context categories such as history or cwd
* whether the runtime should add heuristic redaction for obviously sensitive prompt or history content
* whether a local-only provider mode should be supported later
* whether telemetry export destinations beyond local-first behavior should be introduced
