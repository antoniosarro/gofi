package cli

// Options represents command-line options
type Options struct {
	Config           string
	Module           string
	EnablePagination bool
	ItemsPerPage     int
	EnableTags       bool
	EnableHighlight  bool
	EnableFavorites  bool
	ShowVersion      bool
	ListModules      bool
}
