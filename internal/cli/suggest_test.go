package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/grokipedia/cli/internal/api"
)

func TestOutputSuggestResultsJSON(t *testing.T) {
	resp := &api.SuggestArticleResponse{
		Success: true,
		ID:      "req-123",
		Status:  "PENDING",
		Message: "Request submitted",
	}

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputSuggestResults(resp, "json", "Test Article")

	if err := w.Close(); err != nil {
		t.Fatalf("Failed to close writer: %v", err)
	}
	os.Stdout = oldStdout

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("Failed to read from pipe: %v", err)
	}
	output := buf.String()

	if err != nil {
		t.Errorf("outputSuggestResults() error = %v", err)
	}

	// Verify JSON output
	var parsed api.SuggestArticleResponse
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Errorf("Invalid JSON output: %v", err)
	}

	if !parsed.Success {
		t.Error("Expected success to be true")
	}

	if parsed.ID != "req-123" {
		t.Errorf("Expected ID 'req-123', got '%s'", parsed.ID)
	}
}

func TestOutputSuggestResultsTextSuccess(t *testing.T) {
	resp := &api.SuggestArticleResponse{
		Success: true,
		ID:      "req-456",
		Status:  "PENDING_REVIEW",
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputSuggestResults(resp, "text", "CRISPR")

	if err := w.Close(); err != nil {
		t.Fatalf("Failed to close writer: %v", err)
	}
	os.Stdout = oldStdout

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("Failed to read from pipe: %v", err)
	}
	output := buf.String()

	if err != nil {
		t.Errorf("outputSuggestResults() error = %v", err)
	}

	if !strings.Contains(output, "CRISPR") {
		t.Error("Expected output to contain 'CRISPR'")
	}

	if !strings.Contains(output, "req-456") {
		t.Error("Expected output to contain request ID")
	}

	if !strings.Contains(output, "PENDING_REVIEW") {
		t.Error("Expected output to contain status")
	}

	if !strings.Contains(output, "Successfully submitted") {
		t.Error("Expected success message")
	}
}

func TestOutputSuggestResultsTextFailure(t *testing.T) {
	resp := &api.SuggestArticleResponse{
		Success: false,
		Message: "Article already exists",
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputSuggestResults(resp, "text", "Python")

	if err := w.Close(); err != nil {
		t.Fatalf("Failed to close writer: %v", err)
	}
	os.Stdout = oldStdout

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("Failed to read from pipe: %v", err)
	}
	output := buf.String()

	if err == nil {
		t.Error("Expected error for failed suggestion")
	}

	if !strings.Contains(output, "Python") {
		t.Error("Expected output to contain 'Python'")
	}

	if !strings.Contains(output, "Article already exists") {
		t.Error("Expected output to contain error message")
	}

	if !strings.Contains(output, "Failed to submit") {
		t.Error("Expected failure message")
	}
}

func TestOutputSuggestResultsInvalidFormat(t *testing.T) {
	resp := &api.SuggestArticleResponse{Success: true}

	err := outputSuggestResults(resp, "xml", "Test")

	if err == nil {
		t.Error("Expected error for invalid format")
	}

	if !strings.Contains(err.Error(), "invalid format") {
		t.Errorf("Expected 'invalid format' error, got: %v", err)
	}
}

func TestOutputSuggestResultsEmptyID(t *testing.T) {
	resp := &api.SuggestArticleResponse{
		Success: true,
		Status:  "ACCEPTED",
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputSuggestResults(resp, "text", "Test")

	if err := w.Close(); err != nil {
		t.Fatalf("Failed to close writer: %v", err)
	}
	os.Stdout = oldStdout

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("Failed to read from pipe: %v", err)
	}
	output := buf.String()

	if err != nil {
		t.Errorf("outputSuggestResults() error = %v", err)
	}

	if !strings.Contains(output, "Test") {
		t.Error("Expected output to contain 'Test'")
	}

	if !strings.Contains(output, "ACCEPTED") {
		t.Error("Expected output to contain status")
	}

	// Should not contain Request ID line when ID is empty
	if strings.Contains(output, "Request ID:") {
		t.Error("Should not show Request ID when empty")
	}
}
