package chat

import (
	"testing"
)

func TestNewChatClient(t *testing.T) {
	client := NewChatClient("ws://localhost:8888")
	if client == nil {
		t.Fatal("Expected ChatClient instance, got nil")
	}

	if client.url != "ws://localhost:8888" {
		t.Errorf("Expected URL to be 'ws://localhost:8888', got '%s'", client.url)
	}

	if client.listener == nil {
		t.Error("Expected listener channel to be initialized")
	}
}

func TestMessageArgsToJSON(t *testing.T) {
	msgArgs := &MessageArgs{
		Command: "test.command",
		Params:  []interface{}{"param1", 123},
	}

	jsonStr, err := msgArgs.ToJSON()
	if err != nil {
		t.Errorf("ToJSON should not return error: %v", err)
	}

	expected := `{"cmd":"test.command","args":["param1",123]}` + "\n"
	if jsonStr != expected {
		t.Errorf("Expected JSON to be '%s', got '%s'", expected, jsonStr)
	}
}

func TestParseMessageArgs(t *testing.T) {
	jsonStr := `{"cmd":"test.command","args":["param1",123]}`

	msgArgs := ParseMessageArgs(jsonStr)
	if msgArgs == nil {
		t.Fatal("Expected MessageArgs instance, got nil")
	}

	if msgArgs.Command != "test.command" {
		t.Errorf("Expected Command to be 'test.command', got '%s'", msgArgs.Command)
	}

	if len(msgArgs.Params) != 2 {
		t.Errorf("Expected 2 params, got %d", len(msgArgs.Params))
	}
}
