package persist

import (
	"fmt"
	"time"
	log "github.com/Sirupsen/logrus"
)

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
