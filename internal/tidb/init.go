package tidb

import (
	"database/sql"
	"fmt"

	log "github.com/sirupsen/logrus"
)

func InitDatabase(port uint, password string) error {
	log.Info("init database: trying login without password")

	db, err := sql.Open("mysql", fmt.Sprintf("root:@tcp(localhost:%d)/", port))

	defer db.Close()

	if err != nil {
		log.Debugf("init database: %s", err)
		log.Debug("init database: login as root with password")
		db, err = sql.Open("mysql", fmt.Sprintf("root:%s@tcp(localhost:%d)/", password, port))
	}

	if err != nil {
		log.Error(err)
		return err
	}

	log.Debug("init database: login was successful")

	_, err = db.Exec(fmt.Sprintf("SET PASSWORD FOR 'root'@'%%' = '%s'", password))

	if err != nil {
		log.Error(err)
	} else {
		log.Debug("init database: FLUSH PRIVILEGES")

		_, err = db.Exec("FLUSH PRIVILEGES")
	}

	log.Debug("init database: CREATE DATABASE IF NOT EXISTS photoprism")

	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS photoprism")

	if err != nil {
		log.Error(err)
	}

	return nil
}
