package main

import (
	"context"
	"fmt"

	"github.com/b4nst/clef/internal/config"
)

type Shell struct {
	Profile string `help:"Profile to load." short:"p" default:"default"`
	Shell   string `help:"Shell to use" short:"s" env:"SHELL"`
}

func (s *Shell) Run(ctx context.Context, conf *config.Config) error {
	const defaultShell = "sh"

	if conf == nil {
		return fmt.Errorf("unexpected nil config")
	}

	profile, err := conf.Profile(s.Profile)
	if err != nil {
		return fmt.Errorf("get profile: %w", err)
	}

	return profile.Activate(ctx, s.Shell, conf)
}
