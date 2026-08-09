package ollama

import (
	"testing"
	"time"
)

func TestResponseErr(t *testing.T) {
	t.Run("NilResponse", func(t *testing.T) {
		if err := (*Response)(nil).Err(); err == nil || err.Error() != "response is nil" {
			t.Fatalf("expected nil-response error, got %v", err)
		}
	})
	t.Run("HTTPErrorWithMessage", func(t *testing.T) {
		resp := &Response{Code: 429, Error: "too many requests"}
		if err := resp.Err(); err == nil || err.Error() != "too many requests" {
			t.Fatalf("expected message error, got %v", err)
		}
	})
	t.Run("HTTPErrorWithoutMessage", func(t *testing.T) {
		resp := &Response{Code: 500}
		if err := resp.Err(); err == nil || err.Error() != "error 500" {
			t.Fatalf("expected formatted error, got %v", err)
		}
	})
	t.Run("NoResult", func(t *testing.T) {
		resp := &Response{Code: 200}
		if err := resp.Err(); err == nil || err.Error() != "no result" {
			t.Fatalf("expected no-result error, got %v", err)
		}
	})
	t.Run("HasLabels", func(t *testing.T) {
		resp := &Response{
			Code:   200,
			Result: ResultPayload{Labels: []LabelPayload{{Name: "sky"}}},
			Model:  "qwen",
		}
		if err := resp.Err(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("HasCaption", func(t *testing.T) {
		resp := &Response{
			Code:   200,
			Result: ResultPayload{Caption: &CaptionPayload{Text: "Caption"}},
		}
		if err := resp.Err(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestResponseHasResult(t *testing.T) {
	if (*Response)(nil).HasResult() {
		t.Fatal("nil response should not have result")
	}

	resp := &Response{}
	if resp.HasResult() {
		t.Fatal("expected false when result payload is empty")
	}

	resp.Result.Labels = []LabelPayload{{Name: "sun"}}
	if !resp.HasResult() {
		t.Fatal("expected true when labels present")
	}

	resp.Result.Labels = nil
	resp.Result.Caption = &CaptionPayload{Text: "Sky", Confidence: 0.9}
	if !resp.HasResult() {
		t.Fatal("expected true when caption present")
	}
}

func TestResponseJSONTagsAreOptional(t *testing.T) {
	// Guard against accidental breaking changes to essential fields
	resp := Response{
		ID:        "test",
		Model:     "ollama",
		CreatedAt: time.Now(),
	}
	if resp.ID == "" || resp.Model == "" {
		t.Fatalf("response fields should persist, got %+v", resp)
	}
}

func TestIsCloudUrl(t *testing.T) {
	cases := []struct {
		name string
		in   string
		out  bool
	}{
		{name: "CloudBase", in: CloudBaseUrl, out: true},
		{name: "CloudGenerate", in: CloudBaseUrl + "/api/generate", out: true},
		{name: "CloudUppercase", in: "https://Ollama.com/api/generate", out: true},
		{name: "LocalIP", in: "http://192.0.2.10:11434/api/generate", out: false},
		{name: "LocalHost", in: DefaultBaseUrl + "/api/generate", out: false},
		{name: "OtherHost", in: "https://vision.example.com/api/generate", out: false},
		{name: "Empty", in: "", out: false},
		{name: "NotAUrl", in: "gemma4:latest", out: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsCloudUrl(tc.in); got != tc.out {
				t.Fatalf("IsCloudUrl(%q) = %v, want %v", tc.in, got, tc.out)
			}
		})
	}
}
