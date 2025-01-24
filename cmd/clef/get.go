package main

import (
	"context"
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/b4nst/clef/private/backend"
)

type Get struct {
	Store string `help:"Store to lookup from" short:"s" default:"default"`
	Key   string `arg:"" help:"Key to lookup"`
}

func (g *Get) Run(ctx context.Context, ktx *kong.Context, cli *CLI) error {
	store, err := backend.NewFileStore("./store")
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
