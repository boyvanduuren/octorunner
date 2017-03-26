package design

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
)

var _ = Resource("project", func() {
	BasePath("/projects")
	DefaultMedia(Project)

	Action("list", func() {
		Description("Get all projects")
		Routing(GET(""))
		Response(OK, CollectionOf(Project))
	})

	Action("show", func() {
		Description("Get a project by id")
		Routing(GET("/:projectID"))
		Params(func() {
			Param("projectID", Integer, "Project ID")
		})
		Response(OK)
		Response(NotFound)
	})

	Action("jobs", func() {
		Description("Get all jobs belonging to a project, but without their data")
		Routing(GET("/:projectID/jobs"))
		Params(func() {
			Param("projectID", Integer, "Project ID")
		})
		// We don't want to eagerly fetch all data of every job, so we return
		// a collection of the light Job view.
		Response(OK, func() {
			Media(CollectionOf(Job), "light")
		})
		Response(NotFound)
	})

})

var _ = Resource("job", func() {
	BasePath("/jobs")
	DefaultMedia(Job)

	Action("show", func() {
		Description("Get a job by its ID")
		Routing(GET("/:jobID"))
		Params(func() {
			Param("jobID", Integer, "Job ID")
		})
		Response(OK)
		Response(NotFound)
	})

	Action("showLatest", func() {
		Description("Show the latest job")
		Routing(GET("/latest"))
		Response(OK)
		Response(NotFound)
	})
})
