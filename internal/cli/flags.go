package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/antoniosarro/gofi/internal/config"
	"github.com/antoniosarro/gofi/internal/modules"
)

// ParseFlags parses command-line flags and returns options
func ParseFlags() *Options {
	opts := &Options{}

	flag.StringVar(&opts.Config, "config", config.GetConfigPath(), "Path to config file")
	flag.StringVar(&opts.Module, "m", "application", "Module to launch (application, screenshot, powermenu)")
	flag.BoolVar(&opts.EnablePagination, "pagination", false, "Enable pagination")
	flag.IntVar(&opts.ItemsPerPage, "items-per-page", 8, "Number of items per page")
	flag.BoolVar(&opts.EnableTags, "tags", false, "Show app type tags")
	flag.BoolVar(&opts.EnableHighlight, "highlight", false, "Highlight matching text in search results")
	flag.BoolVar(&opts.EnableFavorites, "favorites", false, "Enable favorites tracking")
	flag.BoolVar(&opts.ShowVersion, "version", false, "Show version information")
	flag.BoolVar(&opts.ListModules, "list-modules", false, "List available modules")

	flag.Usage = printUsage

	flag.Parse()

	return opts
}

// printUsage prints the usage information
func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: gofi [options]\n\n")
	fmt.Fprintf(os.Stderr, "A modular launcher for Linux\n\n")
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nAvailable modules:\n")
	for _, name := range modules.List() {
		if m, err := modules.Get(name); err == nil {
			fmt.Fprintf(os.Stderr, "  %-15s %s\n", name, m.Description())
		}
	}
	fmt.Fprintf(os.Stderr, "\nConfig file location: %s\n", config.GetConfigPath())
}
