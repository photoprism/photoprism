package tidb

import (
	"database/sql"
	"fmt"

	log "github.com/sirupsen/logrus"
)

func InitDatabase(port uint, password string) error {
	log.Print("init database: trying login without password")

	db, err := sql.Open("mysql", fmt.Sprintf("root:@tcp(localhost:%d)/", port))

	defer db.Close()

	if err != nil {
		log.Print(err)
		log.Print("init database: login as root with password")
		db, err = sql.Open("mysql", fmt.Sprintf("root:%s@tcp(localhost:%d)/", password, port))
	}

	if err != nil {
		log.Print(err)
		return err
	}

	log.Print("init database: login was successful")

	_, err = db.Exec(fmt.Sprintf("SET PASSWORD FOR 'root'@'%%' = '%s'", password))

	if err != nil {
		log.Print(err)
	} else {
		log.Print("init database: FLUSH PRIVILEGES")

		_, err = db.Exec("FLUSH PRIVILEGES")

		log.Print(err)

	}

	log.Printf("init database: CREATE DATABASE IF NOT EXISTS photoprism")

	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS photoprism")

	if err != nil {
		log.Print(err)
	}

	return nil
}
