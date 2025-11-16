package heroic

// Library represents the Heroic sideload_apps/library.json structure
type Library struct {
	Games []Game `json:"games"`
}

// Game represents a game entry in the Heroic library
type Game struct {
	Runner      string `json:"runner"`
	AppName     string `json:"app_name"`
	Title       string `json:"title"`
	FolderName  string `json:"folder_name"`
	ArtCover    string `json:"art_cover"`
	IsInstalled bool   `json:"is_installed"`
	ArtSquare   string `json:"art_square"`
}

// GameConfig represents the game configuration (we only need categories)
type GameConfig struct {
	Categories []string `json:"categories"`
}
