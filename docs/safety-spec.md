# Safety Spec

## Purpose and scope

This document defines the MVP safety model for `munch`.

It covers:

* how generated commands are classified by risk
* how confirmation requirements are determined
* which signals the safety evaluator uses
* how the system behaves when classification is uncertain or fails

This document is intentionally narrow. For MVP, the goal is to provide a trustworthy confirmation layer without introducing a complex policy engine.

This document does not define:

* shell sandboxing
* command blocking
* deep shell parsing
* provider prompt behavior
* detailed UI state transitions

Those concerns belong elsewhere.

## Safety goals

The MVP safety layer should:

* reduce accidental destructive actions
* keep clearly read-only workflows low-friction
* make confirmation behavior predictable
* prefer conservative behavior when classification is uncertain
* remain simple enough to reason about and implement reliably

The safety system is a confirmation system, not a security boundary.

## Non-goals

For MVP, safety does not attempt to provide:

* guaranteed prevention of dangerous commands
* perfect command understanding
* full shell grammar parsing
* organization-wide policy controls
* per-tool or per-plugin safety models

The purpose is to catch obvious risk and route those actions through confirmation, not to fully understand arbitrary shell semantics.

## Safety model

The MVP safety model has two stages:

1. the suggestion pipeline produces candidate commands
2. the local safety evaluator classifies those commands and determines whether confirmation is required

Two concepts must remain distinct:

* `risk`
* `requires_confirmation`

`risk` is a classification of the command itself.

`requires_confirmation` is the result of applying the user's configured safety policy to that classification.

Provider-supplied risk labels may exist, but they are advisory only. The local safety evaluator is authoritative in MVP.

## Risk levels

MVP uses three risk levels:

* `low`
* `medium`
* `high`

### `low`

`low` risk is for commands that are primarily read-only or inspection-oriented.

Typical examples:

* `ls`
* `pwd`
* `find`
* `rg`
* `fd`
* `git status`
* `cat`
* `bat`

### `medium`

`medium` risk is for commands that mutate local state but do not obviously imply broad destruction or irreversible impact.

Typical examples:

* `mkdir`
* `cp`
* `mv`
* `git add`
* `git commit`
* `touch`
* `sed -i`
* package install commands
* shell redirection using `>` or `>>`

### `high`

`high` risk is for commands that are destructive, broad-impact, privilege-sensitive, or difficult to reverse safely.

Typical examples:

* `rm`
* `rm -r`
* `rm -rf`
* `git reset --hard`
* `git clean`
* force push
* `dd`
* recursive permission or ownership changes
* any command using `sudo`

## Confirmation policy

MVP uses three user-facing safety levels:

* `low`
* `balanced`
* `strict`

### Policy mapping

#### `low`

Require confirmation for:

* `high`

This is the lightest confirmation mode in MVP.

#### `balanced`

Require confirmation for:

* `medium`
* `high`

This is the recommended default behavior for MVP.

#### `strict`

Require confirmation for:

* any mutating command
* all commands classified as `medium`
* all commands classified as `high`

In practice, this means any command the evaluator considers mutating should require confirmation even if it would otherwise be borderline.

## Classification inputs

For MVP, the safety evaluator uses simple inputs:

* the generated command text
* local heuristic pattern matching
* the configured safety level
* any provider-supplied risk hint, as advisory input only

The evaluator does not depend on:

* full shell AST parsing
* provider reasoning as the source of truth
* shell-specific semantic interpretation

This keeps safety predictable and local.

## MVP heuristic rules

The MVP heuristic layer should stay intentionally simple.

### Always classify as `high`

Commands or patterns in this category include:

* `sudo`
* `rm`
* `rm -r`
* `rm -rf`
* `git reset --hard`
* `git clean`
* force push
* `dd`
* recursive `chmod`
* recursive `chown`

Any command using `sudo` is automatically `high` in MVP.

### Usually classify as `medium`

Commands or patterns in this category include:

* `mkdir`
* `cp`
* `mv`
* `git add`
* `git commit`
* `touch`
* `sed -i`
* `brew install`
* `apt install`
* `npm install -g`
* `pip install`
* shell redirection using `>` or `>>`

Package install commands are `medium` in MVP unless combined with `sudo`, in which case they become `high`.

Shell redirection should bump a command to at least `medium` in MVP, because it mutates local state even when the base command might otherwise look read-only.

### Usually classify as `low`

Commands or patterns in this category include:

* `ls`
* `pwd`
* `find`
* `rg`
* `fd`
* `git status`
* `cat`
* `bat`

### Rules intentionally omitted from MVP

The following do not raise risk by themselves in MVP:

* command chaining with `&&`
* command chaining with `||`
* command chaining with `;`

These may matter later, but treating them as risk signals by default is too noisy for MVP.

## Provider labels

If the model returns a risk label or related safety hint, that signal may be consumed as advisory metadata.

However:

* provider labels do not override local classification
* provider labels do not decide confirmation directly
* local evaluator output is authoritative

This keeps safety decisions local, inspectable, and consistent across providers.

## Failure behavior

The safety evaluator must fail conservatively.

Rules:

* if classification fails, require confirmation
* if classification is uncertain, bias upward
* do not fail open
* do not block command selection entirely in MVP

This means the fallback is additional friction, not silent acceptance and not hard rejection.

## User-facing confirmation behavior

Confirmation occurs:

* after the user selects a suggestion
* before the final action is committed

Confirmation should:

* be based on the local evaluator result
* include a short reason
* remain concise

Examples:

* "This command may delete files recursively."
* "This command modifies repository state."
* "This command writes output to a file."

The purpose is to explain why confirmation is required, not to present a long warning dialog.

## Output requirements

For MVP, the safety evaluator should produce enough data for the runtime and UI to behave deterministically.

At minimum, each evaluated suggestion should carry:

* `risk`
* `requires_confirmation`

The system may also retain:

* a short confirmation reason

This maps directly to the protocol model documented in `protocol.md`.

## Open questions

The following questions remain open for later refinement:

* whether shell redirection should distinguish between harmless local files and sensitive system paths
* whether command chaining should eventually influence classification
* whether installs and downloads should remain `medium` long term
* whether some classes of commands should eventually be blocked rather than only confirmed
