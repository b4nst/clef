package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"

	"github.com/alecthomas/kong"
	"github.com/b4nst/clef"
	"github.com/b4nst/clef/internal/config"
)

type Config struct {
	Editor string `help:"Edit the config with editor" short:"e" env:"EDITOR"`
}

func (g *Config) Run(ktx *kong.Context, cli *CLI) error {
	original, err := os.ReadFile(cli.ConfigFile)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("read config: %w", err)
	}

	content := original
	if len(content) <= 0 {
		content = clef.DefaultConfig
	}

	tmpf, err := os.CreateTemp("", "clef-config_*.toml")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer tmpf.Close()
	if _, err := tmpf.Write(content); err != nil {
		return fmt.Errorf("prepare temp file: %w", err)
	}
	if err := tmpf.Close(); err != nil {
		return fmt.Errorf("release temp file: %w", err)
	}

	if g.Editor == "" {
		g.Editor = "vi"
	}
	cmd := exec.Command(g.Editor, tmpf.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("edit error: %w", err)
	}

	new, err := os.ReadFile(tmpf.Name())
	if err != nil {
		return fmt.Errorf("read edited config: %w", err)
	}

	if bytes.Equal(new, original) {
		fmt.Fprintln(ktx.Stdout, "No change")
		return nil
	}

	// Test the new config
	if _, err := config.Parse(string(new)); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Write the new config
	if err := os.WriteFile(cli.ConfigFile, new, 0660); err != nil {
		return fmt.Errorf("saving new config: %w", err)
	}
	fmt.Printf("Config written to '%s'\n", cli.ConfigFile)

	return nil
}
