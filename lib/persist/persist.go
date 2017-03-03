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

func findProjectID(name string, owner string) (int64, error) {
	var id *int64
	_ = connection.QueryRow("SELECT id() FROM Projects WHERE name = ?1 " +
		"AND owner = ?2", name, owner).Scan(&id)

	if id == nil {
		return -1, nil
	}
	return *id, nil
}

func createProject(name string, owner string) (int64, error) {
	tx, err := connection.Begin()
	if err != nil {
		return -1, err
	}

	res, err := tx.Exec("INSERT INTO Projects VALUES (?1, ?2)", name, owner)
	tx.Commit()
	if err != nil {
		 return -1, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}

	return id, nil
}

func findJobID(projectID int64, commitID string, job string) (int64, error) {
	var id *int64
	_ = connection.QueryRow("SELECT id() FROM Jobs WHERE project = ?1 " +
		"AND commitID = ?2 AND job = ?3", projectID, commitID, job).Scan(&id)

	if id == nil {
		return -1, nil
	}
	return *id, nil
}

func createJob(projectID int64, commitID string, job string) (int64, error) {
	// Make sure we refer to an existing project
	rows, err := connection.Query("SELECT id() FROM Projects WHERE id() = ?1 ", projectID)
	if err != nil || !rows.Next() {
		return -1, fmt.Errorf("Cannot create job for project with ID %d as it doesn't exist", projectID)
	}

	tx, err := connection.Begin()
	if err != nil {
		return -1, err
	}

	res, err := tx.Exec("INSERT INTO Jobs VALUES (?1, ?2, ?3)", projectID, commitID, job)
	tx.Commit()
	if err != nil {
		return -1, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}

	return id, nil
}

func createOutput(jobID int64, data string) (int64, error) {
	// Make sure we refer to an existing project
	rows, err := connection.Query("SELECT id() FROM Jobs WHERE id() = ?1", jobID)
	if err != nil || !rows.Next() {
		return -1, fmt.Errorf("Cannot add output to job with ID %d as it doesn't exist", jobID)
	}


	tx, err := connection.Begin()
	if err != nil {
		return -1, err
	}

	res, err := tx.Exec("INSERT INTO Output VALUES (?1, ?2, now())", jobID, data)
	tx.Commit()
	if err != nil {
		return -1, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}

	return id, nil
}

func WriteOutput(projectName string, projectOwner string, commitID string, job string, data string) error {
	// Get the ID of this project
	log.Debugf("Querying for project ID of project with name %q and owner %q", projectName, projectOwner)
	projectID, err := findProjectID(projectName, projectOwner)
	if err != nil {
		return err
	}

	if projectID == -1 {
		// Project isn't known in the database, so create it
		log.Debugf("Project not found, creating")
		projectID, err = createProject(projectName, projectOwner)
		if err != nil {
			return err
		}
	}
	log.Debugf("Project has ID %d", projectID)

	jobID, err := findJobID(projectID, commitID, job)
	if jobID == -1 {
		// Job isn't known in the database, so create it
		log.Debugf("Job not found, creating")
		jobID, err = createJob(projectID, commitID, job)
		if err != nil {
			return err
		}
	}
	log.Debugf("Job has ID %d", jobID)

	outputID, err := createOutput(jobID, "foobar")
	if err != nil {
		return err
	}
	log.Debugf("Added output %d", outputID)

	return nil
}