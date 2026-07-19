package authn

import (
	"encoding/base64"
	"strings"
	"testing"
)

func unsignedToken(header string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(header)) + ".payload.signature"
}

func TestAccessTokenType(t *testing.T) {
	accepted := unsignedToken(`{"alg":"RS256","typ":"at+jwt"}`)
	if !hasAccessTokenType(accepted) {
		t.Fatal("RFC 9068 access-token type was rejected")
	}
	for name, token := range map[string]string{
		"ID token":       unsignedToken(`{"alg":"RS256","typ":"JWT"}`),
		"missing type":   unsignedToken(`{"alg":"RS256"}`),
		"malformed JSON": unsignedToken(`{"typ":`),
		"two segments":   "header.payload",
		"four segments":  "header.payload.signature.extra",
	} {
		if hasAccessTokenType(token) {
			t.Errorf("%s was accepted as an access token", name)
		}
	}
}

func TestBearerTokenIsBounded(t *testing.T) {
	if token, ok := bearerToken("Bearer compact.jwt.token"); !ok || token != "compact.jwt.token" {
		t.Fatal("valid bearer token was rejected")
	}
	if _, ok := bearerToken("Bearer " + strings.Repeat("a", maxAccessTokenLength+1)); ok {
		t.Fatal("oversized bearer token was accepted")
	}
}
