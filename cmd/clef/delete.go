package main

import (
	"context"
	"fmt"

	"github.com/alecthomas/kong"

	"github.com/b4nst/clef/internal/config"
)

type Delete struct {
	Store string `help:"Store to lookup from" short:"s" default:"default"`
	Key   string `arg:"" help:"Key to lookup"`
}

func (g *Delete) Run(ctx context.Context, ktx *kong.Context, conf *config.Config) error {
	if conf == nil {
		return fmt.Errorf("unexpected nil config")
	}

	store, err := conf.Backend(g.Store)
	if err != nil {
		return fmt.Errorf("load store: %w", err)
	}

	if err := store.Delete(ctx, g.Key); err != nil {
		return fmt.Errorf("delete %s from %s store: %w", g.Key, g.Store, err)
	}

	fmt.Fprintln(ktx.Stdout, g.Key, "deleted")

	return nil
}
