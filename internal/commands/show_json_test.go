package commands

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestShowThumbSizes_JSON(t *testing.T) {
	out, err := RunWithTestContext(ShowThumbSizesCommand, []string{"thumb-sizes", "--json"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var v []map[string]string

	if err = json.Unmarshal([]byte(out), &v); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out)
	}
	if len(v) == 0 {
		t.Fatalf("expected at least one item")
	}
	// Expected keys for thumb-sizes detailed report
	for _, k := range []string{"name", "width", "height", "aspect_ratio", "available", "usage"} {
		if _, ok := v[0][k]; !ok {
			t.Fatalf("expected key '%s' in first item", k)
		}
	}
}

func TestShowSources_JSON(t *testing.T) {
	out, err := RunWithTestContext(ShowSourcesCommand, []string{"sources", "--json"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var v []map[string]string

	if err = json.Unmarshal([]byte(out), &v); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out)
	}
	if len(v) == 0 {
		t.Fatalf("expected at least one item")
	}
	if _, ok := v[0]["source"]; !ok {
		t.Fatalf("expected key 'source' in first item")
	}
	if _, ok := v[0]["priority"]; !ok {
		t.Fatalf("expected key 'priority' in first item")
	}
}

func TestShowScopes_JSON(t *testing.T) {
	out, err := RunWithTestContext(ShowScopesCommand, []string{"scopes", "--json"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var v []map[string]string

	if err = json.Unmarshal([]byte(out), &v); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out)
	}
	if len(v) == 0 {
		t.Fatalf("expected at least one item")
	}
	if _, ok := v[0]["scope"]; !ok {
		t.Fatalf("expected key 'scope' in first item")
	}
	if _, ok := v[0]["description"]; !ok {
		t.Fatalf("expected key 'description' in first item")
	}
}

func TestShowMetadata_JSON(t *testing.T) {
	out, err := RunWithTestContext(ShowMetadataCommand, []string{"metadata", "--json"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var v struct {
		Items []map[string]string `json:"items"`
		Docs  []map[string]string `json:"docs"`
	}

	if err = json.Unmarshal([]byte(out), &v); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out)
	}

	if len(v.Items) == 0 {
		t.Fatalf("expected items")
	}
}

func TestShowConfig_JSON(t *testing.T) {
	out, err := RunWithTestContext(ShowConfigCommand, []string{"config", "--json"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var v struct {
		Sections []struct {
			Title string              `json:"title"`
			Items []map[string]string `json:"items"`
		} `json:"sections"`
	}

	if err = json.Unmarshal([]byte(out), &v); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out)
	}

	if len(v.Sections) == 0 || len(v.Sections[0].Items) == 0 {
		t.Fatalf("expected sections with items")
	}
}

func TestShowConfigOptions_JSON(t *testing.T) {
	out, err := RunWithTestContext(ShowConfigOptionsCommand, []string{"config-options", "--json"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	type options = []map[string]string
	var v = options{}

	if err = json.Unmarshal([]byte(out), &v); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out)
	}

	if len(v) == 0 {
		t.Fatalf("expected sections with items")
	}
}

func TestShowConfigYaml_JSON(t *testing.T) {
	out, err := RunWithTestContext(ShowConfigYamlCommand, []string{"config-yaml", "--json"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	type options = []map[string]string
	var v = options{}

	if err = json.Unmarshal([]byte(out), &v); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out)
	}

	if len(v) == 0 {
		t.Fatalf("expected sections with items")
	}
}

func TestShowFormatConflict_Error(t *testing.T) {
	out, err := RunWithTestContext(ShowSourcesCommand, []string{"sources", "--json", "--csv"})

	if err == nil {
		t.Fatalf("expected error for conflicting flags, got nil; output=%s", out)
	}

	// Expect an ExitCoder with code 2
	if ec, ok := err.(interface{ ExitCode() int }); ok {
		if ec.ExitCode() != 2 {
			t.Fatalf("expected exit code 2, got %d", ec.ExitCode())
		}
	} else {
		t.Fatalf("expected exit coder error, got %T: %v", err, err)
	}
}

func TestShowConfigOptions_MarkdownSections(t *testing.T) {
	out, err := RunWithTestContext(ShowConfigOptionsCommand, []string{"config-options", "--md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "### Authentication") {
		t.Fatalf("expected Markdown section heading '### Authentication' in output\n%s", out[:min(400, len(out))])
	}
}

func TestShowConfigOptions_MarkdownServicesCIDRSectionOrder(t *testing.T) {
	out, err := RunWithTestContext(ShowConfigOptionsCommand, []string{"config-options", "--md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	networkingHeading := strings.Index(out, "### Networking")
	proxyProtoHTTPS := strings.Index(out, "PHOTOPRISM_PROXY_PROTO_HTTPS")
	servicesCIDR := strings.Index(out, "PHOTOPRISM_SERVICES_CIDR")
	disableTLS := strings.Index(out, "PHOTOPRISM_DISABLE_TLS")

	if networkingHeading < 0 || proxyProtoHTTPS < 0 || servicesCIDR < 0 || disableTLS < 0 {
		t.Fatalf("expected networking heading and related rows in output")
	}

	if servicesCIDR < networkingHeading {
		t.Fatalf("expected services-cidr row to appear in Networking section")
	}

	if servicesCIDR < proxyProtoHTTPS {
		t.Fatalf("expected services-cidr row to appear after proxy-proto-https")
	}

	if servicesCIDR > disableTLS {
		t.Fatalf("expected services-cidr row to appear before disable-tls")
	}
}

func TestShowConfigYaml_MarkdownServicesCIDRSectionOrder(t *testing.T) {
	out, err := RunWithTestContext(ShowConfigYamlCommand, []string{"config-yaml", "--md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	networkingHeading := strings.Index(out, "### Networking")
	proxyProtoHTTPS := strings.Index(out, "ProxyProtoHttps")
	servicesCIDR := strings.Index(out, "ServicesCIDR")
	disableTLS := strings.Index(out, "DisableTLS")

	if networkingHeading < 0 || proxyProtoHTTPS < 0 || servicesCIDR < 0 || disableTLS < 0 {
		t.Fatalf("expected networking heading and related rows in output")
	}

	if servicesCIDR < networkingHeading {
		t.Fatalf("expected ServicesCIDR row to appear in Networking section")
	}

	if servicesCIDR < proxyProtoHTTPS {
		t.Fatalf("expected ServicesCIDR row to appear after ProxyProtoHttps")
	}

	if servicesCIDR > disableTLS {
		t.Fatalf("expected ServicesCIDR row to appear before DisableTLS")
	}
}

func TestShowConfigOptions_MarkdownHttpHeaderSettingsOrder(t *testing.T) {
	out, err := RunWithTestContext(ShowConfigOptionsCommand, []string{"config-options", "--md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	webServerHeading := strings.Index(out, "### Web Server")
	httpCompression := strings.Index(out, "PHOTOPRISM_HTTP_COMPRESSION")
	httpHeaderTimeout := strings.Index(out, "PHOTOPRISM_HTTP_HEADER_TIMEOUT")
	httpHeaderBytes := strings.Index(out, "PHOTOPRISM_HTTP_HEADER_BYTES")
	httpIdleTimeout := strings.Index(out, "PHOTOPRISM_HTTP_IDLE_TIMEOUT")
	httpCachePublic := strings.Index(out, "PHOTOPRISM_HTTP_CACHE_PUBLIC")

	if webServerHeading < 0 || httpCompression < 0 || httpHeaderTimeout < 0 || httpHeaderBytes < 0 || httpIdleTimeout < 0 || httpCachePublic < 0 {
		t.Fatalf("expected web server heading and related rows in output")
	}

	if httpHeaderTimeout < webServerHeading {
		t.Fatalf("expected http-header-timeout row to appear in Web Server section")
	}

	if httpHeaderTimeout < httpCompression {
		t.Fatalf("expected http-header-timeout row to appear after http-compression")
	}

	if httpHeaderBytes < httpHeaderTimeout {
		t.Fatalf("expected http-header-bytes row to appear after http-header-timeout")
	}

	if httpIdleTimeout < httpHeaderBytes {
		t.Fatalf("expected http-idle-timeout row to appear after http-header-bytes")
	}

	if httpIdleTimeout > httpCachePublic {
		t.Fatalf("expected http-idle-timeout row to appear before http-cache-public")
	}
}

func TestShowConfigYaml_MarkdownHttpHeaderSettingsOrder(t *testing.T) {
	out, err := RunWithTestContext(ShowConfigYamlCommand, []string{"config-yaml", "--md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	webServerHeading := strings.Index(out, "### Web Server")
	httpCompression := strings.Index(out, "HttpCompression")
	httpHeaderTimeout := strings.Index(out, "HttpHeaderTimeout")
	httpHeaderBytes := strings.Index(out, "HttpHeaderBytes")
	httpIdleTimeout := strings.Index(out, "HttpIdleTimeout")
	httpCachePublic := strings.Index(out, "HttpCachePublic")

	if webServerHeading < 0 || httpCompression < 0 || httpHeaderTimeout < 0 || httpHeaderBytes < 0 || httpIdleTimeout < 0 || httpCachePublic < 0 {
		t.Fatalf("expected web server heading and related rows in output")
	}

	if httpHeaderTimeout < webServerHeading {
		t.Fatalf("expected HttpHeaderTimeout row to appear in Web Server section")
	}

	if httpHeaderTimeout < httpCompression {
		t.Fatalf("expected HttpHeaderTimeout row to appear after HttpCompression")
	}

	if httpHeaderBytes < httpHeaderTimeout {
		t.Fatalf("expected HttpHeaderBytes row to appear after HttpHeaderTimeout")
	}

	if httpIdleTimeout < httpHeaderBytes {
		t.Fatalf("expected HttpIdleTimeout row to appear after HttpHeaderBytes")
	}

	if httpIdleTimeout > httpCachePublic {
		t.Fatalf("expected HttpIdleTimeout row to appear before HttpCachePublic")
	}
}

func TestShowFileFormats_JSON(t *testing.T) {
	out, err := RunWithTestContext(ShowFileFormatsCommand, []string{"file-formats", "--json"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var v []map[string]string

	if err = json.Unmarshal([]byte(out), &v); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out)
	}
	if len(v) == 0 {
		t.Fatalf("expected at least one item")
	}
	// Keys depend on report settings in command: should include format, description, type, extensions
	if _, ok := v[0]["format"]; !ok {
		t.Fatalf("expected key 'format' in first item")
	}
	if _, ok := v[0]["type"]; !ok {
		t.Fatalf("expected key 'type' in first item")
	}
	if _, ok := v[0]["extensions"]; !ok {
		t.Fatalf("expected key 'extensions' in first item")
	}
}

func TestShowVideoSizes_JSON(t *testing.T) {
	out, err := RunWithTestContext(ShowVideoSizesCommand, []string{"video-sizes", "--json"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var v []map[string]string

	if err = json.Unmarshal([]byte(out), &v); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out)
	}
	if len(v) == 0 {
		t.Fatalf("expected at least one item")
	}
	if _, ok := v[0]["size"]; !ok {
		t.Fatalf("expected key 'size' in first item")
	}
	if _, ok := v[0]["usage"]; !ok {
		t.Fatalf("expected key 'usage' in first item")
	}
}
