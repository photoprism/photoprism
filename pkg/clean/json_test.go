package clean

import "testing"

func TestJSON(t *testing.T) {
	t.Run("CodeFence", func(t *testing.T) {
		payload := "```json\n{\"labels\":[]}\n```\nextra"
		expected := "{\"labels\":[]}"
		if got := JSON(payload); got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})
	t.Run("PlainWithPrefix", func(t *testing.T) {
		payload := "Here you go: {\"labels\":[1]} thanks"
		expected := "{\"labels\":[1]}"
		if got := JSON(payload); got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})
	t.Run("Array", func(t *testing.T) {
		payload := "```\n[1,2,3]\n```"
		expected := "[1,2,3]"
		if got := JSON(payload); got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})
	t.Run("Empty", func(t *testing.T) {
		if got := JSON("   "); got != "" {
			t.Fatalf("expected empty, got %q", got)
		}
	})
}
