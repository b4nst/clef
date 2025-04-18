package backend

import (
	"context"
	"fmt"
)

var builderRegistry = make(map[string]BuilderFunc)

type BuilderFunc func() Builder

type Builder interface {
	Build(ctx context.Context, name string) (Store, error)
}

func registerBuilder(name string, builderf func() Builder) {
	if _, exists := builderRegistry[name]; exists {
		panic(fmt.Sprintf("A builder of type %s is already registered", name))
	}
	builderRegistry[name] = builderf
}

func BuilderOf(t string) (Builder, error) {
	if bf, ok := builderRegistry[t]; ok {
		return bf(), nil
	}
	return nil, fmt.Errorf("store has unsupported type '%s'", t)
}
