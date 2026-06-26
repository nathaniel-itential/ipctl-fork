// Copyright 2024 Itential Inc. All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential

package services

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/itential/ipctl/internal/logging"
	"github.com/itential/ipctl/pkg/client"
)

const (
	defaultWorkflowType          = "automation"
	defaultWorkflowCanvasVersion = 3
	defaultWorkflowFontSize      = 12
)

type getWorkflowsResponse struct {
	Items    []Workflow `json:"items"`
	Count    int        `json:"count"`
	End      int        `json:"end"`
	Limit    int        `json:"limit"`
	Next     string     `json:"next"`
	Previous string     `json:"previous"`
	Skip     int        `json:"skip"`
	Total    int        `json:"total"`
}

type Workflow struct {
	Id                 string                 `json:"_id,omitempty"`
	Description        string                 `json:"description,omitempty"`
	CanvasVersion      float64                `json:"canvasVersion"`
	Created            string                 `json:"created,omitempty"`
	CreatedVersion     string                 `json:"createdVersion,omitempty"`
	CreatedBy          interface{}            `json:"created_by,omitempty"`
	EncodingVersion    int                    `json:"encodingVersion,omitempty"`
	ErrorHandler       string                 `json:"errorHandler,omitempty"`
	FontSize           int                    `json:"font_size"`
	Groups             []interface{}          `json:"groups"`
	InputSchema        map[string]interface{} `json:"inputSchema"`
	LastUpdatedVersion string                 `json:"lastUpdatedVersion,omitempty"`
	LastUpdated        string                 `json:"last_updated,omitempty"`
	LastUpdatedBy      interface{}            `json:"last_updated_by,omitempty"`
	Name               string                 `json:"name"`
	OutputSchema       map[string]interface{} `json:"outputSchema"`
	PreAutomationTime  int                    `json:"preAutomationTime"`
	Sla                int                    `json:"sla"`
	Tags               []Tag                  `json:"tags"`
	Tasks              map[string]interface{} `json:"tasks"`
	Transitions        map[string]interface{} `json:"transitions"`
	Type               string                 `json:"type"`
}

// WorkflowService provides methods for managing Itential Platform workflows.
// It handles CRUD operations, import/export, and bulk operations on workflow assets.
type WorkflowService struct {
	BaseService
}

// NewWorkflowService creates a new instance of WorkflowService with the provided client.
func NewWorkflowService(c client.Client) *WorkflowService {
	return &WorkflowService{BaseService: NewBaseService(c)}
}

// NewWorkflow returns a minimum viable workflow struct.
func NewWorkflow(name string) Workflow {

	wfstart := map[string]interface{}{
		"groups": []any{},
		"name":   "workflow_start",
		"nodelocation": map[string]int{
			"x": 0,
			"y": -500,
		},
		"x": 0,
		"y": 0.5,
	}

	wfend := map[string]interface{}{
		"groups": []any{},
		"name":   "workflow_end",
		"nodelocation": map[string]int{
			"x": 0,
			"y": 500,
		},
		"x": 1,
		"y": 0.5,
	}

	ioschema := map[string]interface{}{
		"properties": map[string]interface{}{
			"_id": map[string]interface{}{
				"pattern": "^[0-9a-f]{24}$",
				"type":    "string",
			},
			"initiator": map[string]interface{}{
				"type": "string",
			},
		},
		"required": []any{},
		"type":     "object",
	}

	return Workflow{
		Name:          name,
		Type:          defaultWorkflowType,
		InputSchema:   ioschema,
		OutputSchema:  ioschema,
		CanvasVersion: defaultWorkflowCanvasVersion,
		FontSize:      defaultWorkflowFontSize,
		Groups:        []any{},
		Tags:          []Tag{},
		Transitions: map[string]interface{}{
			"workflow_start": map[string]interface{}{},
			"workflow_end":   map[string]interface{}{},
		},
		Tasks: map[string]interface{}{
			"workflow_start": wfstart,
			"workflow_end":   wfend,
		},
	}
}

// GetAll will retrieve all of the currently configured workflows available on the
// server.  If there are no configured workflows, this function will return an
// empty array.
func (svc *WorkflowService) GetAll() ([]Workflow, error) {
	logging.Trace()

	var workflows []Workflow

	var limit = 100
	var skip = 0

	for {
		// Declare res inside the loop so each page decodes into a freshly
		// allocated struct. Reusing a single res across pages lets
		// encoding/json merge map fields and reuse slice backing arrays,
		// bleeding fields from one page's elements into the next.
		var res getWorkflowsResponse
		if err := svc.GetRequest(&Request{
			uri:    "/automation-studio/workflows",
			params: &QueryParams{Limit: limit, Skip: skip},
		}, &res); err != nil {
			return nil, err
		}

		workflows = append(workflows, res.Items...)

		if len(workflows) == res.Total {
			break
		}

		skip += limit
	}

	logging.Info("Found %v workflow(s)", len(workflows))

	return workflows, nil

}

// Get retrieves the workflow as specified by the name argument.  If the
// specified workflow does not exist, this function will return an error with
// the message 'workflow not found'.   This function uses a query string per
// the API to find the worklow by name.  In some cases, more than one workflow
// could be returned from the server.  In this case, this function will return
// an error with message 'unable to find workflow'
func (svc *WorkflowService) Get(name string) (*Workflow, error) {
	logging.Trace()

	var res getWorkflowsResponse

	if err := svc.GetRequest(&Request{
		uri:   "/automation-studio/workflows",
		query: map[string]string{"equals[name]": name},
	}, &res); err != nil {
		return nil, err
	}

	if res.Total == 0 {
		return nil, errors.New("workflow not found")
	}

	if res.Total > 1 {
		logging.Debug("Get() workflows returned more than one workflow.  This is due to more than one workflow with the same name")
		return nil, errors.New("unable to find workflow")
	}

	return &res.Items[0], nil
}

// Create will add a new workflow asset to the Itential Platform server.  The
// workflow will always be created even if another workflow by the same name
// already exists.
func (svc *WorkflowService) Create(in Workflow) (*Workflow, error) {
	logging.Trace()

	type Response struct {
		Created *Workflow `json:"created"`
		Edit    string    `json:"edit"`
	}

	var res Response

	if err := svc.PostRequest(&Request{
		uri:                "/automation-studio/automations",
		body:               map[string]interface{}{"automation": in},
		expectedStatusCode: http.StatusOK,
	}, &res); err != nil {
		return nil, err
	}

	return res.Created, nil
}

// Delete removes an existing workflow from the set of configured workflows on
// the server.  If the workflow does not exist, this function will return an
// error.
func (svc *WorkflowService) Delete(name string) error {
	logging.Trace()
	return svc.BaseService.Delete(
		fmt.Sprintf("/workflow_builder/workflows/delete/%s", name),
	)
}

// Import will import a workflow into the current server.  If a workflow with
// the same name already exists on the server, this function will return an
// error.
func (svc *WorkflowService) Import(in Workflow) (*Workflow, error) {
	logging.Trace()

	body := map[string]interface{}{
		"automations": []interface{}{in},
	}

	var res *Workflow

	if err := svc.PostRequest(&Request{
		uri:                "/automation-studio/automations/import",
		body:               &body,
		expectedStatusCode: http.StatusOK,
	}, &res); err != nil {
		return nil, err
	}

	return res, nil
}

// Export returns an exported workflow from the server.  An exported workflow
// differents from a Get in that the CreateBy and UpdatedBy fields are
// expanded.
func (svc *WorkflowService) Export(name string) (*Workflow, error) {
	logging.Trace()

	body := map[string]interface{}{
		"options": map[string]interface{}{
			"name": name,
		},
	}

	var res Workflow

	if err := svc.PostRequest(&Request{
		uri:                "/workflow_builder/export",
		body:               &body,
		expectedStatusCode: http.StatusOK,
	}, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// ExportById returns an exported workflow from the server by ID.
// An exported workflow differs from Get in that the CreatedBy and UpdatedBy
// fields are expanded.
func (svc *WorkflowService) ExportById(id string) (*Workflow, error) {
	logging.Trace()

	body := map[string]interface{}{
		"options": map[string]interface{}{
			"_id": id,
		},
	}

	var res Workflow

	if err := svc.PostRequest(&Request{
		uri:                "/workflow_builder/export",
		body:               &body,
		expectedStatusCode: http.StatusOK,
	}, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// GetById retrieves a specific workflow by its unique ID.
// This method fetches all workflows and filters by ID client-side.
// DEPRECATED: Business logic method - prefer using resources.WorkflowResource.GetById
func (svc *WorkflowService) GetById(id string) (*Workflow, error) {
	logging.Trace()

	workflows, err := svc.GetAll()
	if err != nil {
		return nil, err
	}

	for _, wf := range workflows {
		if wf.Id == id {
			return &wf, nil
		}
	}

	return nil, errors.New("workflow not found")
}

// Clear removes all workflows from the server by deleting each workflow individually.
// DEPRECATED: Business logic method - prefer using resources.WorkflowResource.Clear
func (svc *WorkflowService) Clear() error {
	logging.Trace()

	workflows, err := svc.GetAll()
	if err != nil {
		return err
	}

	for _, wf := range workflows {
		if err := svc.Delete(wf.Name); err != nil {
			return err
		}
	}

	return nil
}

// Update modifies an existing workflow on the server.
// The workflow must have a valid ID field. Returns the updated workflow
// or an error if the update fails.
func (svc *WorkflowService) Update(in Workflow) (*Workflow, error) {
	logging.Trace()

	var res *Workflow

	if err := svc.PutRequest(&Request{
		uri:                fmt.Sprintf("/automation-studio/automations/%s", in.Id),
		body:               map[string]interface{}{"update": in},
		expectedStatusCode: http.StatusOK,
	}, &res); err != nil {
		return nil, err
	}

	return res, nil
}
