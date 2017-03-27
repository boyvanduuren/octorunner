package controllers

import (
	"github.com/boyvanduuren/octorunner/lib/webapi/app"
	"github.com/goadesign/goa"
	"github.com/boyvanduuren/octorunner/lib/persist"
	"github.com/goadesign/goa/logging/logrus"
)

// ProjectController implements the project resource.
type ProjectController struct {
	*goa.Controller
}

// NewProjectController creates a project controller.
func NewProjectController(service *goa.Service) *ProjectController {
	return &ProjectController{Controller: service.NewController("ProjectController")}
}

// Jobs runs the jobs action.
func (c *ProjectController) Jobs(ctx *app.JobsProjectContext) error {
	// ProjectController_Jobs: start_implement

	// Put your logic here
	jobs, err := persist.DBConn.FindJobsForProject(int64(ctx.ProjectID))
	if err != nil {
		goalogrus.Entry(ctx).Errorf("While finding jobs for project %q: %q", ctx.ProjectID, err)
		return ctx.NotFound()
	}

	jobCollection := make(app.OctorunnerJobLightCollection, len(jobs))
	for i, job := range jobs {
		jobCollection[i] = &app.OctorunnerJobLight{
			ID: int(job.ID),
			Iteration: int(job.Iteration),
			CommitID: job.CommitID,
			Project: int(job.Project),
			Job: job.Job,
			Status: job.Status,
			Extra: job.Extra,
		}
	}

	return ctx.OKLight(jobCollection)
	// ProjectController_Jobs: end_implement
}

// List runs the list action.
func (c *ProjectController) List(ctx *app.ListProjectContext) error {
	// ProjectController_List: start_implement

	// Put your logic here
	projects, err := persist.DBConn.FindAllProjects()
	if err != nil {
		goalogrus.Entry(ctx).Errorf("While querying all projects: %q", err)
		res := app.OctorunnerProjectCollection{}
		return ctx.OK(res)
	}

	projectCollection := make(app.OctorunnerProjectCollection, len(*projects))
	for i, project := range *projects {
		projectCollection[i] = &app.OctorunnerProject{
			ID: int(project.ID),
			Name: project.Name,
			Owner: project.Owner,
		}
	}

	return ctx.OK(projectCollection)
	// ProjectController_List: end_implement
}

// Show runs the show action.
func (c *ProjectController) Show(ctx *app.ShowProjectContext) error {
	// ProjectController_Show: start_implement

	// Put your logic here
	res, err := persist.DBConn.FindProjectByID(int64(ctx.ProjectID))
	if err != nil {
		goalogrus.Entry(ctx).Errorf("While querying project %q: %q", ctx.ProjectID, err)
		return ctx.NotFound()
	}
	foundProject := &app.OctorunnerProject{
		ID: int(res.ID),
		Name: res.Name,
		Owner: res.Owner,
	}

	return ctx.OK(foundProject)
	// ProjectController_Show: end_implement
}
