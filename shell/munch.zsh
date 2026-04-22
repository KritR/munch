# Bootstrap Zsh integration for early end-to-end testing.
#
# The shell side stays thin: it passes shell-local state via environment
# variables and asks the Go widget binary to return shell-safe assignments.

: "${MUNCH_WIDGET_BIN:=munch-widget}"
: "${MUNCH_WIDGET_ARGS:=}"

function munch-widget-zle() {
  emulate -L zsh
  setopt localoptions no_aliases

  if ! command -v "$MUNCH_WIDGET_BIN" >/dev/null 2>&1; then
    zle -M "munch-widget binary not found: $MUNCH_WIDGET_BIN"
    return 1
  fi

  local request_id assignments
  request_id="req_${EPOCHSECONDS}_${RANDOM}"

  local -a widget_cmd
  widget_cmd=("$MUNCH_WIDGET_BIN" --mode zsh-bridge)
  if [[ -n "$MUNCH_WIDGET_ARGS" ]]; then
    local -a extra_args
    extra_args=(${(z)MUNCH_WIDGET_ARGS})
    widget_cmd+=("${extra_args[@]}")
  fi

  assignments=$(REQUEST_ID="$request_id" ORIGINAL_BUFFER="$BUFFER" PROMPT_TEXT="$BUFFER" CURSOR_POSITION="$CURSOR" "${widget_cmd[@]}") || {
    zle -M "munch-widget execution failed"
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
