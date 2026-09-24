package stytch

import (
	"errors"
	"fmt"
	"github.com/stytchauth/stytch-go/v18/stytch/stytcherror"
	"testing"
)

func TestSDKValueErrorsAreMapped(t *testing.T) {
	err := fmt.Errorf("provider call: %w", stytcherror.Error{StatusCode: 404})
	if !errors.Is(MapError(err), ErrNotFound) {
		t.Fatal("SDK 404 did not map to retry-safe not-found")
	}
}
