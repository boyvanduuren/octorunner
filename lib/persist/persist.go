package persist

import (
	"database/sql"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/cznic/ql"
	"time"
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

func (db *DB) findProjectID(name string, owner string) int64 {
	var id *int64
	_ = db.Connection.QueryRow("SELECT id() FROM Projects WHERE name = ?1 "+
		"AND owner = ?2", name, owner).Scan(&id)

	if id == nil {
		return -1
	}
	return *id
}

func (db *DB) createProject(name string, owner string) (int64, error) {
	tx, err := db.Connection.Begin()
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

func (db *DB) findJobID(projectID int64, commitID string, job string) int64 {
	var id *int64
	_ = db.Connection.QueryRow("SELECT id() FROM Jobs WHERE project = ?1 "+
		"AND commitID = ?2 AND job = ?3", projectID, commitID, job).Scan(&id)

	if id == nil {
		return -1
	}
	return *id
}

func (db *DB) createJob(projectID int64, commitID string, job string) (int64, error) {
	// Make sure we refer to an existing project
	rows, err := db.Connection.Query("SELECT id() FROM Projects WHERE id() = ?1 ", projectID)
	if err != nil || !rows.Next() {
		return -1, fmt.Errorf("Cannot create job for project with ID %d as it doesn't exist", projectID)
	}

	tx, err := db.Connection.Begin()
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

func (db *DB) createOutput(jobID int64, data, date string) (int64, error) {
	// Make sure we refer to an existing project
	rows, err := db.Connection.Query("SELECT id() FROM Jobs WHERE id() = ?1", jobID)
	if err != nil || !rows.Next() {
		return -1, fmt.Errorf("Cannot add output to job with ID %d as it doesn't exist", jobID)
	}

	tx, err := db.Connection.Begin()
	if err != nil {
		return -1, err
	}

	q := fmt.Sprintf("INSERT INTO Output VALUES (?1, ?2, parseTime(%q, ?3))", time.RFC3339)
	res, err := tx.Exec(q, jobID, data, date)
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

/*
CreateOutputWriter returns a function that can be used to write to the "Output" table. That table is used
to write test output to. Test output belongs to a certain job, which in turn belongs to a project. Before
returning the writer we make sure these tuples exist or are created.
 */
func (db *DB) CreateOutputWriter(projectName string, projectOwner string, commitID string,
		job string) (func(string, string) (int64, error), error) {
	var err error
	// Get the ID of this project
	log.Debugf("Querying for project ID of project with name %q and owner %q", projectName, projectOwner)
	projectID := db.findProjectID(projectName, projectOwner)

	if projectID == -1 {
		// Project isn't known in the database, so create it
		log.Debugf("Project not found, creating")
		projectID, err = db.createProject(projectName, projectOwner)
		if err != nil {
			return nil, err
		}
	}
	log.Debugf("Project has ID %d", projectID)

	jobID := db.findJobID(projectID, commitID, job)
	if jobID == -1 {
		// Job isn't known in the database, so create it
		log.Debugf("Job not found, creating")
		jobID, err = db.createJob(projectID, commitID, job)
		if err != nil {
			return nil, err
		}
	}
	log.Debugf("Job has ID %d", jobID)

	log.Debugf("Returning a writer for project %d, job %d", projectID, jobID)
	return func(line, date string) (int64, error) {
		outputID, err := db.createOutput(jobID, line, date)
		if err != nil {
			return -1, err
		}
		return outputID, nil
	}, nil
}
