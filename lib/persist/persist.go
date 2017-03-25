package persist

import (
	"database/sql"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/cznic/ql"
)

type DB struct {
	Connection *sql.DB
}

var DBConn DB

/*
OpenDatabase opens a QL embedded database connection to a db at a certain path.
 */
func OpenDatabase(path string, connectionPtr *DB) error {
	ql.RegisterDriver()
	db, err := sql.Open("ql", path)
	if err != nil {
		return err
	}
	connectionPtr.Connection = db
	log.Infof("Connected to database %q", path)

	err = connectionPtr.initializeDatabase()
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) initializeDatabase() error {
	log.Debug("Starting transaction")
	tx, err := db.Connection.Begin()
	if err != nil {
		return nil
	}

	creationQueries := []string{
		"CREATE TABLE IF NOT EXISTS Projects (name string, owner string)",
		"CREATE TABLE IF NOT EXISTS Jobs (project int, commitID string, job string)",
		"CREATE TABLE IF NOT EXISTS Output (job int, data string, timestamp time)",
		"CREATE UNIQUE INDEX IF NOT EXISTS ProjectsID ON Projects (id())",
		"CREATE UNIQUE INDEX IF NOT EXISTS ProjectRepository ON Projects (name, owner)",
		"CREATE UNIQUE INDEX IF NOT EXISTS JobsID ON Jobs (id())",
		"CREATE UNIQUE INDEX IF NOT EXISTS JobsProjectCommit ON Jobs (project, commitID, job)",
		"CREATE UNIQUE INDEX IF NOT EXISTS OutputID ON Output (id())",
		"CREATE INDEX IF NOT EXISTS OutputJob ON Output (job)",
	}

	for _, q := range creationQueries {
		log.Debugf("Executing query %q", q)
		_, err := tx.Exec(q)
		if err != nil {
			return fmt.Errorf("Error on query %q: %q", q, err)
		}
	}

	err = tx.Commit()
	log.Debug("Transaction committed")
	log.Info("Initialized database")

	return err
}

