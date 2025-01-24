package main

import (
	"fmt"
	"text/tabwriter"

	"github.com/alecthomas/kong"
	"github.com/b4nst/clef/cmd/clef/version"
)

type Version struct {
	Short bool `help:"Only print semver version." short:"s"`
}

func (v *Version) Run(ktx *kong.Context) error {
	if v.Short {
		_, err := fmt.Fprintln(ktx.Stdout, version.Version)
		if err != nil {
			return err
		}
		return nil
	}

	w := tabwriter.NewWriter(ktx.Stdout, 0, 0, 3, ' ', 0)
	format := "Version:\t%s\nCommit:\t%s\nDate:\t%s\nBuilt by:\t%s\n"
	_, err := fmt.Fprintf(w, format, version.Version, version.Commit, version.Date, version.BuiltBy)
	if err != nil {
		return err
	}
	return w.Flush()
}
