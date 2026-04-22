# Bootstrap Fish integration for early end-to-end testing.
#
# The shell side stays thin: it passes shell-local state via environment
# variables and asks the Go widget binary to return Fish-safe assignments.

if not set -q MUNCH_WIDGET_BIN
    set -g MUNCH_WIDGET_BIN munch-widget
end

if not set -q MUNCH_WIDGET_ARGS
    set -g MUNCH_WIDGET_ARGS
end

function munch-widget
    if not test -x "$MUNCH_WIDGET_BIN"
        if not command -sq "$MUNCH_WIDGET_BIN"
            echo "munch-widget binary not found: $MUNCH_WIDGET_BIN" >&2
            commandline -f repaint
            return 1
        end
    end

    set -l request_id "req_"(date +%s)"_"(random)
    set -l widget_cmd "$MUNCH_WIDGET_BIN" --mode fish-bridge

    if test -n "$MUNCH_WIDGET_ARGS"
        set -a widget_cmd (string split ' ' -- $MUNCH_WIDGET_ARGS)
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
        echo "munch-widget execution failed" >&2
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

function munch-widget-bind
    bind \cg munch-widget
    bind -M insert \cg munch-widget
    bind -M default \cg munch-widget
end
