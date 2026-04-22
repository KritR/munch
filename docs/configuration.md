# Configuration

## Purpose and scope

This document defines the MVP configuration model for `munch`.

It covers:

* what users can configure
* where configuration comes from
* how effective settings are resolved
* which defaults apply
* how invalid configuration is handled

This document defines the user-facing settings model, not an extensible configuration framework.

## Configuration goals

The MVP configuration model should:

* provide predictable effective settings
* keep setup simple for users
* use one global configuration surface
* centralize resolution through the configuration manager
* fall back safely when config is missing or invalid

Configuration should tune runtime behavior without requiring code changes, but it should not become a source of brittle runtime complexity.

## Non-goals

For MVP, configuration does not attempt to support:

* repo-local overrides
* live reload
* policy layering
* cache tuning controls

Those may be added later, but they are intentionally out of scope for MVP.

## Configuration model

MVP uses one global user configuration object.

Characteristics:

* configuration is stored in the XDG config home
* configuration is represented as a TOML file
* configuration is resolved once at widget session start
* the configuration manager exposes resolved effective settings to the runtime

There is no repo-local configuration in MVP.

The shell sections in this document are limited to simple shell-specific enablement flags for supported shells. MVP does not support a broader per-shell override mechanism for general runtime behavior.

## Configuration sources and precedence

Configuration is resolved from the following sources, in order:

1. built-in defaults
2. global user config file
3. environment variables for provider auth or secrets only

Environment variables are intentionally limited in scope for MVP. They are used for secret lookup, such as API keys, but do not serve as a general-purpose behavior override layer.

This keeps runtime behavior predictable and avoids hidden precedence rules across multiple configuration channels.

## Config file format and location

The MVP config file format is TOML.

The configuration file should live at:

* `$XDG_CONFIG_HOME/munch/config.toml` when `XDG_CONFIG_HOME` is set
* `~/.config/munch/config.toml` otherwise

The configuration manager is responsible for locating, parsing, and validating this file.

## MVP configuration keys

The following keys are in scope for MVP.

### `[safety]`

* `level`

Allowed values:

* `low`
* `balanced`
* `strict`

### `[provider]`

* `base_url`
* `model`
* `api_key_env`
* `timeout_ms`
* `max_retries`

`api_key_env` names the environment variable that contains the provider API key or equivalent secret.

### `[ui]`

* `visible_suggestion_count`

### `[shell.zsh]`

* `enabled`

### `[shell.fish]`

* `enabled`

### `[telemetry]`

* `enabled`

Cache configuration is intentionally omitted in MVP and remains implementation-defined.

## Defaults

The MVP defaults are:

* `safety.level = "balanced"`
* `provider.timeout_ms = 4000`
* `provider.max_retries = 1`
* `ui.visible_suggestion_count = 5`
* `shell.zsh.enabled = true`
* `shell.fish.enabled = true`
* `telemetry.enabled = false`

Defaults should be applied consistently by the configuration manager so all runtime components see the same effective settings.

## Example config

```toml
[safety]
level = "balanced"

[provider]
base_url = "https://example-provider.invalid/v1"
model = "example-model"
api_key_env = "MUNCH_API_KEY"
timeout_ms = 4000
max_retries = 1

[ui]
visible_suggestion_count = 5

[shell.zsh]
enabled = true

[shell.fish]
enabled = true

[telemetry]
enabled = false
```

## Validation behavior

Validation should be conservative and non-destructive.

Rules:

* unknown keys are ignored
* invalid known values fall back to defaults
* validation should emit warnings for invalid known values
* config parsing or validation must not break shell behavior

Examples:

* if `safety.level` is invalid, use the default `balanced`
* if `provider.timeout_ms` is invalid, use the default timeout
* if `ui.visible_suggestion_count` is invalid, use the default count

Missing provider credentials or provider-specific misconfiguration should not crash configuration resolution. Those issues should surface later as provider integration failures if the runtime attempts to use the provider.

## Runtime consumption

Configuration is resolved once at session start.

The configuration manager is responsible for:

* locating the config file
* parsing TOML
* applying defaults
* validating known values
* returning effective settings

Other runtime components should consume effective settings from the configuration manager rather than parsing raw configuration independently.

This keeps behavior consistent and avoids duplicated config logic across the codebase.

## Relationship to other docs

Configuration influences behavior defined in several other documents:

* `safety.level` affects confirmation behavior in `safety-spec.md`
* provider settings affect request behavior in `provider-integration.md`
* UI settings affect visible suggestion behavior in `state-machine.md`
* telemetry settings affect observability described in `architecture.md`

This document defines the settings surface. Other docs define how those settings affect behavior.

## Future extension points

Possible later additions include:

* repo-local configuration
* per-shell overrides
* live reload
* richer telemetry export settings
* cache controls
* feature flags

These are intentionally deferred so the MVP config model stays small and predictable.

## Open questions

The following questions remain open for later refinement:

* whether provider timeout defaults should vary by provider
* whether unsupported future shell sections should be ignored silently or produce warnings
