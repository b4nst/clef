package main

import (
	"context"
	"fmt"

	"github.com/adrg/xdg"
	"github.com/alecthomas/kong"

	"github.com/b4nst/clef/internal/config"
)

type CLI struct {
	Get     Get     `cmd:"" help:"Lookup a key in a store."`
	Set     Set     `cmd:"" help:"Store a key value pair."`
	Version Version `cmd:"" help:"Print app version."`

	ConfigFile string `help:"Config file" short:"c"`
}

func main() {
	var cli CLI
	cmd := kong.Parse(&cli, kong.Name("clef"), kong.Description("Personal secret manager"))

	cmd.BindTo(context.Background(), (*context.Context)(nil))

	// ignore config errors for version command
	if cmd.Command() != "version" {
		// TODO: use kong resolvers for that.
		cpath := cli.ConfigFile
		if cpath == "" {
			var err error
			cpath, err = xdg.ConfigFile("clef/config.toml")
			cmd.FatalIfErrorf(err, "error getting default config file")
		}
		c, err := config.ParseFile(cpath)
		cmd.FatalIfErrorf(err, fmt.Sprintf("error reading config at %s", cpath))
		cmd.Bind(c)
	}

	cmd.FatalIfErrorf(cmd.Run())
}
