// Code generated by goagen v1.1.0-dirty, command line:
// $ goagen
// --design=github.com/boyvanduuren/octorunner/lib/webapi/design
// --out=$(GOPATH)\src\github.com\boyvanduuren\octorunner\lib\webapi
// --version=v1.1.0-dirty
//
// API "octorunner": Application Media Types
//
// The content of this file is auto-generated, DO NOT MODIFY

package app

import (
	"github.com/goadesign/goa"
	"time"
)

// A job that was ran after a commit on a project (default view)
//
// Identifier: application/vnd.octorunner.job+json; view=default
type OctorunnerJob struct {
	// The git commit ID specific to this job
	CommitID string              `form:"commitID" json:"commitID" xml:"commitID"`
	Data     []*OctorunnerOutput `form:"data,omitempty" json:"data,omitempty" xml:"data,omitempty"`
	// Extra information, this might contain error information
	Extra string `form:"extra" json:"extra" xml:"extra"`
	// Unique job ID
	ID int `form:"id" json:"id" xml:"id"`
	// The iteration ID of this job. A job might be ran multiple times.
	Iteration int `form:"iteration" json:"iteration" xml:"iteration"`
	// The name of the job
	Job string `form:"job" json:"job" xml:"job"`
	// The project this job belongs to
	Project int `form:"project" json:"project" xml:"project"`
	// The status of the job
	Status string `form:"status" json:"status" xml:"status"`
}

// Validate validates the OctorunnerJob media type instance.
func (mt *OctorunnerJob) Validate() (err error) {

	if mt.CommitID == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "commitID"))
	}
	if mt.Job == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "job"))
	}

	if mt.Status == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "status"))
	}
	if mt.Extra == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "extra"))
	}
	if !(mt.Status == "running" || mt.Status == "done" || mt.Status == "error") {
		err = goa.MergeErrors(err, goa.InvalidEnumValueError(`response.status`, mt.Status, []interface{}{"running", "done", "error"}))
	}
	return
}

// A job that was ran after a commit on a project (light view)
//
// Identifier: application/vnd.octorunner.job+json; view=light
type OctorunnerJobLight struct {
	// The git commit ID specific to this job
	CommitID string `form:"commitID" json:"commitID" xml:"commitID"`
	// Extra information, this might contain error information
	Extra string `form:"extra" json:"extra" xml:"extra"`
	// Unique job ID
	ID int `form:"id" json:"id" xml:"id"`
	// The iteration ID of this job. A job might be ran multiple times.
	Iteration int `form:"iteration" json:"iteration" xml:"iteration"`
	// The name of the job
	Job string `form:"job" json:"job" xml:"job"`
	// The project this job belongs to
	Project int `form:"project" json:"project" xml:"project"`
	// The status of the job
	Status string `form:"status" json:"status" xml:"status"`
}

// Validate validates the OctorunnerJobLight media type instance.
func (mt *OctorunnerJobLight) Validate() (err error) {

	if mt.CommitID == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "commitID"))
	}
	if mt.Job == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "job"))
	}

	if mt.Status == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "status"))
	}
	if mt.Extra == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "extra"))
	}
	if !(mt.Status == "running" || mt.Status == "done" || mt.Status == "error") {
		err = goa.MergeErrors(err, goa.InvalidEnumValueError(`response.status`, mt.Status, []interface{}{"running", "done", "error"}))
	}
	return
}

// OctorunnerJobCollection is the media type for an array of OctorunnerJob (default view)
//
// Identifier: application/vnd.octorunner.job+json; type=collection; view=default
type OctorunnerJobCollection []*OctorunnerJob

// Validate validates the OctorunnerJobCollection media type instance.
func (mt OctorunnerJobCollection) Validate() (err error) {
	for _, e := range mt {
		if e != nil {
			if err2 := e.Validate(); err2 != nil {
				err = goa.MergeErrors(err, err2)
			}
		}
	}
	return
}

// OctorunnerJobCollection is the media type for an array of OctorunnerJob (light view)
//
// Identifier: application/vnd.octorunner.job+json; type=collection; view=light
type OctorunnerJobLightCollection []*OctorunnerJobLight

// Validate validates the OctorunnerJobLightCollection media type instance.
func (mt OctorunnerJobLightCollection) Validate() (err error) {
	for _, e := range mt {
		if e != nil {
			if err2 := e.Validate(); err2 != nil {
				err = goa.MergeErrors(err, err2)
			}
		}
	}
	return
}

// Output contains a single line of output of a job (default view)
//
// Identifier: application/vnd.octorunner.output+json; view=default
type OctorunnerOutput struct {
	// The data, which is a single line of stdout or stderr
	Data *string `form:"data,omitempty" json:"data,omitempty" xml:"data,omitempty"`
	// Unique output ID
	ID *int `form:"id,omitempty" json:"id,omitempty" xml:"id,omitempty"`
	// The git commit ID specific to this job
	Timestamp *time.Time `form:"timestamp,omitempty" json:"timestamp,omitempty" xml:"timestamp,omitempty"`
}

// A (github) project that Octorunner ran jobs for (default view)
//
// Identifier: application/vnd.octorunner.project+json; view=default
type OctorunnerProject struct {
	// Unique project ID
	ID int `form:"id" json:"id" xml:"id"`
	// The project name
	Name string `form:"name" json:"name" xml:"name"`
	// The project's owner
	Owner string `form:"owner" json:"owner" xml:"owner"`
}

// Validate validates the OctorunnerProject media type instance.
func (mt *OctorunnerProject) Validate() (err error) {

	if mt.Name == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "name"))
	}
	if mt.Owner == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "owner"))
	}
	return
}

// OctorunnerProjectCollection is the media type for an array of OctorunnerProject (default view)
//
// Identifier: application/vnd.octorunner.project+json; type=collection; view=default
type OctorunnerProjectCollection []*OctorunnerProject

// Validate validates the OctorunnerProjectCollection media type instance.
func (mt OctorunnerProjectCollection) Validate() (err error) {
	for _, e := range mt {
		if e != nil {
			if err2 := e.Validate(); err2 != nil {
				err = goa.MergeErrors(err, err2)
			}
		}
	}
	return
}
