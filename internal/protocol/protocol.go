package protocol

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

const SchemaVersion = 1

type Shell string

const (
	ShellZsh  Shell = "zsh"
	ShellFish Shell = "fish"
)

type Action string

const (
	ActionCancel  Action = "cancel"
	ActionInsert  Action = "insert"
	ActionExecute Action = "execute"
)

type ShellInvocationRequest struct {
	SchemaVersion  int    `json:"schema_version"`
	RequestID      string `json:"request_id"`
	Shell          Shell  `json:"shell"`
	OriginalBuffer string `json:"original_buffer"`
	PromptText     string `json:"prompt_text"`
	CursorPosition int    `json:"cursor_position"`
}

func (r ShellInvocationRequest) Validate() error {
	if r.SchemaVersion != SchemaVersion {
		return fmt.Errorf("unsupported schema version: %d", r.SchemaVersion)
	}
	if r.RequestID == "" {
		return errors.New("request_id is required")
	}
	switch r.Shell {
	case ShellZsh, ShellFish:
	default:
		return fmt.Errorf("unsupported shell: %q", r.Shell)
	}
	if r.CursorPosition < 0 {
		return errors.New("cursor_position must be >= 0")
	}
	return nil
}

type ShellInvocationResponse struct {
	SchemaVersion           int    `json:"schema_version"`
	RequestID               string `json:"request_id"`
	Action                  Action `json:"action"`
	Command                 string `json:"command,omitempty"`
	SelectedSuggestionIndex *int   `json:"selected_suggestion_index,omitempty"`
}

func (r ShellInvocationResponse) Validate() error {
	if r.SchemaVersion != SchemaVersion {
		return fmt.Errorf("unsupported schema version: %d", r.SchemaVersion)
	}
	if r.RequestID == "" {
		return errors.New("request_id is required")
	}
	switch r.Action {
	case ActionCancel:
		return nil
	case ActionInsert, ActionExecute:
		if r.Command == "" {
			return fmt.Errorf("command is required for %q", r.Action)
		}
	default:
		return fmt.Errorf("unsupported action: %q", r.Action)
	}
	if r.SelectedSuggestionIndex != nil && *r.SelectedSuggestionIndex < 0 {
		return errors.New("selected_suggestion_index must be >= 0")
	}
	return nil
}

func DecodeRequest(r io.Reader) (ShellInvocationRequest, error) {
	var req ShellInvocationRequest
	dec := json.NewDecoder(r)
	if err := dec.Decode(&req); err != nil {
		return ShellInvocationRequest{}, err
	}
	if err := req.Validate(); err != nil {
		return ShellInvocationRequest{}, err
	}
	return req, nil
}

func EncodeResponse(w io.Writer, resp ShellInvocationResponse) error {
	if err := resp.Validate(); err != nil {
		return err
	}
	enc := json.NewEncoder(w)
	return enc.Encode(resp)
}
