# Prompting

## Purpose and scope

This document defines the MVP prompting design for `munch`.

It covers:

* the goals of provider-facing prompt construction
* which runtime inputs should be included in prompts
* what the system prompt is responsible for
* which decisions remain local and must not be delegated to the model
* the canonical MVP system prompt text

This document is about prompt design, not provider transport details or runtime state transitions.

## Prompting goals

The MVP prompting layer should:

* generate relevant shell commands from user intent
* use local context to improve relevance
* prefer installed tools
* prefer modern tools when appropriate
* prefer read-only commands unless the user's intent implies mutation
* return machine-parseable structured output reliably
* avoid conversational drift

The prompt should guide the model toward useful command suggestions without turning prompt text into the source of product policy.

## Non-goals

For MVP, prompting does not attempt to provide:

* authoritative safety enforcement
* final confirmation decisions
* general-purpose chat behavior
* provider-specific optimization tricks as the main design artifact
* exhaustive shell-semantic understanding through prompt wording alone

Local runtime components remain responsible for validation, policy, and final action handling.

## Prompting boundary

Prompt design owns:

* instructing the model to generate shell command suggestions
* expressing tool and context preferences
* defining the expected structured output shape
* requesting advisory metadata such as risk hints

Prompt design does not own:

* final confirmation policy
* authoritative safety classification
* domain validation
* reranking after local policy application
* shell buffer behavior

This boundary is important because `munch` treats the provider as a suggestion source, not the source of truth for product policy.

## Prompt inputs

The prompt should be built from curated runtime inputs rather than by dumping raw internal objects indiscriminately.

MVP prompt inputs should include:

* the current task prompt text
* current working directory
* recent command history
* installed tool availability
* normalized repo summary
* desired suggestion count
* the expected output shape

The context should be rendered in a curated, stable, human-readable structure.

MVP should not rely on passing raw large JSON blobs as the primary prompt format. Curated rendering keeps the prompt easier to inspect, easier to evolve, and less likely to include accidental noise.

## System prompt responsibilities

The system prompt should explicitly instruct the model to:

* generate shell commands rather than conversational answers
* return useful alternatives when appropriate
* prefer installed tools over unavailable ones
* prefer modern tools when available
* prefer read-only commands unless the task clearly implies mutation
* include concise descriptions and assumptions when useful
* return only the specified machine-parseable JSON object

The system prompt may also request an advisory `risk` field, but it must not assign final confirmation policy.

## Output requirements

The prompt should require a canonical structured output shape aligned with `protocol.md`.

For MVP:

* the model should return a top-level JSON object
* the object should contain a `suggestions` array
* each suggestion should include:
  * `command`
  * `description`
  * `risk`
  * `assumptions`
  * `uses_tools`
  * optional `confidence`

The prompt should not ask the model to produce `requires_confirmation`.

That field is derived locally after safety evaluation and policy application.

The prompt should also constrain verbosity:

* return only valid JSON
* do not emit preambles or trailing explanation
* keep descriptions short
* keep assumptions short and only include them when useful
* return up to the requested suggestion count rather than padding with low-value alternatives

## Safety and authority boundaries

The prompt may ask the model for advisory risk classification, but only as a hint.

Authority remains local:

* local safety evaluation is authoritative
* provider risk labels are advisory only
* final confirmation policy is local
* the model must not decide whether a suggestion is allowed to execute

The prompt may encourage cautious suggestions, but it must not imply that provider output replaces local validation or policy.

## Prompt construction strategy

MVP prompt construction should be organized into stable sections:

1. system instruction block
2. task block
3. environment/context block
4. output schema block

This keeps the prompt inspectable and stable across providers while still allowing the provider client to adapt formatting as needed.

The suggestion engine should construct the prompt intent. The provider client may adapt the final wrapping to fit a provider's request format, but it should not change the underlying product meaning.

## Canonical system prompt

The following is the canonical MVP system prompt text.

```text
You generate shell command suggestions for an interactive shell widget.

Your job is to return a small set of useful shell commands that help the user accomplish the requested task in their current environment.

Follow these rules:

1. Return only a valid JSON object matching the requested output shape.
2. Do not include any text before or after the JSON object.
3. Generate shell commands, not conversational answers.
4. Prefer commands that use tools known to be installed.
5. Prefer modern tools when they are available and appropriate.
6. Prefer read-only commands unless the user's task clearly implies mutation.
7. When useful, include a fallback command that uses more standard tooling.
8. Keep descriptions short and practical.
9. Keep assumptions short and include them only when they materially help the user understand the command.
10. Include an advisory risk classification for each suggestion using only: low, medium, high.
11. Do not decide final confirmation policy.
12. Do not output extra explanation, markdown, or prose outside the JSON object.

Output a JSON object with this shape:
{
  "suggestions": [
    {
      "command": "string",
      "description": "string",
      "risk": "low|medium|high",
      "assumptions": ["string"],
      "uses_tools": ["string"],
      "confidence": 0.0
    }
  ]
}

Return up to the requested number of suggestions. Prefer a smaller number of high-quality suggestions over padding with weak ones.
```

## Prompt assembly template

The full provider-facing prompt should be assembled from the canonical system prompt plus a structured task and context rendering.

A representative assembly shape is:

```text
System prompt:
<canonical system prompt>

Task:
<prompt_text>

Context:
- cwd: <cwd>
- history:
  - <recent command 1>
  - <recent command 2>
- installed tools:
  - rg: true
  - fd: true
  - jq: false
- repo summary:
  - type: git
  - branch: main
  - dirty: true

Requested suggestion count:
<N>
```

This is an assembly template, not a provider transport format. The provider client may wrap this differently as long as the same semantic content is preserved.

## Provider-specific adaptation

The canonical prompt intent should remain provider-neutral.

The provider client may adapt:

* message wrapping
* exact field placement
* structured-output enforcement mechanism

The provider client should not change:

* the output contract
* the authority boundary between prompt and local safety
* the core tool-preference and read-only preference rules

## Prompt evaluation and iteration

Prompt changes should be evaluated against representative tasks and judged on:

* schema adherence
* command relevance
* installed-tool preference quality
* fallback usefulness
* advisory risk usefulness
* tendency to avoid conversational drift

Prompt iteration should be treated as product behavior work, not just model-tuning trivia. Small wording changes can affect output quality and reliability materially.

## Relationship to other docs

This document depends on:

* `provider-integration.md` for the provider boundary
* `protocol.md` for the canonical payload shapes
* `safety-spec.md` for local confirmation authority
* `privacy-security.md` for context sensitivity and provider disclosure posture

Together, these documents define what the model is asked to do, what data it can see, and which decisions remain local.

## Open questions

The following questions remain open for later refinement:

* whether few-shot examples are needed for schema reliability
* how much context should be trimmed when prompts become large
* whether different providers will require distinct but semantically equivalent prompt wrappers
