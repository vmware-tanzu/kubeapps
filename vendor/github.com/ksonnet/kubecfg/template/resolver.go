package template

import (
	"fmt"
	"net/http"

	"github.com/ksonnet/kubecfg/utils"
	log "github.com/sirupsen/logrus"
)

func (spec *Expander) buildResolver() (utils.Resolver, error) {
	ret := resolverErrorWrapper{}

	switch spec.FailAction {
	case "ignore":
		ret.OnErr = func(error) error { return nil }
	case "warn":
		ret.OnErr = func(err error) error {
			log.Warning(err.Error())
			return nil
		}
	case "error":
		ret.OnErr = func(err error) error { return err }
	default:
		return nil, fmt.Errorf("Unknown resolve failure type: %s", spec.FailAction)
	}

	switch spec.Resolver {
	case "noop":
		ret.Inner = utils.NewIdentityResolver()
	case "registry":
		ret.Inner = utils.NewRegistryResolver(&http.Client{
			Transport: utils.NewAuthTransport(http.DefaultTransport),
		})
	default:
		return nil, fmt.Errorf("Unknown resolver type: %s", spec.Resolver)
	}

	return &ret, nil
}

type resolverErrorWrapper struct {
	Inner utils.Resolver
	OnErr func(error) error
}

func (r *resolverErrorWrapper) Resolve(image *utils.ImageName) error {
	err := r.Inner.Resolve(image)
	if err != nil {
		err = r.OnErr(err)
	}
	return err
}
