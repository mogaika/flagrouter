package provider

import (
	"fmt"

	"github.com/mogaika/flagrouter/router"
)

type Provider interface {
	Init(*router.Router) error
}

var (
	providers map[string]Provider = make(map[string]Provider)
)

func RegisterProvider(name string, pv Provider) {
	providers[name] = pv
}

func Providers() map[string]Provider {
	return providers
}

func InitProviders(r *router.Router) error {
	for name, p := range providers {
		if err := p.Init(r); err != nil {
			return fmt.Errorf("Error when Init '%s' provider: %v", name, err)
		}
	}
	return nil
}
