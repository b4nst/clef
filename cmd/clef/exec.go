package main

import (
	"context"
	"fmt"

	"github.com/b4nst/clef/internal/config"
	"github.com/b4nst/clef/internal/profile"
)

type Exec struct {
	Profile string           `help:"Profile to load." short:"p" optional:""`
	Secret  []profile.Secret `help:"Secrets to load into the env. Format [store.]secret[=env]. If store is empty, default store will be used. If env is empty, secret name will be used as env name." short:"s" optional:""`

	Args []string `arg:""`
}

func (s *Exec) Run(ctx context.Context, conf *config.Config) error {
	if conf == nil {
		return fmt.Errorf("unexpected nil config")
	}

	prof := &profile.Profile{}
	// Load a profile only if explicitly requested, or no secret requested
	if s.Profile != "" || len(s.Secret) <= 0 {
		var err error
		prof, err = conf.Profile(s.Profile)
		if err != nil {
			return fmt.Errorf("get profile: %w", err)
		}
	}

	return prof.Exec(ctx, s.Args, conf, s.Secret...)
}
