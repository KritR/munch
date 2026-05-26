package command

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/krithikr/munch/internal/app"
	fishbridge "github.com/krithikr/munch/internal/bridge/fish"
	zshbridge "github.com/krithikr/munch/internal/bridge/zsh"
	"github.com/krithikr/munch/internal/protocol"
	"github.com/krithikr/munch/internal/runtime"
	"github.com/krithikr/munch/internal/shellinit"
	cli "github.com/urfave/cli/v3"
)

var Version = "dev"

type IO struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

func Run(ctx context.Context, args []string, ioStreams IO) error {
	cmd := New(ioStreams)
	return cmd.Run(ctx, args)
}

func New(ioStreams IO) *cli.Command {
	if ioStreams.Stdin == nil {
		ioStreams.Stdin = io.Reader(nil)
	}
	cmd := &cli.Command{
		Name:                  "munch",
		Usage:                 "AI command suggestions for your shell",
		Version:               Version,
		EnableShellCompletion: true,
		Reader:                ioStreams.Stdin,
		Writer:                ioStreams.Stdout,
		ErrWriter:             ioStreams.Stderr,
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "mode", Value: "session", Usage: "run mode", Hidden: true},
			&cli.StringFlag{Name: "config", Usage: "path to config file"},
			&cli.StringFlag{Name: "dev-action", Value: "none", Usage: "developer action override", Hidden: true},
		},
		Commands: []*cli.Command{
			initCommand(ioStreams.Stdout),
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			return runMode(c.String("mode"), c.String("config"), runtime.DevMode(c.String("dev-action")), ioStreams)
		},
	}
	return cmd
}

func initCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "print shell integration code",
		Commands: []*cli.Command{
			initShellCommand("zsh", stdout),
			initShellCommand("fish", stdout),
		},
	}
}

func initShellCommand(shell string, stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:      shell,
		Usage:     fmt.Sprintf("print %s integration code", shell),
		UsageText: fmt.Sprintf("munch init %s [--key <sequence>|--no-bind]", shell),
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "key", Value: shellinit.DefaultKey(shell), Usage: "key sequence to bind"},
			&cli.BoolFlag{Name: "no-bind", Usage: "print functions without installing a keybinding"},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			script, err := shellinit.Script(shell, shellinit.Options{
				Key:    c.String("key"),
				NoBind: c.Bool("no-bind"),
			})
			if err != nil {
				return err
			}
			_, err = fmt.Fprint(stdout, script)
			return err
		},
	}
}

func runMode(mode string, configPath string, devMode runtime.DevMode, ioStreams IO) error {
	switch mode {
	case "session":
		return runSession(configPath, devMode, ioStreams)
	case "fish-bridge":
		return runFishBridge(configPath, devMode, ioStreams.Stdout)
	case "zsh-bridge":
		return runZshBridge(configPath, devMode, ioStreams.Stdout)
	default:
		return fmt.Errorf("unsupported mode: %s", mode)
	}
}

func runSession(configPath string, devMode runtime.DevMode, ioStreams IO) error {
	slog.Debug("starting session", "config_path", configPath, "dev_mode", devMode)
	req, err := protocol.DecodeRequest(ioStreams.Stdin)
	if err != nil {
		return err
	}
	resp, err := app.RunSession(req, configPath, devMode)
	if err != nil {
		return err
	}
	return protocol.EncodeResponse(ioStreams.Stdout, resp)
}

func runZshBridge(configPath string, devMode runtime.DevMode, stdout io.Writer) error {
	req, err := zshbridge.RequestFromEnv()
	if err != nil {
		return err
	}
	resp, err := app.RunSession(req, configPath, devMode)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(stdout, zshbridge.ResponseAssignments(resp))
	return err
}

func runFishBridge(configPath string, devMode runtime.DevMode, stdout io.Writer) error {
	req, err := fishbridge.RequestFromEnv()
	if err != nil {
		return err
	}
	resp, err := app.RunSession(req, configPath, devMode)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(stdout, fishbridge.ResponseAssignments(resp))
	return err
}
