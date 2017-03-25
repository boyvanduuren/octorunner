package persist

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
