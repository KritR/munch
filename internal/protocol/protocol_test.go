package protocol

import (
	"bytes"
	"strings"
	"testing"
)

func TestDecodeRequestValid(t *testing.T) {
	reqJSON := `{"schema_version":1,"request_id":"req_1","shell":"zsh","original_buffer":"ls","prompt_text":"ls","cursor_position":2}`

	req, err := DecodeRequest(strings.NewReader(reqJSON))
	if err != nil {
		t.Fatalf("DecodeRequest() error = %v", err)
	}

	if req.RequestID != "req_1" {
		t.Fatalf("unexpected request id: %q", req.RequestID)
	}
}

func TestDecodeRequestInvalidShell(t *testing.T) {
	reqJSON := `{"schema_version":1,"request_id":"req_1","shell":"bash","original_buffer":"","prompt_text":"","cursor_position":0}`

	if _, err := DecodeRequest(strings.NewReader(reqJSON)); err == nil {
		t.Fatal("expected invalid shell error")
	}
}

func TestDecodeRequestIgnoresUnknownFields(t *testing.T) {
	reqJSON := `{"schema_version":1,"request_id":"req_1","shell":"zsh","original_buffer":"ls","prompt_text":"ls","cursor_position":2,"unknown":"value"}`

	if _, err := DecodeRequest(strings.NewReader(reqJSON)); err != nil {
		t.Fatalf("expected unknown field to be ignored, got %v", err)
	}
}

func TestEncodeResponseRequiresCommandForInsert(t *testing.T) {
	var buf bytes.Buffer
	resp := ShellInvocationResponse{
		SchemaVersion: SchemaVersion,
		RequestID:     "req_1",
		Action:        ActionInsert,
	}

	if err := EncodeResponse(&buf, resp); err == nil {
		t.Fatal("expected missing command error")
	}
}

func TestEncodeResponseCancel(t *testing.T) {
	var buf bytes.Buffer
	resp := ShellInvocationResponse{
		SchemaVersion: SchemaVersion,
		RequestID:     "req_1",
		Action:        ActionCancel,
	}

	if err := EncodeResponse(&buf, resp); err != nil {
		t.Fatalf("EncodeResponse() error = %v", err)
	}
}
