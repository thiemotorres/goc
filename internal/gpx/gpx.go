package gpx

// Route represents a loaded GPX route
type Route struct{}

// Load parses a GPX file
func Load(path string) (*Route, error) {
	return &Route{}, nil
}
