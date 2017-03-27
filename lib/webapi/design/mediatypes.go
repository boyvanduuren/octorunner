package design

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
)

// Projects are the base data entity and have 1..n jobs.
var Project = MediaType("application/vnd.octorunner.project+json", func() {
	Description("A (github) project that Octorunner ran jobs for")
	Attributes(func() {
		Attribute("id", Integer, "Unique project ID", func() {
			Example(1)
		})
		Attribute("name", String, "The project name", func() {
			Example("octorunner")
		})
		Attribute("owner", String, "The project's owner", func() {
			Example("boyvanduuren")
		})
		Required("id", "name", "owner")
	})
	View("default", func() {
		Attribute("id")
		Attribute("name")
		Attribute("owner")
	})
})

// Jobs are units of work that ran after a push to a Github repository. These units
// of work produce output on STDOUT and STDERR. Job entities are used to group the output
// that belongs to a single job.
var Job = MediaType("application/vnd.octorunner.job+json", func() {
	Description("A job that was ran after a commit on a project")
	Attributes(func() {
		Attribute("id", Integer, "Unique job ID", func() {
			Example(1)
		})
		Attribute("iteration", Integer, "The iteration ID of this job. A job might be ran multiple times.", func() {
			Example(5)
		})
		Attribute("project", Integer, "The project this job belongs to", func() {
			Example(1)
		})
		Attribute("commitID", String, "The git commit ID specific to this job", func() {
			Example("093a16cb43d696d32ae73a529c6165b80c1ce844")
		})
		Attribute("job", String, "The name of the job", func() {
			Example("default")
		})
		Attribute("status", String, "The status of the job", func() {
			Example("running")
			Enum("running", "done", "error")
		})
		Attribute("extra", String, "Extra information, this might contain error information", func() {
			Example("Some error message")
		})
		Attribute("data", ArrayOf(Output))
		Required("id", "project", "commitID", "job", "iteration", "status", "extra")
	})
	View("default", func() {
		Attribute("id")
		Attribute("iteration")
		Attribute("project")
		Attribute("commitID")
		Attribute("job")
		Attribute("status")
		Attribute("extra")
		Attribute("data")
	})
	View("light", func() {
		Attribute("id")
		Attribute("iteration")
		Attribute("project")
		Attribute("commitID")
		Attribute("job")
		Attribute("status")
		Attribute("extra")
	})
})

// The output belonging to a job. Every line has its own output row.
var Output = MediaType("application/vnd.octorunner.output+json", func() {
	Description("Output contains a single line of output of a job")
	Attributes(func() {
		Attribute("id", Integer, "Unique output ID", func() {
			Example(1)
		})
		Attribute("data", String, "The data, which is a single line of stdout or stderr", func() {
			Example("some stdout line")
		})
		Attribute("timestamp", DateTime, "The git commit ID specific to this job")
	})
	View("default", func() {
		Attribute("id")
		Attribute("data")
		Attribute("timestamp")
	})
})
