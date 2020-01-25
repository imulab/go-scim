package api

import "github.com/imulab/go-scim/v2/pkg/spec"

type applicationContext struct {
	args arguments

	// lazy init caches
	serviceProviderConfigCache *spec.ServiceProviderConfig
}

func (ctx applicationContext) MustServiceProviderConfig() *spec.ServiceProviderConfig {
	if ctx.serviceProviderConfigCache == nil {
		var err error
		ctx.serviceProviderConfigCache, err = ctx.args.ParseServiceProviderConfig()
		if err != nil {
			panic(err)
		}
	}
	return ctx.serviceProviderConfigCache
}
