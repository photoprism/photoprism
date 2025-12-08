package vision

import "testing"

func TestServiceEndpoint(t *testing.T) {
	tests := []struct {
		name       string
		svc        Service
		wantURI    string
		wantMethod string
	}{
		{
			name:       "Disabled",
			svc:        Service{Disabled: true, Uri: "https://vision.example.com"},
			wantURI:    "",
			wantMethod: "",
		},
		{
			name:       "WithBasicAuth",
			svc:        Service{Uri: "https://vision.example.com/api", Username: "user", Password: "secret"},
			wantURI:    "https://user:secret@vision.example.com/api",
			wantMethod: ServiceMethod,
		},
		{
			name:       "UsernameOnly",
			svc:        Service{Uri: "https://vision.example.com/", Username: "scoped"},
			wantURI:    "https://scoped@vision.example.com/",
			wantMethod: ServiceMethod,
		},
		{
			name:       "PreserveExistingUser",
			svc:        Service{Uri: "https://keep:me@vision.example.com", Username: "ignored", Password: "ignored"},
			wantURI:    "https://keep:me@vision.example.com",
			wantMethod: ServiceMethod,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uri, method := tt.svc.Endpoint()
			if uri != tt.wantURI {
				t.Fatalf("uri: got %q want %q", uri, tt.wantURI)
			}
			if method != tt.wantMethod {
				t.Fatalf("method: got %q want %q", method, tt.wantMethod)
			}
		})
	}
}

func TestServiceCredentialsAndHeaders(t *testing.T) {
	t.Setenv("VISION_USER", "alice")
	t.Setenv("VISION_PASS", "hunter2")
	t.Setenv("VISION_MODEL", "GEMMA3:Latest")
	t.Setenv("VISION_ORG", "org-123")
	t.Setenv("VISION_PROJECT", "proj-abc")

	svc := Service{
		Username: "${VISION_USER}",
		Password: "${VISION_PASS}",
		Model:    "${VISION_MODEL}",
		Org:      "${VISION_ORG}",
		Project:  "${VISION_PROJECT}",
	}

	user, pass := svc.BasicAuth()
	if user != "alice" || pass != "hunter2" {
		t.Fatalf("basic auth: got %q/%q", user, pass)
	}

	if got := svc.GetModel(); got != "gemma3:latest" {
		t.Fatalf("model override: got %q", got)
	}

	if got := svc.EndpointOrg(); got != "org-123" {
		t.Fatalf("org: got %q", got)
	}

	if got := svc.EndpointProject(); got != "proj-abc" {
		t.Fatalf("project: got %q", got)
	}
}
