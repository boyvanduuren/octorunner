// Code generated by goagen v1.1.0-dirty, command line:
// $ goagen
// --design=github.com/boyvanduuren/octorunner/lib/webapi/design
// --out=$(GOPATH)\src\github.com\boyvanduuren\octorunner\lib\webapi
// --version=v1.1.0-dirty
//
// API "octorunner": Application Contexts
//
// The content of this file is auto-generated, DO NOT MODIFY

package app

import (
	"github.com/goadesign/goa"
	"golang.org/x/net/context"
	"net/http"
	"strconv"
)

// ShowJobContext provides the job show action context.
type ShowJobContext struct {
	context.Context
	*goa.ResponseData
	*goa.RequestData
	JobID int
}

// NewShowJobContext parses the incoming request URL and body, performs validations and creates the
// context used by the job controller show action.
func NewShowJobContext(ctx context.Context, r *http.Request, service *goa.Service) (*ShowJobContext, error) {
	var err error
	resp := goa.ContextResponse(ctx)
	resp.Service = service
	req := goa.ContextRequest(ctx)
	req.Request = r
	rctx := ShowJobContext{Context: ctx, ResponseData: resp, RequestData: req}
	paramJobID := req.Params["jobID"]
	if len(paramJobID) > 0 {
		rawJobID := paramJobID[0]
		if jobID, err2 := strconv.Atoi(rawJobID); err2 == nil {
			rctx.JobID = jobID
		} else {
			err = goa.MergeErrors(err, goa.InvalidParamTypeError("jobID", rawJobID, "integer"))
		}
	}
	return &rctx, err
}

// OK sends a HTTP response with status code 200.
func (ctx *ShowJobContext) OK(r *OctorunnerJob) error {
	ctx.ResponseData.Header().Set("Content-Type", "application/vnd.octorunner.job+json")
	return ctx.ResponseData.Service.Send(ctx.Context, 200, r)
}

// OKLight sends a HTTP response with status code 200.
func (ctx *ShowJobContext) OKLight(r *OctorunnerJobLight) error {
	ctx.ResponseData.Header().Set("Content-Type", "application/vnd.octorunner.job+json")
	return ctx.ResponseData.Service.Send(ctx.Context, 200, r)
}

// NotFound sends a HTTP response with status code 404.
func (ctx *ShowJobContext) NotFound() error {
	ctx.ResponseData.WriteHeader(404)
	return nil
}

// ShowLatestJobContext provides the job showLatest action context.
type ShowLatestJobContext struct {
	context.Context
	*goa.ResponseData
	*goa.RequestData
}

// NewShowLatestJobContext parses the incoming request URL and body, performs validations and creates the
// context used by the job controller showLatest action.
func NewShowLatestJobContext(ctx context.Context, r *http.Request, service *goa.Service) (*ShowLatestJobContext, error) {
	var err error
	resp := goa.ContextResponse(ctx)
	resp.Service = service
	req := goa.ContextRequest(ctx)
	req.Request = r
	rctx := ShowLatestJobContext{Context: ctx, ResponseData: resp, RequestData: req}
	return &rctx, err
}

// OK sends a HTTP response with status code 200.
func (ctx *ShowLatestJobContext) OK(r *OctorunnerJob) error {
	ctx.ResponseData.Header().Set("Content-Type", "application/vnd.octorunner.job+json")
	return ctx.ResponseData.Service.Send(ctx.Context, 200, r)
}

// OKLight sends a HTTP response with status code 200.
func (ctx *ShowLatestJobContext) OKLight(r *OctorunnerJobLight) error {
	ctx.ResponseData.Header().Set("Content-Type", "application/vnd.octorunner.job+json")
	return ctx.ResponseData.Service.Send(ctx.Context, 200, r)
}

// NotFound sends a HTTP response with status code 404.
func (ctx *ShowLatestJobContext) NotFound() error {
	ctx.ResponseData.WriteHeader(404)
	return nil
}

// JobsProjectContext provides the project jobs action context.
type JobsProjectContext struct {
	context.Context
	*goa.ResponseData
	*goa.RequestData
	ProjectID int
}

// NewJobsProjectContext parses the incoming request URL and body, performs validations and creates the
// context used by the project controller jobs action.
func NewJobsProjectContext(ctx context.Context, r *http.Request, service *goa.Service) (*JobsProjectContext, error) {
	var err error
	resp := goa.ContextResponse(ctx)
	resp.Service = service
	req := goa.ContextRequest(ctx)
	req.Request = r
	rctx := JobsProjectContext{Context: ctx, ResponseData: resp, RequestData: req}
	paramProjectID := req.Params["projectID"]
	if len(paramProjectID) > 0 {
		rawProjectID := paramProjectID[0]
		if projectID, err2 := strconv.Atoi(rawProjectID); err2 == nil {
			rctx.ProjectID = projectID
		} else {
			err = goa.MergeErrors(err, goa.InvalidParamTypeError("projectID", rawProjectID, "integer"))
		}
	}
	return &rctx, err
}

// OKLight sends a HTTP response with status code 200.
func (ctx *JobsProjectContext) OKLight(r OctorunnerJobLightCollection) error {
	ctx.ResponseData.Header().Set("Content-Type", "application/vnd.octorunner.job+json; type=collection")
	if r == nil {
		r = OctorunnerJobLightCollection{}
	}
	return ctx.ResponseData.Service.Send(ctx.Context, 200, r)
}

// NotFound sends a HTTP response with status code 404.
func (ctx *JobsProjectContext) NotFound() error {
	ctx.ResponseData.WriteHeader(404)
	return nil
}

// ListProjectContext provides the project list action context.
type ListProjectContext struct {
	context.Context
	*goa.ResponseData
	*goa.RequestData
}

// NewListProjectContext parses the incoming request URL and body, performs validations and creates the
// context used by the project controller list action.
func NewListProjectContext(ctx context.Context, r *http.Request, service *goa.Service) (*ListProjectContext, error) {
	var err error
	resp := goa.ContextResponse(ctx)
	resp.Service = service
	req := goa.ContextRequest(ctx)
	req.Request = r
	rctx := ListProjectContext{Context: ctx, ResponseData: resp, RequestData: req}
	return &rctx, err
}

// OK sends a HTTP response with status code 200.
func (ctx *ListProjectContext) OK(r OctorunnerProjectCollection) error {
	ctx.ResponseData.Header().Set("Content-Type", "application/vnd.octorunner.project+json; type=collection")
	if r == nil {
		r = OctorunnerProjectCollection{}
	}
	return ctx.ResponseData.Service.Send(ctx.Context, 200, r)
}

// ShowProjectContext provides the project show action context.
type ShowProjectContext struct {
	context.Context
	*goa.ResponseData
	*goa.RequestData
	ProjectID int
}

// NewShowProjectContext parses the incoming request URL and body, performs validations and creates the
// context used by the project controller show action.
func NewShowProjectContext(ctx context.Context, r *http.Request, service *goa.Service) (*ShowProjectContext, error) {
	var err error
	resp := goa.ContextResponse(ctx)
	resp.Service = service
	req := goa.ContextRequest(ctx)
	req.Request = r
	rctx := ShowProjectContext{Context: ctx, ResponseData: resp, RequestData: req}
	paramProjectID := req.Params["projectID"]
	if len(paramProjectID) > 0 {
		rawProjectID := paramProjectID[0]
		if projectID, err2 := strconv.Atoi(rawProjectID); err2 == nil {
			rctx.ProjectID = projectID
		} else {
			err = goa.MergeErrors(err, goa.InvalidParamTypeError("projectID", rawProjectID, "integer"))
		}
	}
	return &rctx, err
}

// OK sends a HTTP response with status code 200.
func (ctx *ShowProjectContext) OK(r *OctorunnerProject) error {
	ctx.ResponseData.Header().Set("Content-Type", "application/vnd.octorunner.project+json")
	return ctx.ResponseData.Service.Send(ctx.Context, 200, r)
}

// NotFound sends a HTTP response with status code 404.
func (ctx *ShowProjectContext) NotFound() error {
	ctx.ResponseData.WriteHeader(404)
	return nil
}
