package tidb

import (
	"database/sql"
	"fmt"
)

func InitDatabase(port uint) error {
	db, err := sql.Open("mysql", fmt.Sprintf("root:@tcp(localhost:%d)/", port))

	if err != nil {
		return err
	}

	defer db.Close()

	_, err = db.Exec("CREATE DATABASE photoprism")

	if err != nil {
		return err
	}

	return nil
}
