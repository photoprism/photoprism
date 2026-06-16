package txt

import "testing"

func TestJoinAnd(t *testing.T) {
	tests := []struct {
		name   string
		input  []string
		expect string
	}{
		{"empty", []string{}, ""},
		{"single", []string{"caption"}, "caption"},
		{"two", []string{"caption", "labels"}, "caption and labels"},
		{"three", []string{"captions", "labels", "faces"}, "captions, labels, and faces"},
		{"many", []string{"one", "two", "three", "four"}, "one, two, three, and four"},
	}

	for _, tc := range tests {
		if got := JoinAnd(tc.input); got != tc.expect {
			t.Fatalf("%s: expected %q, got %q", tc.name, tc.expect, got)
		}
	}
}

func TestJoinOr(t *testing.T) {
	tests := []struct {
		name   string
		input  []string
		expect string
	}{
		{"empty", []string{}, ""},
		{"single", []string{"admin"}, "admin"},
		{"two", []string{"admin", "guest"}, "admin or guest"},
		{"three", []string{"admin", "guest", "user"}, "admin, guest, or user"},
		{"many", []string{"admin", "manager", "user", "viewer"}, "admin, manager, user, or viewer"},
	}

	for _, tc := range tests {
		if got := JoinOr(tc.input); got != tc.expect {
			t.Fatalf("%s: expected %q, got %q", tc.name, tc.expect, got)
		}
	}
}
