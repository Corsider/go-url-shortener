package storage

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"go-url-shortener/internal/storage/inmemory"
	"go-url-shortener/internal/storage/postgres"
	"testing"
)

// Testing in-memory storage functions
func TestInMemoryStorage(t *testing.T) {
	memory := inmemory.New()
	err := memory.Save("original.com", 0)
	assert.NoError(t, err)

	result, err := memory.Load(0)
	assert.NoError(t, err)
	assert.Equal(t, "original.com", result)

	num, err := memory.GetLastId()
	assert.NoError(t, err)
	assert.Equal(t, 1, num)

	ok, index := memory.CheckExistence("not_there.com")
	assert.Equal(t, false, ok)
	assert.Equal(t, index, 0)

	ok, index = memory.CheckExistence("original.com")
	assert.Equal(t, true, ok)
	assert.Equal(t, index, 0)
}

// Testing Postgres storage functions
func TestPostgresStorage_Save(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	postgresStorage := &postgres.UPostgresStorage{DB: db}
	mock.ExpectExec("INSERT INTO urls").WithArgs("original.com").WillReturnResult(sqlmock.NewResult(1, 1))
	err := postgresStorage.Save("original.com", 1)
	assert.NoError(t, err)
}

func TestPostgresStorage_Load(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	postgresStorage := &postgres.UPostgresStorage{DB: db}
	rows := sqlmock.NewRows([]string{"original"}).AddRow("original.com")
	mock.ExpectQuery("^SELECT original FROM urls WHERE id=\\$1$").WithArgs(2).WillReturnRows(rows)
	res, err := postgresStorage.Load(1)

	assert.NoError(t, err)
	assert.Equal(t, "original.com", res)
}

func TestPostgresStorage_GetLastId(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	postgresStorage := &postgres.UPostgresStorage{DB: db}
	rows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery("^SELECT count\\(\\*\\) FROM urls$").WillReturnRows(rows)
	res, err := postgresStorage.GetLastId()

	assert.NoError(t, err)
	assert.Equal(t, 1, res)
}

func TestPostgresStorage_CheckExistence(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	postgresStorage := &postgres.UPostgresStorage{DB: db}
	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery("^SELECT id FROM urls WHERE original=\\$1$").WithArgs("original.com").WillReturnRows(rows)
	exists, id := postgresStorage.CheckExistence("original.com")

	assert.True(t, exists)
	assert.Equal(t, 0, id)
}

func TestPostgresStorage_Close(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	postgresStorage := &postgres.UPostgresStorage{DB: db}
	mock.ExpectClose()
	err := postgresStorage.Close()
	assert.NoError(t, err)
}
