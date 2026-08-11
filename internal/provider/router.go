package provider

import (
	"errors"
	"fmt"

	"github.com/ismetkoralay/heimdall/internal/config"
)

var (
	ErrUnknownModel    = errors.New("error unknown model")
	ErrUnknownProvider = errors.New("error unknown provider")
)

type Router struct {
	models    map[string]config.ModelProviderConfig
	providers map[string]Provider
}

func NewRouter(
	models map[string]config.ModelProviderConfig,
	providers map[string]Provider,
) *Router {
	return &Router{
		models:    models,
		providers: providers,
	}
}

func (r *Router) Resolve(model string) (Provider, error) {
	m, ok := r.models[model]
	if !ok {
		return nil, ErrUnknownModel
	}

	p, ok := r.providers[m.ProviderName]
	if !ok {
		return nil, fmt.Errorf("%w: expected provider: %s for model: %s", ErrUnknownProvider, m.ProviderName, model)
	}

	return p, nil
}
