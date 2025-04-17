package main

import (
	"context"
	"fmt"

	"github.com/adrg/xdg"
	"github.com/alecthomas/kong"

	"github.com/b4nst/clef/internal/config"
)

type CLI struct {
	Get     Get     `cmd:"" help:"Lookup a key in a store." aliases:"fetch"`
	Set     Set     `cmd:"" help:"Store a key value pair." aliases:"put, store"`
	Delete  Delete  `cmd:"" help:"Delete a key from a store." aliases:"rm"`
	Version Version `cmd:"" help:"Print app version."`
	Config  Config  `cmd:"" help:"Manage clef configuration."`
	Shell   Shell   `cmd:"" help:"Load a shell with secrets injected as env variable."`
	Exec    Exec    `cmd:"" help:"Execute a command with secrets injected as env variable."`

	ConfigFile string `help:"Config file" short:"c" default:"${config_file}"`
}

func ConfigProvider(cli *CLI) (*config.Config, error) {
	conf, err := config.ParseFile(cli.ConfigFile)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	return conf, nil
}

func main() {
	xpath, _ := xdg.ConfigFile("clef/config.toml")
	var cli CLI
	cmd := kong.Parse(&cli,
		kong.Name("clef"),
		kong.Description("Personal secret manager"),
		kong.Vars{"config_file": xpath},
		kong.BindToProvider(ConfigProvider),
		kong.BindTo(context.Background(), (*context.Context)(nil)),
	)

	cmd.FatalIfErrorf(cmd.Run())
}
