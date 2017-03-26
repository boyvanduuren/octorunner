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

func ToProjectMedia(job *persist.Job) (*app.OctorunnerJob) {
	dataCollection := make([]*app.OctorunnerOutput, len(job.Data))
	for i, output := range job.Data {
		outputID := new(int)
		*outputID = int(output.ID)
		dataCollection[i] = &app.OctorunnerOutput{
			ID: outputID,
			Data: &output.Data,
			Timestamp: &output.Timestamp,
		}
	}

	return &app.OctorunnerJob{
		ID: int(job.ID),
		CommitID: job.CommitID,
		Project: int(job.Project),
		Job: job.Job,
		Data: dataCollection,

	}
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

	return ctx.OK(ToProjectMedia(res))
	// JobController_Show: end_implement
}

// ShowLatest runs the showLatest action.
func (c *JobController) ShowLatest(ctx *app.ShowLatestJobContext) error {
	// JobController_ShowLatest: start_implement

	// Put your logic here
	latestJobID := persist.DBConn.GetLatestJobID()
	if latestJobID < 1 {
		return ctx.NotFound()
	}

	res, err := persist.DBConn.FindJobWithData(latestJobID)
	if err != nil {
		goalogrus.Entry(ctx).Errorf("While querying latest job %q: %q", int(latestJobID), err)
		return ctx.NotFound()
	}

	return ctx.OK(ToProjectMedia(res))
	// JobController_ShowLatest: end_implement
}
