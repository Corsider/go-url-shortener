package postgres

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
)

type UPostgresStorage struct {
	DB *sql.DB
}

func New(connection string) *UPostgresStorage {
	database, err := sql.Open("postgres", connection)
	if err != nil {
		log.Fatal(err)
	}
	err = database.Ping()
	if err != nil {
		log.Fatal(err)
	}
	return &UPostgresStorage{DB: database}
}

func (p *UPostgresStorage) Save(original string, id int) error {
	_, err := p.DB.Exec("INSERT INTO urls (original) VALUES ($1)", original)
	return err
}

func (p *UPostgresStorage) Load(id int) (string, error) {
	var original string
	err := p.DB.QueryRow("SELECT original FROM urls WHERE id=$1", id+1).Scan(&original)
	return original, err
}

func (p *UPostgresStorage) GetLastId() (int, error) {
	var res int
	err := p.DB.QueryRow("SELECT count(*) FROM urls").Scan(&res)
	return res, err
}

func (p *UPostgresStorage) CheckExistence(original string) (bool, int) {
	var id int
	err := p.DB.QueryRow("SELECT id FROM urls WHERE original=$1", original).Scan(&id)
	if err != nil {
		return false, 0
	}
	return true, id - 1
}

func (p *UPostgresStorage) Close() error {
	return p.DB.Close()
}
