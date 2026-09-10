package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/pkg/cfg"
)

// TestShippingMethodsYAML guards the shipping.methods field names against
// aconfig's map-to-struct matching, which only title-cases the first letter of
// a YAML key and ignores `yaml:` tags for structs nested in a slice — so every
// method field must be a single word.
func TestShippingMethodsYAML(t *testing.T) {
	yaml := `
shipping:
  methods:
    - { code: pickup, title: "Самовывоз", kind: self_pickup, free: true }
    - { code: cdek, title: "СДЭК", kind: pickup, provider: cdek, tariff: "136", markup: "10" }
`
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}

	// satisfy the `required` config validations we don't care about here
	for k, v := range map[string]string{
		"POSTGRES_ADDR": "x", "POSTGRES_DB_NAME": "x", "POSTGRES_USER": "x",
		"POSTGRES_PASSWORD": "x", "REDIS_ADDR": "x",
	} {
		t.Setenv(k, v)
	}

	conf := new(config.Config)
	if err := cfg.Load(conf, cfg.WithLoaderConfig(cfg.Config{
		Args: []string{}, SkipFlags: true, Files: []string{path}, MergeFiles: true,
	})); err != nil {
		t.Fatalf("load config: %v", err)
	}

	if len(conf.Shipping.Methods) != 2 {
		t.Fatalf("methods: %+v", conf.Shipping.Methods)
	}
	self, cdek := conf.Shipping.Methods[0], conf.Shipping.Methods[1]
	if self.Code != "pickup" || self.Kind != "self_pickup" || !self.Free {
		t.Fatalf("self_pickup: %+v", self)
	}
	if cdek.Provider != "cdek" || cdek.Tariff != "136" || cdek.Markup != "10" {
		t.Fatalf("cdek: %+v", cdek)
	}
}
