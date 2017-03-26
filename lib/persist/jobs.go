package persist

import "fmt"

type Job struct {
	ID int64
	Project int64
	CommitID string
	Job string
	Data []*Output
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

// Find all jobs that belong to a specific project. This doesn't query the data belonging to every job.
// Used for the webapi get "api/projects/:ProjectID/jobs".
func (db *DB) FindJobsForProject(projectID int64) ([]Job, error) {
	var jobs []Job

	rows, err := db.Connection.Query("SELECT id(), commitID, job FROM Jobs WHERE project = ?1", projectID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var id int64
		var commitID, job string

		rows.Scan(&id, &commitID, &job)
		jobs = append(jobs, Job{
			ID: id,
			Project: projectID,
			CommitID: commitID,
			Job: job,
			Data: nil,
		})
	}

	return jobs, nil
}
