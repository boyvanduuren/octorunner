package controllers

import (
	"github.com/boyvanduuren/octorunner/lib/webapi/app"
	"github.com/goadesign/goa"
	"github.com/boyvanduuren/octorunner/lib/persist"
	"github.com/goadesign/goa/logging/logrus"
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
	res, err := persist.DBConn.FindJobWithData(int64(ctx.JobID))
	if err != nil {
		goalogrus.Entry(ctx).Errorf("While querying job %q: %q", int(ctx.JobID), err)
		return ctx.NotFound()
	}


	dataCollection := make([]*app.OctorunnerOutput, len(res.Data))
	for i, output := range res.Data {
		outputID := new(int)
		*outputID = int(output.ID)
		dataCollection[i] = &app.OctorunnerOutput{
			ID: outputID,
			Data: &output.Data,
			Timestamp: &output.Timestamp,
		}
	}

	foundProject := &app.OctorunnerJob{
		ID: int(res.ID),
		CommitID: res.CommitID,
		Project: int(res.Project),
		Job: res.Job,
		Data: dataCollection,

	}

	return ctx.OK(foundProject)
	// JobController_Show: end_implement
}

// ShowLatest runs the showLatest action.
func (c *JobController) ShowLatest(ctx *app.ShowLatestJobContext) error {
	// JobController_ShowLatest: start_implement

	// Put your logic here

	// JobController_ShowLatest: end_implement
	res := &app.OctorunnerJob{}
	return ctx.OK(res)
}
