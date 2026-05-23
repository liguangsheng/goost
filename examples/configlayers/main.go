// configlayers demonstrates env + defaultmap: load base settings from an
// environment-shaped map, then lazily derive per-tenant settings on demand.
//
// Run from examples/: go run ./configlayers
package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/liguangsheng/goost/defaultmap"
	"github.com/liguangsheng/goost/env"
)

type Config struct {
	Region       string        `env:"REGION,default=us-east-1"`
	DefaultPlan  string        `env:"DEFAULT_PLAN,default=free"`
	Timeout      time.Duration `env:"TIMEOUT,default=250ms"`
	TenantAdmins []string      `env:"TENANT_ADMINS"`
}

type TenantConfig struct {
	Name    string
	Region  string
	Plan    string
	Admin   string
	Timeout time.Duration
}

func main() {
	var cfg Config
	if err := env.LoadFromMap(&cfg, map[string]string{
		"REGION":        "eu-west-1",
		"DEFAULT_PLAN":  "standard",
		"TIMEOUT":       "500ms",
		"TENANT_ADMINS": "alpha:alice,beta:bob",
	}); err != nil {
		panic(err)
	}

	admins := parseTenantAdmins(cfg.TenantAdmins)
	tenants := defaultmap.Make(func(name string) TenantConfig {
		return TenantConfig{
			Name:    name,
			Region:  cfg.Region,
			Plan:    cfg.DefaultPlan,
			Admin:   admins.Get(name),
			Timeout: cfg.Timeout,
		}
	})

	alpha := tenants.Get("alpha")
	beta := tenants.Get("beta")
	preview, loaded := tenants.GetOrInit("preview")

	fmt.Printf("alpha=%s/%s admin=%s timeout=%s\n", alpha.Region, alpha.Plan, alpha.Admin, alpha.Timeout)
	fmt.Printf("beta=%s/%s admin=%s timeout=%s\n", beta.Region, beta.Plan, beta.Admin, beta.Timeout)
	fmt.Printf("preview admin=%q loaded=%v total=%d\n", preview.Admin, loaded, tenants.Len())
}

func parseTenantAdmins(values []string) *defaultmap.Map[string, string] {
	admins := defaultmap.Make(func(string) string { return "" })
	for _, value := range values {
		name, admin, ok := strings.Cut(value, ":")
		if ok {
			admins.Set(name, admin)
		}
	}
	return admins
}
