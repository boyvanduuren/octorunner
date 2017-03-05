package persist

import (
	"database/sql"
	"os"
	"testing"
	"time"
)

var conn *sql.DB

const testDbName = "test.db"

func init() {
	// remove the test db, if it exists
	os.Remove(testDbName)
	OpenDatabase(testDbName, &conn)
}

func TestCreateProject(t *testing.T) {
	projectName := "TestCreateProject"
	projectOwner := "bcd"

	// Create a project
	id, err := createProject(projectName, projectOwner, conn)
	if err != nil {
		t.Fatal(err)
	}
	if id < 1 {
		t.Fatalf("Unexpected id: %d", id)
	}

	// Create it again, this should result in an error
	id, err = createProject(projectName, projectOwner, conn)
	if err == nil {
		t.Fatal("Expected an error, but didn't get one")
	}
	if id != -1 {
		t.Fatalf("Expected ID -1, got %d", id)
	}
}

func TestFindProject(t *testing.T) {
	projectName := "TestFindProject"
	projectOwner := "bcd"

	id := findProjectID(projectName, projectOwner, conn)
	if id != -1 {
		t.Fatal("Expected ID -1 when searching for non-existent project")
	}

	// Create the project
	createdID, err := createProject(projectName, projectOwner, conn)
	if err != nil {
		t.Fatal(err)
	}

	// Make sure the IDs match
	id = findProjectID(projectName, projectOwner, conn)
	if id != createdID {
		t.Fatal("Expected same IDs when searching for newly created project, got %d and %d", id, createdID)
	}

}

func TestCreateJob(t *testing.T) {
	// We should first create a project
	projectName := "TestCreateJob"
	projectOwner := "bcd"

	projectID, err := createProject(projectName, projectOwner, conn)
	if err != nil {
		t.Fatal(err)
	}

	// Create a job belonging to the project we just created
	jobCommitID := "deadbeef"
	jobName := "jobname"
	jobID, err := createJob(projectID, jobCommitID, jobName, conn)
	if err != nil {
		t.Fatal(err)
	}
	if jobID < 1 {
		t.Fatalf("Unexpected id: %d", jobID)
	}

	// Create a duplicate job (that is, the projectID, commitID and job's name already exist)
	_, err = createJob(projectID, jobCommitID, jobName, conn)
	if err == nil {
		t.Fatal("Expected error while creating duplicate job")
	}

	// Create a job for a projectID that doesn't exist, this should error
	_, err = createJob(projectID+1, "cafebabe", "jobname", conn)
	if err == nil {
		t.Fatal("Expected an error while creating a job for a projectID that doesn't exist")
	}

}
func TestFindJob(t *testing.T) {
	// We should first create a project
	projectName := "TestFindJob"
	projectOwner := "bcd"

	projectID, err := createProject(projectName, projectOwner, conn)
	if err != nil {
		t.Fatal(err)
	}

	// Search for a job that doesn't exist
	commitID := "deadc0de"
	jobName := "default"
	id := findJobID(projectID, commitID, jobName, conn)
	if id != -1 {
		t.Fatal("Expected ID -1 when searching for non-existent job")
	}

	// Create the project
	createdID, err := createJob(projectID, commitID, jobName, conn)
	if err != nil {
		t.Fatal(err)
	}

	// Make sure the IDs match
	id = findJobID(projectID, commitID, jobName, conn)
	if id != createdID {
		t.Fatal("Expected same IDs when searching for newly created job, got %d and %d", id, createdID)
	}

}

func TestCreateOutputWriter(t *testing.T) {
	projectName := "TestCreateOutputWriter"
	projectOwner := "bcd"
	commitID := "00bab10c"
	jobName := "default"

	writer, err := CreateOutputWriter(projectName, projectOwner, commitID, jobName, conn)
	if err != nil {
		t.Fatal(err)
	}

	// Make sure the project and job are created
	projectID := findProjectID(projectName, projectOwner, conn)
	if projectID == -1 {
		t.Fatal("Couldn't find project, even though it should have been created")
	}
	jobID := findJobID(projectID, commitID, jobName, conn)
	if jobID == -1 {
		t.Fatal("Couldn't find job, even though it should have been created")
	}

	timestampNow := time.Now()
	// Write a proper output tuple
	messageID, err := writer("message", timestampNow.Format(time.RFC3339))
	if err != nil {
		t.Fatal(err)
	}
	if messageID < 1 {
		t.Fatal("Got an unexpected message ID")
	}

	// Use an invalid time format
	_, err = writer("invalid time", timestampNow.Format(time.RFC1123))
	if err == nil {
		t.Fatal("Expected an error when using invalid timestamp")
	}

	// Creating a second writer on the same job should be OK
	_, err = CreateOutputWriter(projectName, projectOwner, commitID, jobName, conn)
	if err != nil {
		t.Fatal(err)
	}

}
