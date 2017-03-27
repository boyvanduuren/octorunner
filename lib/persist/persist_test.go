package persist

import (
	"os"
	"testing"
	"time"
)

var conn DB

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
	id, err := conn.createProject(projectName, projectOwner)
	if err != nil {
		t.Fatal(err)
	}
	if id < 1 {
		t.Fatalf("Unexpected id: %d", id)
	}

	// Create it again, this should result in an error
	id, err = conn.createProject(projectName, projectOwner)
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

	id := conn.findProjectID(projectName, projectOwner)
	if id != -1 {
		t.Fatal("Expected ID -1 when searching for non-existent project")
	}

	// Create the project
	createdID, err := conn.createProject(projectName, projectOwner)
	if err != nil {
		t.Fatal(err)
	}

	// Make sure the IDs match
	id = conn.findProjectID(projectName, projectOwner)
	if id != createdID {
		t.Fatalf("Expected same IDs when searching for newly created project, got %d and %d", id, createdID)
	}

}

func TestCreateJob(t *testing.T) {
	// We should first create a project
	projectName := "TestCreateJob"
	projectOwner := "bcd"

	projectID, err := conn.createProject(projectName, projectOwner)
	if err != nil {
		t.Fatal(err)
	}

	// Create a job belonging to the project we just created
	jobCommitID := "deadbeef"
	jobName := "jobname"
	jobID, err := conn.createJob(projectID, jobCommitID, jobName)
	if err != nil {
		t.Fatal(err)
	}
	if jobID < 1 {
		t.Fatalf("Unexpected id: %d", jobID)
	}

	// Create a duplicate job (that is, the projectID, commitID and job's name already exist)
	// This shouldn't error, because a new iteration ID is generated for this job.
	_, err = conn.createJob(projectID, jobCommitID, jobName)
	if err != nil {
		t.Fatalf("Unexpected error while creating duplicate job: %q", err)
	}

	// Create a job for a projectID that doesn't exist, this should error
	_, err = conn.createJob(projectID+1, "cafebabe", "jobname")
	if err == nil {
		t.Fatal("Expected an error while creating a job for a projectID that doesn't exist")
	}

}
func TestFindJob(t *testing.T) {
	// We should first create a project
	projectName := "TestFindJob"
	projectOwner := "bcd"

	projectID, err := conn.createProject(projectName, projectOwner)
	if err != nil {
		t.Fatal(err)
	}

	// Search for a job that doesn't exist
	commitID := "deadc0de"
	jobName := "default"
	id := conn.findJobID(projectID, commitID, jobName, 1)
	if id != -1 {
		t.Fatal("Expected ID -1 when searching for non-existent job")
	}

	// Create the job
	createdID_01, err := conn.createJob(projectID, commitID, jobName)
	if err != nil {
		t.Fatal(err)
	}

	// Make sure the IDs match
	id = conn.findJobID(projectID, commitID, jobName, 1)
	if id != createdID_01 {
		t.Fatalf("Expected same IDs when searching for newly created job, got %d and %d", id, createdID_01)
	}

	// Create a new iteration of the job
	createdID_02, err := conn.createJob(projectID, commitID, jobName)
	if err != nil {
		t.Fatal(err)
	}

	id = conn.findJobID(projectID, commitID, jobName, 2)
	if id != createdID_02 {
		t.Fatalf("Expected same IDs when searching for newly created job, got %d and %d", id, createdID_02)
	}

	ids, err := conn.findJobIDs(projectID, commitID, jobName)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 {
		t.Fatalf("Error while retrieving created job IDs, expected array of size 2, got %d", len(ids))
	}
	if ids[0] != createdID_01 {
		t.Fatalf("Expected same IDs when searching for job IDs, got %d and %d", ids[0], createdID_01)
	}
	if ids[1] != createdID_02 {
		t.Fatalf("Expected same IDs when searching for job IDs, got %d and %d", ids[1], createdID_02)
	}

}

func TestCreateOutputWriter(t *testing.T) {
	projectName := "TestCreateOutputWriter"
	projectOwner := "bcd"
	commitID := "00bab10c"
	jobName := "default"

	writer, err := conn.CreateOutputWriter(projectName, projectOwner, commitID, jobName)
	if err != nil {
		t.Fatal(err)
	}

	// Make sure the project and job are created
	projectID := conn.findProjectID(projectName, projectOwner)
	if projectID == -1 {
		t.Fatal("Couldn't find project, even though it should have been created")
	}
	jobIDs, err := conn.findJobIDs(projectID, commitID, jobName)
	if len(jobIDs) == 0 {
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
	_, err = conn.CreateOutputWriter(projectName, projectOwner, commitID, jobName)
	if err != nil {
		t.Fatal(err)
	}

}
