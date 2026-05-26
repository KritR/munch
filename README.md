# munch

`munch` is an AI shell command widget for Zsh and Fish.

It opens an inline terminal UI from your current shell prompt, turns a natural-language task into command suggestions, and inserts or executes the selected command.

## Install from source

Prerequisites:

* Go 1.24 or newer
* Zsh or Fish
* a Cerebras API key for provider-backed suggestions

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

Make sure the install prefix is on `PATH`.

## Configure

Create the config file at:

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

## Set Up Zsh

Add this to `~/.zshrc`:

```zsh
eval "$(munch init zsh)"
```

By default this binds `Ctrl-G`. To choose a different binding:

```zsh
eval "$(munch init zsh --key '^X^M')"
```

## Set Up Fish

Add this to `~/.config/fish/config.fish`:

```fish
munch init fish | source
```

By default this binds `Ctrl-G`. To choose a different binding:

```fish
munch init fish --key \cx\cm | source
```

## Shell Completions

Zsh:

```zsh
munch completion zsh > "${fpath[1]}/_munch"
```

Fish:

```fish
munch completion fish > ~/.config/fish/completions/munch.fish
```

## Verify

Start a new shell, then run:

```sh
munch --version
```

Type a task at your prompt, press `Ctrl-G`, and choose a suggestion from the widget.

For debugging, set `MUNCH_LOG_FILE` before invoking the widget:

```sh
export MUNCH_LOG_FILE=/tmp/munch.log
```
