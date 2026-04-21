# **AI Shell Command Widget (munch) — Design Document (Draft v0.1)**

## **1. Overview**

This project introduces a **shell-native AI command widget** for Fish and Zsh that enables users to generate executable shell commands from natural language prompts.

Unlike traditional history search (e.g., reverse-i-search), which requires recalling previously executed commands, this widget is **intent-driven**: the user describes a task, and the system returns a set of context-aware command suggestions tailored to the current environment.

The widget integrates directly into the shell experience, behaving like a native interactive feature rather than a standalone tool. It prioritizes:

* immediacy (fast suggestions)
* relevance (local context awareness)
* safety (user-controlled execution)

---

## **2. Goals**

### Primary Goals

* Provide **fast, relevant command suggestions** from natural language input
* Feel **native to interactive shell workflows** (similar to history search)
* Reduce friction in:

  * command recall
  * command composition
  * tool discovery
* Leverage **local context** (cwd, tools, repo state, history) to improve results
* Maintain **explicit user control over execution**

### Secondary Goals

* Encourage use of **modern tooling** (e.g., `rg`, `fd`, `jq`)
* Provide **clear explanations** for suggested commands
* Support **configurable safety policies**

---

## **3. Non-Goals**

* Autonomous command execution without user confirmation
* Replacing the shell or acting as a full shell environment
* Acting as a general-purpose conversational assistant
* Full replacement of history search (this is complementary)
* Deep introspection of local system state beyond defined context boundaries

---

## **4. User Experience**

### Entry Point

* User triggers the widget via a keybinding in Fish or Zsh
* The widget opens as a **transient bottom panel**

### Interaction Flow

1. User types a natural-language task:

   ```
   find all large log files modified today
   ```
2. After **500 ms of idle typing**, suggestions are generated
3. Suggestions appear in a ranked list
4. User navigates suggestions using arrow keys
5. On selection:

   * The selected command **replaces the current shell line**
   * If the command exceeds the configured safety threshold:

     * A confirmation prompt is shown before insertion
6. Optional:

   * A separate shortcut inserts and immediately executes

### Exit Behavior

* `Esc`: closes widget without modifying shell state
* Accepting a suggestion closes the widget
* Errors do not modify the shell buffer

### States

* Idle (no input)
* Typing
* Loading (waiting for suggestions)
* Results available
* Confirmation (for risky commands)
* Error (API failure / parsing failure)

---

## **5. Interaction Model**

### Key Actions

* Open widget
* Edit prompt
* Navigate suggestions
* Accept suggestion (insert into buffer)
* Accept + execute
* Cancel/close
* Confirm (if required)

### Buffer Behavior

* Accepted suggestion **replaces entire shell line**
* Cursor is positioned at end of inserted command

### Prompt Behavior

* Prompt starts empty on open (v1)
* Future enhancement: seed from existing shell buffer

---

## **6. Context Model**

The system constructs a normalized context object for each query:

```json
{
  "shell": "zsh",
  "cwd": "/Users/foo/project",
  "prompt": "kill process on port 3000",
  "history": [
    "npm run dev",
    "lsof -i :3000",
    "git status",
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
  "safety_level": "balanced"
}
```

### Notes

* History is **raw last 10 commands**, no outputs
* Tools are derived from a **predefined capability list**
* Repo context is **normalized**, not raw CLI output
* If no repo is present → `repo.type = null`

---

## **7. Suggestion Policy**

The system generates and ranks suggestions based on:

### Preferences

* Prefer **installed tools**
* Prefer **modern tools** (e.g., `rg` over `grep`)
* Provide **fallback commands** using standard tools when appropriate
* Prefer **read-only commands** unless mutation is implied

### Output Characteristics

Each suggestion includes:

* command
* short description
* risk classification
* assumptions (optional)
* confidence score (optional)

### Example

```json
{
  "command": "lsof -i :3000",
  "description": "List processes using port 3000",
  "risk": "low"
}
```

---

## **8. Safety Model**

Safety is **policy-driven and configurable**.

### Safety Levels

* `low`: minimal confirmation
* `balanced`: confirm medium/high risk commands
* `strict`: confirm any command that mutates system or repo

### Risk Categories

* **low**: read-only (`ls`, `find`, `rg`)
* **medium**: local mutation (`mkdir`, `cp`, `git add`)
* **high**: destructive or irreversible (`rm`, `git reset --hard`, `dd`)

### Enforcement

* Commands exceeding threshold require confirmation **before insertion**
* Confirmation includes **specific reasoning**, e.g.:

  > This command may delete files recursively. Insert anyway?

### Risk Classification Source

* Primary: model-provided label
* Secondary: local heuristic validation (pattern matching)

---

## **9. UI Design**

### Layout

```
-----------------------------------------
munch                          loading...
-----------------------------------------

Task:
find large files modified today

-----------------------------------------
> fd -e log -t f -x ls -lh {}
  Uses fd to find log files and list sizes

  find . -type f -name '*.log' -mtime -1 -exec ls -lh {} \;
  Standard find-based alternative

-----------------------------------------
```

### Components

* header (status, model, loading)
* prompt input
* suggestion list
* optional inline explanation
* confirmation overlay (when needed)

### Behavior

* ~5 visible suggestions
* selected item highlighted
* explanations always visible (v1)
* loading indicator shown during request

---

## **10. Suggestion Engine**

Responsibilities:

* debounce input (500 ms)
* manage async requests
* drop stale responses
* call OpenRouter API
* parse structured output
* rerank/filter results locally
* attach risk classification

### Request lifecycle

1. user types → debounce reset
2. debounce fires → request sent
3. response arrives → validated
4. UI updated if request is current

---

## **11. Structured Response Schema**

```json
{
  "suggestions": [
    {
      "command": "rg -n TODO .",
      "description": "Search for TODOs using ripgrep",
      "risk": "low",
      "uses_tools": ["rg"],
      "assumptions": ["rg is installed"],
      "confidence": 0.9
    }
  ]
}
```

---

## **12. Concurrency and Debounce**

* 500 ms idle delay before generation
* each request tagged with version ID
* stale responses ignored
* UI remains responsive during requests

---

## **13. Performance Targets**

* widget opens instantly (<100 ms perceived)
* suggestions appear within ~1–2 seconds
* no UI blocking during network calls
* repeated queries may be cached

---

## **14. Failure Handling**

Cases:

* API unavailable
* invalid JSON
* timeout
* empty results

Behavior:

* show error message in UI
* allow retry
* do not modify shell buffer

---

## **15. Privacy and Context Boundaries**

Sent:

* prompt
* last 10 commands
* cwd
* tool availability
* repo summary

Not sent:

* file contents
* environment variables
* full history
* secrets

---

## **16. Rollout Plan**

### MVP

* Zsh + Fish support
* bottom panel UI
* debounce + suggestions
* structured responses
* safety confirmation
* insert + optional execute

### Next

* Bash support
* richer tool detection
* improved repo support (`jj`)
* caching improvements
* configurable preferences
