package database

import (
	"fmt"
	"iredparser/internal/parser"
)

func GetTestDB(dsn string) (*Database, error) {
	return Connect(dsn)
}

func GetTestServer(num int) *parser.Server {
	return &parser.Server{
		Name: fmt.Sprintf("Server %d", num),
	}
}

func GetTestDBWithServer(num int) (*Database, *ServerModel, error) {
	db, err := GetTestDB(":memory:")
	if err != nil {
		return nil, nil, err
	}
	server := GetTestServer(num)

	model, err := db.UpsertServer(server)
	if err != nil {
		return nil, nil, err
	}

	return db, model, nil
}
