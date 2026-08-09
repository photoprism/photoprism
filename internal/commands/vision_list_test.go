package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVisionEndpoint(t *testing.T) {
	cases := []struct {
		name   string
		uri    string
		method string
		want   string
	}{
		{
			name:   "Plain",
			uri:    "http://ollama:11434/api/generate",
			method: "POST",
			want:   "POST http://ollama:11434/api/generate",
		},
		{ //nolint:gosec // example URL, the password in it is exactly what this case redacts
			name:   "BasicAuth",
			uri:    "https://vision:secret@vision.example.com/api/generate",
			method: "POST",
			want:   "POST https://vision:xxxxx@vision.example.com/api/generate",
		},
		{
			name:   "UsernameOnly",
			uri:    "https://vision@vision.example.com/api/generate",
			method: "POST",
			want:   "POST https://vision@vision.example.com/api/generate",
		},
		{
			name:   "QueryIsKept",
			uri:    "https://api.example.com/v1/responses?tier=flex",
			method: "POST",
			want:   "POST https://api.example.com/v1/responses?tier=flex",
		},
		{
			name:   "MissingUri",
			uri:    "",
			method: "POST",
			want:   "",
		},
		{
			name:   "MissingMethod",
			uri:    "https://api.example.com/v1/responses",
			method: "",
			want:   "",
		},
		{
			name:   "Unparsable",
			uri:    "://nope",
			method: "POST",
			want:   "POST ?",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, visionEndpoint(tc.uri, tc.method))
		})
	}
}
