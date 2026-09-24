package polar

import (
	"context"
	"errors"
	"net/url"

	"github.com/moasq/go-b2b-starter/internal/modules/billing/domain"
	platform "github.com/moasq/go-b2b-starter/internal/platform/polar"
)

type Adapter struct{ client *platform.Client }

func NewPolarAdapter(client *platform.Client) domain.BillingProvider { return &Adapter{client: client} }
func (p *Adapter) ProductID() string                                 { return p.client.ProductID() }
func (p *Adapter) Enabled() bool                                     { return p.client.Enabled() }
func (p *Adapter) GetCustomerState(ctx context.Context, externalID string) (*domain.CustomerState, error) {
	state := &domain.CustomerState{}
	err := p.client.GetJSON(ctx, "/v1/customers/external/"+url.PathEscape(externalID)+"/state", state)
	if errors.Is(err, platform.ErrNotFound) {
		return &domain.CustomerState{ExternalID: externalID}, nil
	}
	return state, err
}
func (p *Adapter) GetCheckoutSession(ctx context.Context, id string) (*domain.CheckoutSession, error) {
	session := &domain.CheckoutSession{}
	err := p.client.GetJSON(ctx, "/v1/checkouts/"+url.PathEscape(id), session)
	return session, err
}
