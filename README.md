# munch

`munch` is an AI shell command widget for Zsh and Fish.

It opens an inline terminal UI from your current shell prompt, turns a natural-language task into command suggestions, and inserts or executes the selected command.

## Basic Usage

1. Type a task at your shell prompt.
2. Press `Ctrl-G`.
3. Pick a suggested command from the widget.
4. Insert it into the prompt or run it directly.

Example:

```text
$ find all large log files modified today
```

Then press `Ctrl-G` and let `munch` turn that into shell commands.

A short terminal recording lives at [docs/recordings/demo.cast](/Users/krithikr/Desktop/playground/munch/docs/recordings/demo.cast:1).

## Install

### macOS

Install with Homebrew:

```sh
brew install --cask KritR/munch/munch
```

Or tap first:

```sh
brew tap KritR/munch https://github.com/KritR/homebrew-munch
brew install --cask munch
```

### Linux

Download the matching `munch_<version>_linux_<arch>.tar.gz` archive from GitHub Releases, extract it, and place the `munch` binary on `PATH`.

## Setup

Create the config directory:

```sh
mkdir -p "${XDG_CONFIG_HOME:-$HOME/.config}/munch"
```

Then create `${XDG_CONFIG_HOME:-$HOME/.config}/munch/config.toml`:

```toml
[provider]
base_url = "https://api.cerebras.ai"
model = "llama3.1-8b"
api_key_env = "CEREBRAS_API_KEY"
timeout_ms = 4000
max_retries = 1

[ui]
visible_suggestion_count = 5

[safety]
level = "balanced"

[telemetry]
enabled = false
```

Export your provider key in your shell config:

```sh
export CEREBRAS_API_KEY="..."
```

Enable shell integration.

Zsh:

```zsh
eval "$(munch init zsh)"
```

Fish:

```fish
munch init fish | source
```

Start a new shell and verify the install:

```sh
munch --version
```

## Usage

Once installed and configured:

* type a task at the prompt
* press `Ctrl-G` to open the widget
* move through suggestions
* accept one to insert or execute it

Useful commands:

```sh
munch --version
munch init zsh
munch init fish
```

## Customization

Change the default key binding in Zsh:

```zsh
eval "$(munch init zsh --key '^X^M')"
```

Change the default key binding in Fish:

```fish
munch init fish --key \cx\cm | source
```

Disable automatic keybinding and wire it yourself:

```zsh
eval "$(munch init zsh --no-bind)"
```

```fish
munch init fish --no-bind | source
```

Generate shell completions.

Zsh:

```zsh
munch completion zsh > "${fpath[1]}/_munch"
```

Fish:

```fish
munch completion fish > ~/.config/fish/completions/munch.fish
```

## Development

### Install from source

Prerequisites:

* Go 1.24 or newer
* Zsh or Fish

Build the binary:

```sh
make build
```

Install it somewhere on `PATH`:

```sh
make install PREFIX=/usr/local
```

For a user-local install:

```sh
make install PREFIX="$HOME/.local"
```

For debugging:

```sh
export MUNCH_LOG_FILE=/tmp/munch.log
```
