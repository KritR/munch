## What’s locked in

* **Shells first:** Fish and Zsh
* **UI form:** transient bottom panel
* **Suggestions:** multiple visible suggestions, current shape is good
* **History context:** raw last 10 commands
* **Buffer behavior:** selected suggestion replaces the whole line
* **Safety:** confirmation happens on accept into buffer, and safety level should be configurable

That last point is important: safety is now a **policy layer**, not a hardcoded rule.

## What this implies architecturally

I think the design now wants four clear layers:

### 1. Shell adapter layer

Per-shell integration for Fish and Zsh.

Responsibilities:

* bind keys
* collect shell state
* launch the widget process
* receive selected action/result
* replace the current command line
* optionally execute if the returned action says so

### 2. Context collector

Builds a normalized request context from shell/runtime state.

Inputs:

* prompt text
* last 10 commands
* cwd
* installed tools map
* repo summary
* shell type
* safety level

### 3. Suggestion engine

Responsible for:

* debounce
* request scheduling
* OpenRouter call
* result parsing
* risk labeling
* filtering/reranking based on installed tools and safety policy

### 4. UI layer

Bottom-panel interface that:

* accepts prompt input
* shows loading state
* renders ranked suggestions
* supports selection
* prompts for confirmation when policy requires it

I think that separation is clean and gives you room to evolve the model logic without rewriting shell bindings.

## Updated design doc outline

Here’s the outline I’d use.

### 1. Overview

A shell-native AI command widget for Fish and Zsh that converts natural-language task prompts into context-aware shell commands. The widget behaves like an intent-driven alternative to history search: instead of recalling a prior command, the user describes a goal and gets ranked executable suggestions.

### 2. Goals

* feel native to interactive shell use
* minimize friction for command discovery and recall
* prefer installed and modern tools
* keep execution under explicit user control
* use local context to improve relevance without over-sharing
* support configurable safety levels

### 3. Non-goals

* fully autonomous shell agent
* background task execution
* general-purpose conversational interface
* full shell history replacement
* arbitrary environment introspection beyond the defined context surface

### 4. User experience

Flow:

* user triggers widget from Fish or Zsh
* bottom panel opens
* user types a task prompt
* after 500 ms idle, suggestions refresh
* user navigates suggestions
* selecting a suggestion replaces the current shell line
* if suggestion is above configured safety threshold, user must confirm before insertion
* optional explicit shortcut inserts and executes

This section should also define:

* close/cancel behavior
* loading behavior
* empty-state behavior
* invalid-response behavior

### 5. Interaction model

Key actions:

* open widget
* edit prompt
* move selection
* accept into buffer
* accept and execute
* cancel
* maybe cycle detail pane / explanation visibility

This section should also define whether prompt input starts empty every time or can seed from current line.

### 6. Shell support strategy

#### Zsh

* custom ZLE widget
* collect current buffer state
* replace `BUFFER`
* optionally invoke `zle accept-line`

#### Fish

* custom binding/function
* collect commandline state
* replace command line
* optionally execute via fish commandline APIs

This section should explicitly state that shell-specific code is thin and delegates to the shared widget binary.

### 7. Context model

Normalized context sent to the suggestion engine:

```json
{
  "shell": "zsh",
  "cwd": "/path/to/project",
  "prompt": "find all large log files from today",
  "history": [
    "rg TODO src",
    "git status",
    "fd test",
    "... up to 10"
  ],
  "tools": {
    "rg": true,
    "fd": true,
    "jq": true,
    "fzf": false,
    "bat": true,
    "git": true,
    "jj": false
  },
  "repo": {
    "type": "git",
    "branch": "main",
    "dirty": true,
    "staged_count": 1,
    "unstaged_count": 3,
    "untracked_count": 2,
    "ahead": 0,
    "behind": 0
  },
  "safety_level": "medium"
}
```

The doc should define:

* the exact installed-tools starter list
* the normalized repo schema
* what happens when no repo is present

### 8. Suggestion policy

Rules for ranking and filtering suggestions:

* prefer installed tools
* prefer modern tools where appropriate
* prefer read-only commands unless prompt implies mutation
* generate multiple alternatives when useful
* include assumptions/explanations
* avoid suggesting unavailable tools unless clearly marked as alternatives

This section should also define whether the model can emit both:

* primary command
* fallback command using more standard tooling

### 9. Safety model

This is now an important section.

Define:

* configurable safety levels
* risk classification of generated commands
* threshold at which confirmation is required
* distinction between:

  * low-risk
  * medium-risk
  * high-risk / destructive

Example policy:

* **low:** confirm nothing
* **medium:** confirm commands that mutate repo/filesystem in broad ways
* **high:** confirm any mutating command

We should also define whether safety classification is:

* model-provided
* locally heuristic
* or both

My recommendation: both.
The model can label risk, but the client should also run a local heuristic pass for patterns like:

* `rm`
* `mv`
* `chmod -R`
* `chown -R`
* `sed -i`
* `git reset --hard`
* `git clean`
* force push
* recursive deletes
* shell redirection to important paths

### 10. UI design

Bottom-panel layout:

* header/status line
* prompt input
* suggestions list
* optional detail pane or inline explanation
* confirmation modal/state for risky suggestions

Also define:

* number of visible suggestions
* whether explanations are always visible
* whether command wraps or scrolls horizontally
* loading indicator behavior

### 11. Suggestion engine design

Responsibilities:

* debounce prompt changes at 500 ms
* suppress stale responses
* query OpenRouter
* parse strict structured response
* rerank/filter according to local policy
* hand back UI-ready suggestions

This section should include:

* prompt design
* schema definition
* stale request handling
* retry/fallback behavior

### 12. Structured response schema

Something like:

```json
{
  "suggestions": [
    {
      "command": "fd -e log . | xargs ls -lh",
      "description": "List log files using fd and show sizes",
      "risk": "low",
      "uses_tools": ["fd", "xargs", "ls"],
      "assumptions": ["fd is installed"],
      "confidence": 0.86
    }
  ]
}
```

We may also want:

* `category`
* `mutates_state`
* `requires_confirmation`
* `fallback_command`

### 13. Concurrency and debounce

Define:

* 500 ms idle before generation
* generation/version IDs
* stale response dropping
* maybe request cancellation if supported
* how the UI behaves while an earlier request is still in flight

### 14. Performance targets

Example targets:

* widget opens quickly enough to feel shell-native
* suggestions begin appearing within ~1–2 seconds under normal conditions
* no UI blocking during network requests
* repeated identical prompts may use cache

### 15. Caching

Likely:

* in-memory session cache
* keyed by prompt + relevant context subset
* optional disk cache later

Need to decide whether history and repo state should be part of the cache key, or whether only prompt + toolset + repo type matter.

### 16. Failure handling

* API unavailable
* invalid JSON
* no suggestions
* timeout
* partial context collection failure

Behavior should be graceful:

* show error
* allow retry
* never corrupt shell state

### 17. Privacy and local context boundaries

Document exactly what is sent.
This is especially important since history and repo metadata are included.

### 18. Rollout plan

#### MVP

* Zsh + Fish
* bottom panel widget
* prompt + debounce
* structured suggestions
* configurable safety threshold
* replace whole line
* accept-and-execute shortcut

#### Later

* Bash
* richer tool detection
* better repo support
* smarter caching
* user-tunable modern-vs-standard preference

## The main unresolved design questions now

We’re close, but I think these still need decisions.

### 1. Safety level model

Should safety be:

* `off / low / medium / high`
  or
* a numeric threshold
  or
* separate toggles like:

  * confirm file mutations
  * confirm repo mutations
  * confirm network/install commands

My instinct is to start with:

* `low`
* `balanced`
* `strict`

That’s user-friendly and still expressive enough.

### 2. Confirmation copy and granularity

When a command crosses the threshold, do we show:

* one generic confirmation prompt
  or
* a reasoned prompt like:
  “This command may delete files recursively. Insert anyway?”

I strongly prefer reasoned confirmation.

### 3. Prompt seeding

When opening the widget:

* start with empty prompt
* or seed from current shell buffer if one exists

There’s a nice workflow if the current line already contains some text and the widget uses that as the starting prompt.

### 4. Suggestion count

Do we want:

* 3 suggestions
* 5 suggestions
* 7 suggestions

I think 4 or 5 is probably the sweet spot.

### 5. Modern tools preference strength

If `rg` and `grep` are both available, should we:

* strongly prefer `rg`
* or show both
* or show modern first and traditional second

I’d lean toward:

* modern first
* traditional fallback when helpful

### 6. Repo coverage

For MVP, is it:

* `git` only
* `git` + `jj`

You mentioned `jj/git/etc`, but for the first pass I’d probably normalize around `git` and optionally `jj` if it matters to you immediately.

## My current recommendation

If I were writing the first draft, I’d lock these in unless you want to change them:

* safety levels: `low`, `balanced`, `strict`
* confirmation includes specific risk reason
* widget starts with empty prompt, but current shell line can be added later
* 5 visible suggestions
* modern tool first, traditional fallback second
* repo support for MVP: `git`, with `jj` as a near-term extension

## A compact “decisions” section for the doc

I’d put this near the top:

**Decisions**

* The product is a shell-native AI command widget, not a standalone command palette.
* MVP targets Fish and Zsh.
* The UI is a transient bottom panel.
* Suggestions are generated after 500 ms of idle typing.
* Accept replaces the full shell line.
* Accept-and-execute is a separate explicit action.
* Model context includes prompt, last 10 commands, cwd, installed tools map, and normalized repo summary.
* Ranking prefers installed and modern tools.
* Safety is policy-driven and user-configurable.
* Commands above the configured risk threshold require confirmation before insertion.

## The next thing I’d do

I think the most useful next move is to write the actual design doc opening sections in prose:

* Overview
* Goals / non-goals
* User experience
* Safety model

That will expose any remaining ambiguity fast.

