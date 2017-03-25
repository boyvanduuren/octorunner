package controllers

import (
	"github.com/boyvanduuren/octorunner/lib/webapi/app"
	"github.com/goadesign/goa"
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

	// ProjectController_Jobs: end_implement
	res := app.OctorunnerJobCollection{}
	return ctx.OK(res)
}

// List runs the list action.
func (c *ProjectController) List(ctx *app.ListProjectContext) error {
	// ProjectController_List: start_implement

	// Put your logic here

	// ProjectController_List: end_implement
	res := app.OctorunnerProjectCollection{}
	return ctx.OK(res)
}

// Show runs the show action.
func (c *ProjectController) Show(ctx *app.ShowProjectContext) error {
	// ProjectController_Show: start_implement

	// Put your logic here

	// ProjectController_Show: end_implement
	res := &app.OctorunnerProject{}
	return ctx.OK(res)
}
