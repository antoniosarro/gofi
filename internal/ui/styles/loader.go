package styles

import (
	"os"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// Load loads CSS styles with priority levels
func Load(defaultCSS, globalCSS, moduleCSS string) error {
	// Load default CSS
	if err := loadCSS(defaultCSS, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION); err != nil {
		return err
	}

	// Load global CSS if specified
	if globalCSS != "" {
		if _, err := os.Stat(globalCSS); err == nil {
			if err := loadCSS(globalCSS, gtk.STYLE_PROVIDER_PRIORITY_USER); err != nil {
				return err
			}
		}
	}

	// Load module-specific CSS if specified
	if moduleCSS != "" {
		if _, err := os.Stat(moduleCSS); err == nil {
			if err := loadCSS(moduleCSS, gtk.STYLE_PROVIDER_PRIORITY_USER+1); err != nil {
				return err
			}
		}
	}

	return nil
}

// loadCSS loads a CSS file with the specified priority
func loadCSS(path string, priority uint) error {
	cssProvider := gtk.NewCSSProvider()
	cssProvider.LoadFromPath(path)

	gtk.StyleContextAddProviderForDisplay(
		gdk.DisplayGetDefault(),
		cssProvider,
		priority,
	)

	return nil
}
