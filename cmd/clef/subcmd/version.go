package subcmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/alecthomas/kong"
)

type Version struct {
	Short bool `help:"Only print semver version." short:"s"`
}

var (
	// Version string to be injected at runtime
	version = "devel"
	commit  = "HEAD"
	date    = "NaN"
	builtBy = "manual"
)

func (v *Version) Run(ktx *kong.Context) error {
	if v.Short {
		_, err := fmt.Fprintln(ktx.Stdout, version)
		if err != nil {
			return err
		}
		return nil
	}

	w := tabwriter.NewWriter(ktx.Stdout, 0, 0, 3, ' ', 0)
	format := "Version:\t%s\nCommit:\t%s\nDate:\t%s\nBuilt by:\t%s\n"
	_, err := fmt.Fprintf(w, format, version, commit, date, builtBy)
	if err != nil {
		return err
	}
	return w.Flush()
}
