package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/b4nst/clef/private/backend"
)

type Set struct {
	Store string   `help:"Store to store to" short:"s" default:"default"`
	Key   string   `help:"Key to store to" short:"k" required:""`
	Value []string `arg:"" help:"Value to store."`
}

func (s *Set) Run(ctx context.Context, ktx *kong.Context, cli *CLI) error {
	store, err := backend.NewFileStore("./store")
	if err != nil {
		return fmt.Errorf("could not load store: %w", err)
	}

	v := strings.Join(s.Value, " ")
	if err := store.Set(ctx, s.Key, v); err != nil {
		return fmt.Errorf("error settings %s to %s store: %w", s.Key, s.Store, err)
	}

	fmt.Fprintln(ktx.Stdout, s.Key+" set")
	return nil
}
