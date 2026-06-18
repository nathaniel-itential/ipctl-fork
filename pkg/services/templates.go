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

// Template represents a template in the Automation Studio
type Template struct {
	Id            string                 `json:"_id,omitempty"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Type          string                 `json:"type"`
	Command       string                 `json:"command"`
	Group         string                 `json:"group"`
	Template      string                 `json:"template"`
	Data          string                 `json:"data"`
	Created       string                 `json:"created"`
	CreatedBy     any                    `json:"createdBy"`
	LastUpdated   string                 `json:"lastUpdated"`
	LastUpdatedBy any                    `json:"lastUpdateBy,omitempty"`
	Namespace     map[string]interface{} `json:"namespace,omitempty"`
	Tags          []interface{}          `json:"tags"`
}

// TemplateService provides methods for managing templates
type TemplateService struct {
	BaseService
}

// NewTemplateService creates a new TemplateService with the given client
func NewTemplateService(c client.Client) *TemplateService {
	return &TemplateService{BaseService: NewBaseService(c)}
}

// NewTemplate creates a new Template instance with the given parameters
// If type t is empty, defaults to "textfsm"
func NewTemplate(name, group, description, t string) Template {
	logging.Trace()

	if t == "" {
		t = "textfsm"
	}

	return Template{
		Name:        name,
		Group:       group,
		Description: description,
		Type:        t,
	}
}

// GetAll retrieves all templates from the server
func (svc *TemplateService) GetAll() ([]Template, error) {
	logging.Trace()

	var templates []Template

	var limit = 100
	var skip = 0

	// NOTE (privateip) I believe that if the query params are not specified
	// this API will simply return all items which is contrary to the API
	// documentation.  Need to test
	for {
		// Declare res inside the loop so each page decodes into a freshly
		// allocated struct. Reusing a single res across pages lets
		// encoding/json merge map fields and reuse slice backing arrays,
		// bleeding fields from one page's elements into the next.
		var res PaginatedResponse
		if err := svc.GetRequest(&Request{
			uri:    "/automation-studio/templates",
			params: &QueryParams{Limit: limit, Skip: skip},
		}, &res); err != nil {
			return nil, err
		}

		for _, ele := range res.Items {
			var t Template
			if err := Unmarshal(ele, &t); err != nil {
				return nil, err
			}
			templates = append(templates, t)
		}

		if len(templates) == res.Total {
			break
		}

		skip += limit
	}

	logging.Info("GetAll found %v template(s)", len(templates))

	return templates, nil
}

// Get retrieves a template by its ID
func (svc *TemplateService) Get(id string) (*Template, error) {
	logging.Trace()

	var res *Template
	var uri = fmt.Sprintf("/automation-studio/templates/%s", id)

	// FIXME (privateip) This can be optimzied by using query params instead of
	// iterating over all configured templates
	if err := svc.BaseService.Get(uri, &res); err != nil {
		return nil, err
	}

	return res, nil
}

// GetByName retrieves a template by name using client-side filtering.
// DEPRECATED: Business logic method - prefer using resources.TemplateResource.GetByName
func (svc *TemplateService) GetByName(name string) (*Template, error) {
	logging.Trace()

	templates, err := svc.GetAll()
	if err != nil {
		return nil, err
	}

	for i := range templates {
		if templates[i].Name == name {
			return &templates[i], nil
		}
	}

	return nil, errors.New("template not found")
}

// Create creates a new template
func (svc *TemplateService) Create(in Template) (*Template, error) {
	logging.Trace()

	body := map[string]map[string]interface{}{
		"template": map[string]interface{}{
			"name":        in.Name,
			"group":       in.Group,
			"type":        in.Type,
			"description": in.Description,
		},
	}

	type Response struct {
		Template *Template `json:"created"`
		Edit     string    `json:"edit"`
	}

	var res Response

	if err := svc.PostRequest(&Request{
		uri:                "/automation-studio/templates",
		body:               &body,
		expectedStatusCode: http.StatusOK,
	}, &res); err != nil {
		return nil, err
	}

	return res.Template, nil
}

// Delete removes a template by its ID
func (svc *TemplateService) Delete(id string) error {
	logging.Trace()
	return svc.BaseService.Delete(
		fmt.Sprintf("/automation-studio/templates/%s", id),
	)
}

// Import imports a template into the system
func (svc *TemplateService) Import(in Template) (*Template, error) {
	logging.Trace()

	body := map[string][]Template{"templates": []Template{in}}

	type Response struct {
		Imported []struct {
			Succcess bool      `json:"success"`
			Message  string    `json:"message"`
			Original *Template `json:"original"`
		} `json:"imported"`
	}

	var res Response

	if err := svc.PostRequest(&Request{
		uri:                "/automation-studio/templates/import",
		body:               &body,
		expectedStatusCode: http.StatusOK,
	}, &res); err != nil {
		return nil, err
	}

	logging.Info("%s", res.Imported[0].Message)

	return res.Imported[0].Original, nil
}

// Export exports a template by its ID
func (svc *TemplateService) Export(id string) (*Template, error) {
	logging.Trace()

	var res *Template
	var uri = fmt.Sprintf("/automation-studio/templates/%s/export", id)

	if err := svc.BaseService.Get(uri, &res); err != nil {
		return nil, err
	}

	return res, nil
}
