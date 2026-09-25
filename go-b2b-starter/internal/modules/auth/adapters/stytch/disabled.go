package stytch

import (
	"context"
	"github.com/moasq/go-b2b-starter/internal/modules/auth"
)

// DisabledAuthProvider rejects all sessions until a real provider is configured.
type DisabledAuthProvider struct{}

func (DisabledAuthProvider) VerifyToken(context.Context, string) (*auth.Identity, error) {
	return nil, auth.ErrInvalidToken
}
