package registry

import (
	"micro-scheduler/internal/conf"
	"time"

	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/hashicorp/consul/api"
	"github.com/google/wire"
)

// ProviderSet is registry providers.
var ProviderSet = wire.NewSet(NewRegistry)

// NewRegistry 创建服务注册器
func NewRegistry(c *conf.Registry) (registry.Registrar, registry.Discovery, error) {
	consulConfig := api.DefaultConfig()
	consulConfig.Address = c.Consul.Address
	if c.Consul.Scheme != "" {
		consulConfig.Scheme = c.Consul.Scheme
	}

	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		return nil, nil, err
	}

	r := consul.New(consulClient)
	return r, r, nil
}

// RegisterOption 注册选项
type RegisterOption struct {
	TTL time.Duration
}

// DefaultRegisterOption 默认注册选项
func DefaultRegisterOption() *RegisterOption {
	return &RegisterOption{
		TTL: 30 * time.Second,
	}
}