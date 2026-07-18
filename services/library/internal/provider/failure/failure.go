// Package failure defines the provider-neutral error vocabulary shared by
// low-level clients and their domain adapters.
package failure

import "errors"

// Kind is a stable failure category, independent of any provider protocol.
type Kind string

const (
	Invalid      Kind = "invalid"
	Unauthorized Kind = "unauthorized"
	Forbidden    Kind = "forbidden"
	NotFound     Kind = "not_found"
	RateLimited  Kind = "rate_limited"
	Unavailable  Kind = "unavailable"
)

// Classified is implemented by provider failures that can cross the plugin boundary.
type Classified interface {
	error
	FailureKind() Kind
}

// Classify finds a provider failure through an error wrapping chain.
func Classify(err error) (Kind, bool) {
	var classified Classified
	if !errors.As(err, &classified) {
		return "", false
	}
	return classified.FailureKind(), true
}
