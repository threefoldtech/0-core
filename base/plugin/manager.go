package plugin

//Manager a plugin manager
type Manager struct {
	path string
}

//New create a new plugin manager
func New(path string) (*Manager, error) {
	return &Manager{path}, nil
}

// func Route(name string)
