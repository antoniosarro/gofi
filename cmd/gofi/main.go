package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/antoniosarro/gofi/internal/cli"
	"github.com/antoniosarro/gofi/internal/config"
	"github.com/antoniosarro/gofi/internal/modules"
	"github.com/antoniosarro/gofi/internal/ui/styles"

	// Import modules to trigger init() registration
	_ "github.com/antoniosarro/gofi/internal/modules/application"
	_ "github.com/antoniosarro/gofi/internal/modules/powermenu"
	_ "github.com/antoniosarro/gofi/internal/modules/screenshot"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

const (
	appID   = "com.github.antoniosarro.gofi"
	version = "0.1.0"
)

var (
	activeModule modules.Module
	activeWindow modules.Window
)

func main() {

	// Parse command-line flags
	opts := cli.ParseFlags()

	if opts.ShowVersion {
		fmt.Printf("gofi version %s\n", version)
		os.Exit(0)
	}

	if opts.ListModules {
		fmt.Println("Available modules:")
		for _, name := range modules.List() {
			if m, err := modules.Get(name); err == nil {
				fmt.Printf("  %-15s %s\n", name, m.Description())
			}
		}
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load(opts.Config)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Merge CLI options with config
	moduleConfig := opts.MergeWithConfig(cfg)

	// Get the requested module
	module, err := modules.Get(opts.Module)
	if err != nil {
		log.Fatalf("Error: %v\nUse -list-modules to see available modules", err)
	}

	// Check if module is enabled
	if !moduleConfig.Enabled {
		log.Fatalf("Module '%s' is disabled in config", opts.Module)
	}

	activeModule = module

	// Suppress GTK/GDK debug output
	if os.Getenv("DEBUG") != "1" {
		log.SetOutput(io.Discard)
	} else {
		log.SetOutput(os.Stderr)
	}

	// Set up signal handling
	setupSignalHandler()

	app := gtk.NewApplication(appID, gio.ApplicationFlagsNone)

	app.ConnectActivate(func() {
		// Load CSS using styles loader
		// Expand paths for tilde support
		globalCSS := config.ExpandPath(cfg.GlobalCSS)
		moduleCSS := config.ExpandPath(moduleConfig.CustomCSS)

		if err := styles.Load("assets/style.css", globalCSS, moduleCSS); err != nil {
			log.Printf("Warning: Failed to load CSS: %v", err)
		}

		// Initialize module
		if err := module.Initialize(moduleConfig); err != nil {
			log.Fatalf("Error initializing module '%s': %v", module.Name(), err)
		}

		// Create window
		window, err := module.CreateWindow(app)
		if err != nil {
			log.Fatalf("Error creating window for module '%s': %v", module.Name(), err)
		}
		activeWindow = window

		// Connect window close event
		window.Widget().ConnectCloseRequest(func() bool {
			window.Shutdown()
			return false
		})

		window.Show()
	})

	if code := app.Run([]string{os.Args[0]}); code > 0 {
		cleanup()
		os.Exit(code)
	}

	cleanup()
}

func setupSignalHandler() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("Received interrupt signal, cleaning up...")
		cleanup()
		os.Exit(0)
	}()
}

func cleanup() {
	if activeModule != nil {
		if err := activeModule.Cleanup(); err != nil {
			log.Printf("Error during cleanup: %v", err)
		}
	}
}
