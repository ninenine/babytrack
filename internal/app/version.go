package app

var version = "dev"

// SetVersion sets the application version (called from main)
func SetVersion(v string) {
	version = v
}

// GetVersion returns the application version
func GetVersion() string {
	return version
}
