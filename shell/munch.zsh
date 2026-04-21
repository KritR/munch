# Bootstrap Zsh integration for early end-to-end testing.
#
# This keeps the shell side thin while the Go widget binary owns request
# validation and response generation. JSON encoding/decoding is bridged with
# python3 for now so the shell code stays simple during the bootstrap phase.

: "${MUNCH_WIDGET_BIN:=munch-widget}"

function _munch_widget_build_request_json() {
  local request_id="$1"
  local buffer="$2"
  local cursor="$3"

  REQUEST_ID="$request_id" \
  ORIGINAL_BUFFER="$buffer" \
  PROMPT_TEXT="$buffer" \
  CURSOR_POSITION="$cursor" \
  python3 - <<'PY'
import json
import os

payload = {
    "schema_version": 1,
    "request_id": os.environ["REQUEST_ID"],
    "shell": "zsh",
    "original_buffer": os.environ["ORIGINAL_BUFFER"],
    "prompt_text": os.environ["PROMPT_TEXT"],
    "cursor_position": int(os.environ["CURSOR_POSITION"]),
}

print(json.dumps(payload))
PY
}

function _munch_widget_parse_response_assignments() {
  local response="$1"

  RESPONSE_JSON="$response" python3 - <<'PY'
import json
import os
import shlex
import sys

data = json.loads(os.environ["RESPONSE_JSON"])
action = data["action"]
command = data.get("command", "")

print(f"MUNCH_ACTION={shlex.quote(action)}")
print(f"MUNCH_COMMAND={shlex.quote(command)}")
PY
}

function munch-widget-zle() {
  emulate -L zsh
  setopt localoptions no_aliases

  if ! command -v "$MUNCH_WIDGET_BIN" >/dev/null 2>&1; then
    zle -M "munch-widget binary not found: $MUNCH_WIDGET_BIN"
    return 1
  fi

  if ! command -v python3 >/dev/null 2>&1; then
    zle -M "python3 is required for the bootstrap JSON bridge"
    return 1
  fi

  local request_id payload response assignments
  request_id="req_${EPOCHSECONDS}_${RANDOM}"
  payload=$(_munch_widget_build_request_json "$request_id" "$BUFFER" "$CURSOR") || {
    zle -M "failed to build munch request"
    return 1
  }

  response=$(printf '%s' "$payload" | "$MUNCH_WIDGET_BIN" --mode session) || {
    zle -M "munch-widget execution failed"
    return 1
  }

  assignments=$(_munch_widget_parse_response_assignments "$response") || {
    zle -M "invalid munch-widget response"
    return 1
  }

  eval "$assignments"

  case "$MUNCH_ACTION" in
    cancel)
      zle reset-prompt
      ;;
    insert)
      BUFFER="$MUNCH_COMMAND"
      CURSOR=${#BUFFER}
      zle reset-prompt
      ;;
    execute)
      BUFFER="$MUNCH_COMMAND"
      CURSOR=${#BUFFER}
      zle accept-line
      ;;
    *)
      zle -M "unsupported munch action: $MUNCH_ACTION"
      return 1
      ;;
  esac
}

function munch-widget-bind-zle() {
  zle -N munch-widget-zle
  bindkey '^G' munch-widget-zle
}
