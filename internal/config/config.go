package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/b4nst/clef/internal/backend"
)

type Config struct {
	DefaultStore string                      `toml:"default_store"`
	Stores       map[string]*StoreDefinition `toml:"stores"`
}

func parse(decoder *toml.Decoder) (*Config, error) {
	config := new(Config)
	md, err := decoder.Decode(config)
	if err != nil {
		return nil, err
	}
	for name, def := range config.Stores {
		if def.Type == "" {
			return nil, fmt.Errorf("missing type for store %s", name)
		}
		b, err := backend.BuilderOf(def.Type)
		if err != nil {
			return nil, fmt.Errorf("failed to get builder for store %s: %w", name, err)
		}
		if err := md.PrimitiveDecode(def.Config, b); err != nil {
			return nil, err
		}
		def.builder = b
	}

	return config, nil
}

func Parse(conf string) (*Config, error) {
	return parse(toml.NewDecoder(strings.NewReader(conf)))
}

func ParseFile(path string) (*Config, error) {
	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	return parse(toml.NewDecoder(fp))
}

func (c *Config) Backend(name string) (backend.Store, error) {
	// System store is a special OSStore used to store system secrets
	if name == backend.SystemStoreNameSpace {
		return backend.SystemStore, nil
	}

	if name == "" || name == "default" {
		name = c.DefaultStore
	}

	def, ok := c.Stores[name]
	if !ok {
		return nil, fmt.Errorf("%s store not found in configuration", name)
	}

	return def.builder.Build(name)
}
