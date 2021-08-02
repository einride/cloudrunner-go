package cloudrunner

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"text/tabwriter"

	"go.einride.tech/cloudrunner/cloudconfig"
	"go.einride.tech/cloudrunner/cloudruntime"
)

func printUsage(w io.Writer, config *cloudconfig.Config) {
	_, _ = fmt.Fprintf(w, "\nUsage of %s:\n\n", path.Base(os.Args[0]))
	flag.CommandLine.PrintDefaults()
	_, _ = fmt.Fprintf(w, "\nRuntime configuration of %s:\n\n", path.Base(os.Args[0]))
	config.PrintUsage(w)
	_, _ = fmt.Fprintf(w, "\nBuild-time configuration of %s:\n\n", path.Base(os.Args[0]))
	tabs := tabwriter.NewWriter(w, 1, 0, 4, ' ', 0)
	_, _ = fmt.Fprintf(tabs, "LDFLAG\tTYPE\tVALUE\n")
	_, _ = fmt.Fprintf(
		tabs,
		"%v\t%v\t%v\n",
		"go.einride.tech/cloudrunner/cloudruntime.serviceVersion",
		"string",
		cloudruntime.ServiceVersionFromLinkerFlags(),
	)
	_ = tabs.Flush()
}
