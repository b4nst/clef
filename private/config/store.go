package config

import (
	"github.com/BurntSushi/toml"
	"github.com/b4nst/clef/private/backend"
)

type StoreDefinition struct {
	Type   string         `toml:"type"`
	Config toml.Primitive `toml:"config"`

	builder backend.Builder
}

func (sd *StoreDefinition) Builder() backend.Builder {
	return sd.builder
}
