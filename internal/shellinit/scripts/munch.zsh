# Bootstrap Zsh integration for munch.
#
# The shell side stays thin: it passes shell-local state via environment
# variables and asks the Go binary to return shell-safe assignments.

: "${MUNCH_BIN:=munch}"
: "${MUNCH_ARGS:=}"

function __munch_zle_widget() {
  emulate -L zsh
  setopt localoptions no_aliases

  if ! command -v "$MUNCH_BIN" >/dev/null 2>&1; then
    zle -M "munch binary not found: $MUNCH_BIN"
    return 1
  fi

  local request_id assignments
  request_id="req_${EPOCHSECONDS}_${RANDOM}"

  local -a widget_cmd
  widget_cmd=("$MUNCH_BIN" --mode zsh-bridge)
  if [[ -n "$MUNCH_ARGS" ]]; then
    local -a extra_args
    extra_args=(${(z)MUNCH_ARGS})
    widget_cmd+=("${extra_args[@]}")
  fi

  assignments=$(REQUEST_ID="$request_id" ORIGINAL_BUFFER="$BUFFER" PROMPT_TEXT="$BUFFER" CURSOR_POSITION="$CURSOR" "${widget_cmd[@]}") || {
    zle -M "munch execution failed"
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

function __munch_bind_zle() {
  zle -N __munch_zle_widget
{{BINDINGS}}
}

__munch_bind_zle
