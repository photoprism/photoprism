package clean

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDNSLabel(t *testing.T) {
	cases := []struct {
		in   string
		want string
		name string
	}{
		{" Client Credentials幸", "client-credentials", "basic normalization"},
		{" My.Host/Name:Prod ", "my-host-name-prod", "separators to dash"},
		{"a---b___c   d", "a-b-c-d", "collapse dashes"},
		{"-._a--", "a", "trim leading trailing"},
		{strings.Repeat("a", 40), strings.Repeat("a", 32), "clip length"},
		{"!!!", "", "all invalid"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, DNSLabel(tc.in))
		})
	}
}
