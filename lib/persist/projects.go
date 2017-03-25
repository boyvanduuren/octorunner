package persist

import (
	"fmt"
)

// Project struct, fairly self-explanatory.
type Project struct {
	ID int64
	Name string
	Owner string
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

// FindAllProjects returns all the projects. Used for the webapi get "api/projects/".
func (db *DB) FindAllProjects() (*[]Project, error) {
	var results []Project

	rows, err := db.Connection.Query("SELECT id(), name, owner FROM Projects")

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var id int64
		var name, owner string

		rows.Scan(&id, &name, &owner)
		results = append(results, Project{ID: id, Name: name, Owner: owner})
	}

	return &results, nil
}

// FindProjectById find a single project by its ID. Used for the webapi get "api/projects/:ProjectID".
func (db *DB) FindProjectByID(id int64) (*Project, error) {
	var name, owner string

	row := db.Connection.QueryRow("SELECT name, owner FROM Projects WHERE id() = ?1", id)
	row.Scan(&name, &owner)

	if name == "" || owner == "" {
		return nil, fmt.Errorf("Couldn't find project with ID %q", id)
	}

	return &Project{
		ID: id,
		Name: name,
		Owner: owner,
	}, nil
}