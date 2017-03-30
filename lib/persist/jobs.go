package persist

import (
	"database/sql"
	"fmt"
)

type Job struct {
	ID        int64
	Iteration int64
	Project   int64
	CommitID  string
	Job       string
	Status    string
	Extra     string
	Data      []*Output
}

type JobStatus int

const (
	STATUS_DONE JobStatus = iota
	STATUS_RUNNING
	STATUS_ERROR
)

func statusToString(status JobStatus) string {
	var statusText string
	switch status {
	case STATUS_DONE:
		statusText = "done"
	case STATUS_RUNNING:
		statusText = "running"
	case STATUS_ERROR:
		statusText = "error"
	}
	return statusText
}

func (db *DB) findJobID(projectID int64, commitID string, job string, iteration int64) int64 {
	var id *int64
	_ = db.Connection.QueryRow("SELECT id() FROM Jobs WHERE project = ?1 "+
		"AND commitID = ?2 AND job = ?3 AND iteration = ?4",
		projectID, commitID, job, iteration).Scan(&id)

	if id == nil {
		return -1
	}
	return *id
}

func (db *DB) findJobIDs(projectID int64, commitID string, job string) ([]int64, error) {
	var IDs []int64
	rows, err := db.Connection.Query("SELECT id() FROM Jobs WHERE project = ?1 "+
		"AND commitID = ?2 AND job = ?3 ORDER BY id() ASC",
		projectID, commitID, job)
	if err != nil {
		return []int64{}, err
	}

	for rows.Next() {
		var id int64
		rows.Scan(&id)
		IDs = append(IDs, id)
	}

	return IDs, nil
}

func (db *DB) createJob(projectID int64, commitID string, job string) (int64, error) {
	// Make sure we refer to an existing project
	rows, err := db.Connection.Query("SELECT id() FROM Projects WHERE id() = ?1 ", projectID)
	if err != nil || !rows.Next() {
		return -1, fmt.Errorf("Cannot create job for project with ID %d as it doesn't exist", projectID)
	}

	// Retrieve the latest iteration ID of this job, which might not exist
	var latestJobIteration int64
	row := db.Connection.QueryRow("SELECT iteration FROM Jobs WHERE project = ?1", projectID)
	err = row.Scan(&latestJobIteration)
	if err == sql.ErrNoRows {
		latestJobIteration = 0
	} else if err != nil {
		return -1, err
	}

	tx, err := db.Connection.Begin()
	if err != nil {
		return -1, err
	}

	res, err := tx.Exec("INSERT INTO Jobs (project, commitID, job, status, iteration)"+
		" VALUES (?1, ?2, ?3, ?4, ?5)", projectID, commitID, job, "running", latestJobIteration+1)
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

// UpdateJobStatus sets the status of a job and allows for some extra information to be passed as string.
func (db *DB) UpdateJobStatus(jobID int64, status JobStatus, extra string) error {
	tx, err := db.Connection.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("UPDATE Jobs SET status = ?1, extra = ?2 WHERE id() = ?3",
		statusToString(status), extra, jobID)
	tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// Find all jobs that belong to a specific project. This doesn't query the data belonging to every job.
// Used for the webapi get "api/projects/:ProjectID/jobs".
func (db *DB) FindJobsForProject(projectID int64) ([]Job, error) {
	var jobs []Job

	rows, err := db.Connection.Query("SELECT id(), iteration, commitID, job, status, extra "+
		"FROM Jobs WHERE project = ?1", projectID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var id, iteration int64
		var commitID, job, status, extra string

		rows.Scan(&id, &iteration, &commitID, &job, &status, &extra)
		jobs = append(jobs, Job{
			ID:        id,
			Iteration: iteration,
			Project:   projectID,
			CommitID:  commitID,
			Job:       job,
			Data:      nil,
			Status:    status,
			Extra:     extra,
		})
	}

	return jobs, nil
}

// FindJobWithData finds a job and returns it, with all the
// Output data related to it already fetched.
func (db *DB) FindJobWithData(jobID int64) (*Job, error) {
	var iteration int64
	var commitID, job, status, extra string

	row := db.Connection.QueryRow("SELECT iteration, commitID, job, status, extra FROM Jobs WHERE id() = ?1", jobID)
	row.Scan(&iteration, &commitID, &job, &status, &extra)

	if commitID == "" {
		return nil, fmt.Errorf("Couldn't find project with ID %q", jobID)
	}

	data, err := db.findAllOutputForJob(jobID)
	if err != nil {
		return nil, err
	}

	return &Job{
		ID:        jobID,
		Iteration: iteration,
		Project:   jobID,
		CommitID:  commitID,
		Job:       job,
		Status:    status,
		Extra:     extra,
		Data:      data,
	}, nil
}

// GetLatestJobID returns the latest jobID. The return value will be < 1 if no job was found.
func (db *DB) GetLatestJobID() int64 {
	var jobID int64

	row := db.Connection.QueryRow("SELECT id() FROM Jobs ORDER BY id() DESC LIMIT 1")
	row.Scan(&jobID)

	return jobID
}
