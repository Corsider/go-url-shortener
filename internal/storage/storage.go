package storage

// Server storage interface
type UrlStorage interface {
	Save(original string, id int) error
	Load(id int) (string, error)
	GetLastId() (int, error)
	CheckExistence(original string) (bool, int)
	Close() error
}
