# Bootstrap Fish integration for munch.
#
# The shell side stays thin: it passes shell-local state via environment
# variables and asks the Go binary to return Fish-safe assignments.

if not set -q MUNCH_BIN
    set -g MUNCH_BIN munch
end

if not set -q MUNCH_ARGS
    set -g MUNCH_ARGS
end

function __munch_widget
    if not test -x "$MUNCH_BIN"
        if not command -sq "$MUNCH_BIN"
            echo "munch binary not found: $MUNCH_BIN" >&2
            commandline -f repaint
            return 1
        end
    end

    set -l request_id "req_"(date +%s)"_"(random)
    set -l widget_cmd "$MUNCH_BIN" --mode fish-bridge

    if test -n "$MUNCH_ARGS"
        set -a widget_cmd (string split ' ' -- $MUNCH_ARGS)
    end

    set -l assignments (
        env \
            REQUEST_ID="$request_id" \
            ORIGINAL_BUFFER=(commandline) \
            PROMPT_TEXT=(commandline) \
            CURSOR_POSITION=(commandline -C) \
            $widget_cmd
    )
    or begin
        echo "munch execution failed" >&2
        commandline -f repaint
        return 1
    end

    eval $assignments

    switch "$MUNCH_ACTION"
        case cancel
            commandline -f repaint
        case insert
            commandline -r -- "$MUNCH_COMMAND"
            commandline -f repaint
        case execute
            commandline -r -- "$MUNCH_COMMAND"
            commandline -f execute
        case '*'
            echo "unsupported munch action: $MUNCH_ACTION" >&2
            commandline -f repaint
            return 1
    end
end

function __munch_bind
{{BINDINGS}}
end

__munch_bind
