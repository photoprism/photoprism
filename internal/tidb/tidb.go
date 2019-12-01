package tidb

import (
	"database/sql"
	"fmt"
)

func InitDatabase(port uint, password string) error {
	log.Info("init database")

	db, err := sql.Open("mysql", fmt.Sprintf("root:@tcp(localhost:%d)/", port))

	defer db.Close()

	if err != nil {
		log.Debug(err.Error())
		log.Debug("database login as root with password")
		db, err = sql.Open("mysql", fmt.Sprintf("root:%s@tcp(localhost:%d)/", password, port))
	}

	if err != nil {
		log.Error(err.Error())
		return err
	}

	if password != "" {
		log.Debug("set database password")

		_, err = db.Exec(fmt.Sprintf("SET PASSWORD FOR 'root'@'%%' = '%s'", password))

		if err != nil {
			log.Error(err.Error())
		} else {
			log.Debug("flush database privileges")

			_, err = db.Exec("FLUSH PRIVILEGES")
		}
	}

	log.Debug("create database if not exists")

	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS photoprism")

	if err != nil {
		log.Error(err.Error())
	}

	log.Info("database created")

	return nil
}
