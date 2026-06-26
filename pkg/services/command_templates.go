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

type CommandTemplateRule struct {
	Rule     string `json:"rule"`
	Eval     string `json:"eval"`
	Severity string `json:"severity"`
}

type CommandTemplateCommand struct {
	Command  string                `json:"command"`
	PassRule bool                  `json:"passRule"`
	Rules    []CommandTemplateRule `json:"rules"`
}

type CommandTemplate struct {
	Id             string                   `json:"_id,omitempty"`
	Name           string                   `json:"name"`
	Os             string                   `json:"os"`
	PassRule       bool                     `json:"passRule"`
	IgnoreWarnings bool                     `json:"ignoreWarnings"`
	Commands       []CommandTemplateCommand `json:"commands"`
	Tags           []Tag                    `json:"tags"`
	Created        int                      `json:"created"`
	CreatedBy      string                   `json:"createdBy"`
	LastUpdated    int                      `json:"lastUpdated"`
	LastUpdatedBy  string                   `json:"lastUpdatedBy"`
	Namespace      any                      `json:"namespace"`
}

type CommandTemplateService struct {
	BaseService
}

func NewCommandTemplateService(c client.Client) *CommandTemplateService {
	return &CommandTemplateService{BaseService: NewBaseService(c)}
}

func NewCommandTemplate(name string) CommandTemplate {
	logging.Trace()

	cmd := CommandTemplateCommand{
		PassRule: true,
		Rules: []CommandTemplateRule{
			CommandTemplateRule{
				Eval:     "contains",
				Severity: "error",
			},
		},
	}

	return CommandTemplate{
		Id:       name,
		Name:     name,
		Commands: []CommandTemplateCommand{cmd},
		Tags:     []Tag{},
	}
}

// Get will return the specified command template from the server.  If the
// specified template does not exist, this function will return an error.
func (svc *CommandTemplateService) Get(name string) (*CommandTemplate, error) {
	logging.Trace()

	var res []*CommandTemplate
	var uri = fmt.Sprintf("/mop/listATemplate/%s", name)

	if err := svc.BaseService.Get(uri, &res); err != nil {
		return nil, err
	}

	if len(res) == 0 {
		return nil, errors.New("command template not found")
	} else if len(res) > 1 {
		return nil, errors.New("unable to find command template")
	}

	return res[0], nil
}

func (svc *CommandTemplateService) GetByName(name string) (*CommandTemplate, error) {
	logging.Trace()

	elements, err := svc.GetAll()
	if err != nil {
		return nil, err
	}

	var res *CommandTemplate

	for _, ele := range elements {
		if ele.Name == name {
			res = &ele
			break
		}
	}

	if res == nil {
		return nil, errors.New("command-template not found")
	}

	return res, nil
}

func (svc *CommandTemplateService) GetAll() ([]CommandTemplate, error) {
	logging.Trace()
	var res []CommandTemplate
	var uri = "/mop/listTemplates"
	if err := svc.BaseService.Get(uri, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func (svc *CommandTemplateService) Create(in CommandTemplate) (*CommandTemplate, error) {
	logging.Trace()

	// NOTE (hashdigest) The Id must be set to the same value as the name
	// otherwise the document cannot be found later
	in.Id = in.Name

	body := map[string]CommandTemplate{"mop": in}

	type Response struct {
		Result        map[string]interface{} `json:"result"`
		Ops           []CommandTemplate      `json:"ops"`
		InsertedCount int                    `json:"insertedCount"`
		InsertedIds   map[string]interface{} `json:"insertedIds"`
	}

	var res Response

	if err := svc.PostRequest(&Request{
		uri:                "/mop/createTemplate",
		body:               &body,
		expectedStatusCode: http.StatusOK,
	}, &res); err != nil {
		return nil, err
	}

	return &res.Ops[0], nil
}

func (svc *CommandTemplateService) Delete(name string) error {
	logging.Trace()

	type Response struct {
		Acknowledged bool `json:"acknowledged"`
		DeletedCount int  `json:"deletedCount"`
	}

	var res Response

	if err := svc.PostRequest(&Request{
		uri:                fmt.Sprintf("/mop/deleteTemplate/%s", name),
		expectedStatusCode: http.StatusOK,
	}, &res); err != nil {
		return err
	}

	return nil
}

func (svc *CommandTemplateService) Clear() error {
	logging.Trace()
	elements, err := svc.GetAll()
	if err != nil {
		return err
	}
	for _, ele := range elements {
		if err := svc.Delete(ele.Id); err != nil {
			return err
		}
	}
	return nil
}

func (svc *CommandTemplateService) Import(in CommandTemplate) error {
	logging.Trace()

	body := map[string]interface{}{
		"type":     "templates",
		"template": in,
	}

	return svc.PostRequest(&Request{
		uri:                "/mop/import",
		body:               &body,
		expectedStatusCode: http.StatusOK,
	}, nil)
}

func (svc *CommandTemplateService) Export(name string) (*CommandTemplate, error) {
	logging.Trace()

	body := map[string]interface{}{
		"options": map[string]interface{}{
			"name": name,
		},
		"type": "templates",
	}

	var res *CommandTemplate

	if err := svc.PostRequest(&Request{
		uri:                "/mop/export",
		body:               &body,
		expectedStatusCode: http.StatusOK,
	}, &res); err != nil {
		return nil, err
	}

	return res, nil
}
