package persist

import "fmt"

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
