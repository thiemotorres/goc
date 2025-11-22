package data

// Store handles ride persistence
type Store struct{}

// NewStore creates a new data store
func NewStore() *Store {
	return &Store{}
}
