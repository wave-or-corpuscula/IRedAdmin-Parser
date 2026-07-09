// Package synccommon is used for common tools in sync process
package synccommon

import "iredparser/internal/database"

type SyncSession struct {
	ServerID int64
}

func GetTestDB() (*database.Database, error) {
	return database.Connect(":memory:")
	// return database.Connect("../../../../data/ireddata.db")
}
