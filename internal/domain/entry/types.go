package entry

// AppType represents the source/type of the application
type AppType string

const (
	AppTypeAll       AppType = "All"
	AppTypeSystem    AppType = "System"
	AppTypeNixSystem AppType = "Nix-Sys"
	AppTypeNixHome   AppType = "Nix-Home"
	AppTypeFlatpak   AppType = "Flatpak"
	AppTypeGame      AppType = "Games"
	AppTypeOther     AppType = "Other"
)

// String returns the string representation of AppType
func (a AppType) String() string {
	return string(a)
}

// IsValid checks if the AppType is valid
func (a AppType) IsValid() bool {
	switch a {
	case AppTypeAll, AppTypeSystem, AppTypeNixSystem,
		AppTypeNixHome, AppTypeFlatpak, AppTypeGame, AppTypeOther:
		return true
	}
	return false
}
