package gateway

import (
	"net/url"
	"testing"
)

func TestValidateRequiresBothTokens(t *testing.T) {
	cases := []struct {
		name    string
		ingest  map[string]struct{}
		query   map[string]struct{}
		wantErr bool
	}{
		{name: "both set", ingest: tokenSet("write"), query: tokenSet("read"), wantErr: false},
		{name: "no query token", ingest: tokenSet("write"), query: nil, wantErr: true},
		{name: "no ingest token", ingest: nil, query: tokenSet("read"), wantErr: true},
		{name: "neither", ingest: nil, query: nil, wantErr: true},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			config := Config{IngestTokens: testCase.ingest, QueryTokens: testCase.query}
			if err := config.Validate(); (err != nil) != testCase.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, testCase.wantErr)
			}
		})
	}
}

// The read token must never authorize a write. This is the invariant that the
// removed queryTokens=ingestTokens fallback silently broke.
func TestReadTokenIsNotAWriteToken(t *testing.T) {
	config := Config{IngestTokens: tokenSet("write"), QueryTokens: tokenSet("read")}

	if config.AuthorizedQuery("Bearer read") != true {
		t.Error("query token must authorize reads")
	}
	if config.AuthorizedQuery("Bearer write") != false {
		t.Error("ingest token must not authorize reads")
	}
	if authorizedWith(config.IngestTokens, "Bearer read") != false {
		t.Error("query token must not authorize writes")
	}
	if authorizedWith(config.IngestTokens, "Bearer write") != true {
		t.Error("ingest token must authorize writes")
	}
}

func TestAuthorizedWith(t *testing.T) {
	tokens := tokenSet("alpha", "beta")
	cases := []struct {
		header string
		want   bool
	}{
		{header: "Bearer alpha", want: true},
		{header: "Bearer beta", want: true},
		{header: "Bearer  alpha ", want: true}, // surrounding space is trimmed
		{header: "Bearer gamma", want: false},
		{header: "alpha", want: false},        // missing scheme
		{header: "bearer alpha", want: false}, // scheme is case-sensitive
		{header: "Basic alpha", want: false},
		{header: "", want: false},
		{header: "Bearer ", want: false},
	}
	for _, testCase := range cases {
		if got := authorizedWith(tokens, testCase.header); got != testCase.want {
			t.Errorf("authorizedWith(%q) = %v, want %v", testCase.header, got, testCase.want)
		}
	}
}

// An empty token set rejects every caller rather than admitting them.
func TestAuthorizedWithEmptySetRejects(t *testing.T) {
	if authorizedWith(map[string]struct{}{}, "Bearer anything") {
		t.Fatal("an empty token set must reject every caller")
	}
}

func TestParseTokens(t *testing.T) {
	tokens := parseTokens(" alpha , beta ,, ")
	if len(tokens) != 2 {
		t.Fatalf("parseTokens returned %d tokens, want 2", len(tokens))
	}
	for _, want := range []string{"alpha", "beta"} {
		if _, ok := tokens[want]; !ok {
			t.Errorf("parseTokens dropped %q", want)
		}
	}
	if len(parseTokens("")) != 0 {
		t.Error("an empty value must yield no tokens")
	}
}

func TestPrometheusParamsAllowList(t *testing.T) {
	allowed := []string{"query", "time"}

	params, ok := prometheusParams(url.Values{
		"query":   {"up"},
		"time":    {"123"},
		"evil":    {"dropped"},
		"timeout": {"also dropped"},
	}, allowed)
	if !ok {
		t.Fatal("a well-formed query must be accepted")
	}
	if params.Get("query") != "up" || params.Get("time") != "123" {
		t.Errorf("allowed params were not forwarded: %v", params)
	}
	if params.Has("evil") || params.Has("timeout") {
		t.Errorf("params outside the allow-list were forwarded: %v", params)
	}
}

func TestPrometheusParamsRejectsBadQueries(t *testing.T) {
	allowed := []string{"query", "time"}
	cases := map[string]url.Values{
		"missing query":  {"time": {"1"}},
		"empty query":    {"query": {""}},
		"oversize query": {"query": {longString(4097)}},
		"oversize param": {"query": {"up"}, "time": {longString(513)}},
	}
	for name, values := range cases {
		if _, ok := prometheusParams(values, allowed); ok {
			t.Errorf("%s: expected rejection", name)
		}
	}
}

func tokenSet(values ...string) map[string]struct{} {
	if len(values) == 0 {
		return nil
	}
	tokens := make(map[string]struct{}, len(values))
	for _, value := range values {
		tokens[value] = struct{}{}
	}
	return tokens
}

func longString(size int) string {
	buffer := make([]byte, size)
	for index := range buffer {
		buffer[index] = 'a'
	}
	return string(buffer)
}
