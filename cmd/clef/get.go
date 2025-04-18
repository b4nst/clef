package main

import (
	"context"
	"fmt"

	"github.com/alecthomas/kong"

	"github.com/b4nst/clef/internal/config"
)

type Get struct {
	Store string `help:"Store to lookup from" short:"s" default:"default"`
	Key   string `arg:"" help:"Key to lookup"`
}

func (g *Get) Run(ctx context.Context, ktx *kong.Context, conf *config.Config) error {
	if conf == nil {
		return fmt.Errorf("unexpected nil config")
	}

	store, err := conf.Backend(ctx, g.Store)
	if err != nil {
		return fmt.Errorf("could not load store: %w", err)
	}

	v, err := store.Get(ctx, g.Key)
	if err != nil {
		return fmt.Errorf("error getting %s from %s store: %w", g.Key, g.Store, err)
	}

	fmt.Fprintln(ktx.Stdout, v)
	return nil
}
