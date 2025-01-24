package main

import (
	"context"

	"github.com/alecthomas/kong"
)

type CLI struct {
	Get     Get     `cmd:"" help:"Lookup a key in a store."`
	Set     Set     `cmd:"" help:"Store a key value pair."`
	Version Version `cmd:"" help:"Print app version."`
}

func main() {
	var cli CLI
	cmd := kong.Parse(&cli, kong.Name("clef"), kong.Description("Personal secret manager"))

	cmd.BindTo(context.Background(), (*context.Context)(nil))

	cmd.FatalIfErrorf(cmd.Run())
}
