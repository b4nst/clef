package main

import (
	"github.com/alecthomas/kong"
	"github.com/b4nst/clef/cmd/clef/subcmd"
)

type CLI struct {
	Version subcmd.Version `cmd:"" help:"Print app version."`
}

func main() {
	var cli CLI
	cmd := kong.Parse(&cli, kong.Name("clef"), kong.Description("Personal secret manager"))

	cmd.FatalIfErrorf(cmd.Run())
}
