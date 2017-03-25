package controllers

import (
	"github.com/boyvanduuren/octorunner/lib/webapi/app"
	"github.com/goadesign/goa"
)

// JobController implements the job resource.
type JobController struct {
	*goa.Controller
}

// NewJobController creates a job controller.
func NewJobController(service *goa.Service) *JobController {
	return &JobController{Controller: service.NewController("JobController")}
}

// Show runs the show action.
func (c *JobController) Show(ctx *app.ShowJobContext) error {
	// JobController_Show: start_implement

	// Put your logic here

	// JobController_Show: end_implement
	res := &app.OctorunnerJob{}
	return ctx.OK(res)
}

// ShowLatest runs the showLatest action.
func (c *JobController) ShowLatest(ctx *app.ShowLatestJobContext) error {
	// JobController_ShowLatest: start_implement

	// Put your logic here

	// JobController_ShowLatest: end_implement
	res := &app.OctorunnerJob{}
	return ctx.OK(res)
}
