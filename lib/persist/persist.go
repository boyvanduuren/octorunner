package persist

import (
	"github.com/cznic/ql"
	"database/sql"
	log "github.com/Sirupsen/logrus"
	"fmt"
)

var connection *sql.DB

func OpenDatabase(path string) error {
	ql.RegisterDriver()
	db, err := sql.Open("ql", path)
	if err != nil {
		return err
	}
	connection = db
	log.Infof("Connected to database %q", path)

	err = initializeDatabase()
	if err != nil {
		return err
	}

	return nil
}

func initializeDatabase() error {
	log.Debug("Starting transaction")
	tx, err := connection.Begin()
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
		"CREATE UNIQUE INDEX IF NOT EXISTS JobsProjectCommit ON Jobs (project, commitID)",
		"CREATE UNIQUE INDEX IF NOT EXISTS OutputID ON Output (id())",
		"CREATE UNIQUE INDEX IF NOT EXISTS OutputJob ON Output (job)",
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

func findProjectID(name string, owner string) (int, error) {
	var id *int

	//tx, err := connection.Begin()
	//if err != nil {
	//	return -1, err
	//}

	_ = connection.QueryRow("SELECT id() FROM Projects WHERE name = ? AND owner = ?", name, owner)
	//_ = connection.QueryRow("SELECT id() FROM Projects WHERE name = \"projectName\" AND owner = \"projectOwner\"").Scan(&id)
	//if err != nil {
	//	return -1, err
	//}

	if id == nil {
		return -1, nil
	}

	//tx.Commit()

	return *id, nil
}

func createProject(name string, owner string) error {
	tx, err := connection.Begin()
	if err != nil {
		return err
	}

	tx.Exec("INSERT INTO Projects VALUES (\"?\", \"?\")", name, owner)
	tx.Commit()

	return nil
}

func WriteOutput(projectName string, projectOwner string, commitID string, job string, data string) {
	// Get the ID of this project
	findProjectID(projectName, projectOwner)
}