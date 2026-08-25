package client

import (
	"strings"
	"testing"
)

// PickRegionOpenApiMeta is pure — these run offline, unlike the integration
// tests in relyt_client_test.go which hit a real deployment.

func metas(in ...*OpenApiMetaInfo) *[]*OpenApiMetaInfo {
	out := in
	return &out
}

func TestPickRegionOpenApiMeta_PrefersOpenapi(t *testing.T) {
	got, err := PickRegionOpenApiMeta("alibabacloud", "cn-hongkong", metas(
		&OpenApiMetaInfo{Type: "web_console", URI: "https://console.example.com"},
		&OpenApiMetaInfo{Type: "openapi", URI: "https://api.example.com"},
		&OpenApiMetaInfo{Type: "database", URI: "db.example.com:5432"},
	))
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got.URI != "https://api.example.com" {
		t.Fatalf("expected the openapi entry, got %q", got.URI)
	}
}

// The previous implementation rejected anything other than exactly one entry,
// so registering a second endpoint type broke every regional resource.
func TestPickRegionOpenApiMeta_MultipleEntriesAreFine(t *testing.T) {
	got, err := PickRegionOpenApiMeta("alibabacloud", "cn-hongkong", metas(
		&OpenApiMetaInfo{Type: "web_console", URI: "https://console.example.com"},
		&OpenApiMetaInfo{Type: "data_api", URI: "https://data.example.com"},
		&OpenApiMetaInfo{Type: "openapi", URI: "https://api.example.com"},
	))
	if err != nil {
		t.Fatalf("unexpected err with 3 entries: %v", err)
	}
	if got.Type != "openapi" {
		t.Fatalf("expected openapi, got %q", got.Type)
	}
}

func TestPickRegionOpenApiMeta_FallbackToWebConsole(t *testing.T) {
	got, err := PickRegionOpenApiMeta("alibabacloud", "cn-hongkong", metas(
		&OpenApiMetaInfo{Type: "database", URI: "jdbc:postgresql://x:5432"},
		&OpenApiMetaInfo{Type: "web_console", URI: "https://console.example.com/dms/12345/"},
	))
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got.Type != "web_console" {
		t.Fatalf("expected web_console fallback, got %q", got.Type)
	}
}

// The failure actually observed on the Alibaba Cloud deployment: the region had
// no endpoint registered at all.
func TestPickRegionOpenApiMeta_EmptyNamesTheRegion(t *testing.T) {
	_, err := PickRegionOpenApiMeta("alibabacloud", "cn-hongkong", metas())
	if err == nil {
		t.Fatal("expected an error for an empty endpoint list")
	}
	for _, want := range []string{"alibabacloud", "cn-hongkong"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error should name %q so the operator knows what to fix, got: %v", want, err)
		}
	}
}

func TestPickRegionOpenApiMeta_NilData(t *testing.T) {
	if _, err := PickRegionOpenApiMeta("alibabacloud", "cn-hongkong", nil); err == nil {
		t.Fatal("expected an error for nil data")
	}
}

// The pre-migration code returned the entry unseen whenever exactly one came
// back. AWS deployments are known to work that way and their endpoint type has
// never been observed, so a lone entry must keep winning whatever its type.
func TestPickRegionOpenApiMeta_SingleEntryOfAnyTypeWins(t *testing.T) {
	for _, typ := range []string{"", "something_new"} {
		got, err := PickRegionOpenApiMeta("aws", "us-east-1", metas(
			&OpenApiMetaInfo{Type: typ, URI: "https://api.example.com"},
		))
		if err != nil {
			t.Fatalf("type %q: unexpected err: %v", typ, err)
		}
		if got.URI != "https://api.example.com" {
			t.Fatalf("type %q: expected the single entry, got %q", typ, got.URI)
		}
	}
}

// Neither usable type present: the message must list what did come back, so the
// cause is visible without server-side access.
func TestPickRegionOpenApiMeta_UnusableTypesAreListed(t *testing.T) {
	_, err := PickRegionOpenApiMeta("alibabacloud", "cn-hongkong", metas(
		&OpenApiMetaInfo{Type: "database", URI: "jdbc:postgresql://x:5432"},
		&OpenApiMetaInfo{Type: "sso", URI: "https://sso.example.com"},
	))
	if err == nil {
		t.Fatal("expected an error when neither openapi nor web_console is present")
	}
	for _, want := range []string{"database", "sso"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error should list the types actually returned, missing %q: %v", want, err)
		}
	}
}
