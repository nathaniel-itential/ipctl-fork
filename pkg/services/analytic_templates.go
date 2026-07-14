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

type AnalyticTemplateCommandRule struct {
	Type      string `json:"type"`
	PreRegex  string `json:"preRegex"`
	PostRegex string `json:"postRegex"`
	Evaluator string `json:"evaulator"`
	Severity  string `json:"severity"`
}

type AnalyticTemplateCommand struct {
	PreRawCommand  string                        `json:"preRawCommand"`
	PostRawCommand string                        `json:"postRawCommand"`
	PassRule       bool                          `json:"passRule"`
	Rules          []AnalyticTemplateCommandRule `json:"rules"`
}

type AnalyticTemplate struct {
	Id              string                    `json:"_id"`
	Name            string                    `json:"name"`
	Tags            []Tag                     `json:"tags"`
	PassRule        bool                      `json:"passRule"`
	PrePostCommands []AnalyticTemplateCommand `json:"prepostCommands"`
	Created         int                       `json:"created"`
	CreatedBy       string                    `json:"createdBy"`
	LastUpdated     int                       `json:"lastUpdated"`
	LastUpdatedBy   string                    `json:"lastUpdatedBy"`
}

type AnalyticTemplateService struct {
	BaseService
}

func NewAnalyticTemplateService(c client.Client) *AnalyticTemplateService {
	return &AnalyticTemplateService{BaseService: NewBaseService(c)}
}

func NewAnalyticTemplate(name string) AnalyticTemplate {
	logging.Trace()
	return AnalyticTemplate{
		Id:   name,
		Name: name,
		Tags: []Tag{},
		PrePostCommands: []AnalyticTemplateCommand{
			AnalyticTemplateCommand{
				PassRule: true,
				Rules: []AnalyticTemplateCommandRule{
					AnalyticTemplateCommandRule{
						Type:      "regex",
						Evaluator: "=",
						Severity:  "error",
					},
				},
			},
		},
	}
}

// GetAll returns all configured analytic-templates found on the server. If
// there are no analytic templates configured, this function will return an
// empty array
func (svc *AnalyticTemplateService) GetAll() ([]AnalyticTemplate, error) {
	logging.Trace()

	var templates []AnalyticTemplate
	var uri = "/mop/listAnalyticTemplates"

	if err := svc.BaseService.Get(uri, &templates); err != nil {
		return nil, err
	}

	return templates, nil
}

// Get returns the specified analytic template.  If the template specified by
// name does not exist, this function will return an error
func (svc *AnalyticTemplateService) Get(name string) (*AnalyticTemplate, error) {
	logging.Trace()

	var template []AnalyticTemplate
	var uri = fmt.Sprintf("/mop/listAnAnalyticTemplate/%s", name)

	if err := svc.BaseService.Get(uri, &template); err != nil {
		return nil, err
	}

	if len(template) == 0 {
		return nil, errors.New("analytic template not found")
	} else if len(template) > 1 {
		return nil, errors.New("return more than 1 template")
	}

	return &template[0], nil
}

func (svc *AnalyticTemplateService) Create(in AnalyticTemplate) (*AnalyticTemplate, error) {
	logging.Trace()

	body := map[string]AnalyticTemplate{"template": in}

	type Response struct {
		Result        map[string]interface{} `json:"result"`
		Ops           []AnalyticTemplate     `json:"ops"`
		InsertedCount int                    `json:"insertedCount"`
		InsertedIds   map[string]interface{} `json:"insertedIds"`
	}

	var res Response

	if err := svc.PostRequest(&Request{
		uri:                "/mop/createAnalyticTemplate",
		body:               &body,
		expectedStatusCode: http.StatusOK,
	}, &res); err != nil {
		return nil, err
	}

	return &res.Ops[0], nil
}

// Delete will remove the specified analytic template from the server.
func (svc *AnalyticTemplateService) Delete(id string) error {
	logging.Trace()
	return svc.PostRequest(&Request{
		uri:                fmt.Sprintf("/mop/deleteAnalyticTemplate/%s", id),
		expectedStatusCode: http.StatusOK,
	}, nil)
}

// Import will import an analytic template
func (svc *AnalyticTemplateService) Import(in AnalyticTemplate) error {
	logging.Trace()

	body := map[string]interface{}{
		"type":     "analytic",
		"template": in,
	}

	return svc.Post("/mop/import", &body, nil)
}

func (svc *AnalyticTemplateService) Export(name string) (*AnalyticTemplate, error) {
	logging.Trace()

	body := map[string]interface{}{
		"options": map[string]interface{}{
			"name": name,
		},
		"type": "analytic",
	}

	var res *AnalyticTemplate

	if err := svc.PostRequest(&Request{
		uri:                "/mop/export",
		body:               &body,
		expectedStatusCode: http.StatusOK,
	}, &res); err != nil {
		return nil, err
	}
	return res, nil
}
