package clean

import "testing"

func TestThumb(t *testing.T) {
	cases := []struct {
		name string
		in   string
		out  string
	}{
		{"valid", "6f6cbaa6ae8ead9da7ee99ab66aca1ae7eed8d5c", "6f6cbaa6ae8ead9da7ee99ab66aca1ae7eed8d5c"},
		{"upper", "6F6CBAA6AE8EAD9DA7EE99AB66ACA1AE7EED8D5C", "6f6cbaa6ae8ead9da7ee99ab66aca1ae7eed8d5c"},
		{"trimmed", " 6f6cbaa6ae8ead9da7ee99ab66aca1ae7eed8d5c ", "6f6cbaa6ae8ead9da7ee99ab66aca1ae7eed8d5c"},
		{"invalidLength", "123", ""},
		{"invalidChars", "zz6cbaa6ae8ead9da7ee99ab66aca1ae7eed8d5c", ""},
	}

	for _, tc := range cases {
		if got := Thumb(tc.in); got != tc.out {
			t.Fatalf("%s: Thumb(%q) = %q, want %q", tc.name, tc.in, got, tc.out)
		}
	}
}

func TestThumbCrop(t *testing.T) {
	cases := []struct {
		name string
		in   string
		out  string
	}{
		{"valid", "6f6cbaa6ae8ead9da7ee99ab66aca1ae7eed8d5c-0910162fd2fd", "6f6cbaa6ae8ead9da7ee99ab66aca1ae7eed8d5c-0910162fd2fd"},
		{"upper", "6F6CBAA6AE8EAD9DA7EE99AB66ACA1AE7EED8D5C-0910162FD2FD", "6f6cbaa6ae8ead9da7ee99ab66aca1ae7eed8d5c-0910162fd2fd"},
		{"trimmed", " 6f6cbaa6ae8ead9da7ee99ab66aca1ae7eed8d5c-0910162fd2fd ", "6f6cbaa6ae8ead9da7ee99ab66aca1ae7eed8d5c-0910162fd2fd"},
		{"missingDash", "6f6cbaa6ae8ead9da7ee99ab66aca1ae7eed8d5c0910162fd2fd", ""},
		{"invalidHash", "zz6cbaa6ae8ead9da7ee99ab66aca1ae7eed8d5c-0910162fd2fd", ""},
		{"invalidCrop", "6f6cbaa6ae8ead9da7ee99ab66aca1ae7eed8d5c-zzzz", ""},
		{"shortCrop", "6f6cbaa6ae8ead9da7ee99ab66aca1ae7eed8d5c-0910", ""},
	}

	for _, tc := range cases {
		if got := ThumbCrop(tc.in); got != tc.out {
			t.Fatalf("%s: ThumbCrop(%q) = %q, want %q", tc.name, tc.in, got, tc.out)
		}
	}
}
