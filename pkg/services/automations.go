// Copyright 2024 Itential Inc. All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential

package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/itential/ipctl/internal/logging"
	"github.com/itential/ipctl/pkg/client"
)

// automationCollection represents a collection of automations returned by the API
type automationCollection struct {
	Message  string       `json:"message"`
	Data     []Automation `json:"data"`
	Metadata Metadata     `json:"metadata"`
}

// AutomationGbacEntry represents a GBAC (Group-Based Access Control) entry for automations
type AutomationGbacEntry struct {
	Name        string `json:"name"`
	Provenance  string `json:"provenance"`
	Description string `json:"description"`
}

// AutomationGbac represents GBAC permissions for an automation
type AutomationGbac struct {
	Write []interface{} `json:"write"`
	Read  []interface{} `json:"read"`
}

// Automation represents an automation in the Itential Platform
type Automation struct {
	Id            string `json:"_id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	ComponentName string `json:"componentName"`
	ComponentType string `json:"componentType"`

	// ComponentId field does not exist when exporting the automation but it
	// does exist when getting it
	ComponentId string `json:"componentId,omitempty"`

	Gbac          AutomationGbac `json:"gbac"`
	Created       string         `json:"created"`
	CreatedBy     string         `json:"createdBy"`
	LastUpdated   string         `json:"lastUpdated"`
	LastUpdatedBy string         `json:"lastUpdatedBy"`

	// Triggers does not exist when getting the autoatmion but it does exists
	// when exporting it
	Triggers []Trigger `json:"triggers,omitempty"`
}

// AutomationService provides methods for managing automations
type AutomationService struct {
	BaseService
}

// NewAutomation creates a new Automation instance with the given name and description
func NewAutomation(name, desc string) Automation {
	logging.Trace()
	return Automation{
		Name:          name,
		Description:   desc,
		ComponentType: "workflows",
	}
}

// NewAutomationService creates a new AutomationService with the given client
func NewAutomationService(c client.Client) *AutomationService {
	return &AutomationService{BaseService: NewBaseService(c)}
}

// Get implements `GET /operations-manager/automations/{id}`
func (svc *AutomationService) Get(id string) (*Automation, error) {
	logging.Trace()

	type Response struct {
		Message string      `json:"message"`
		Data    *Automation `json:"data"`
	}

	var res Response
	var uri = fmt.Sprintf("/operations-manager/automations/%s", id)

	if err := svc.BaseService.Get(uri, &res); err != nil {
		return nil, err
	}

	logging.Info("%s", res.Message)

	return res.Data, nil
}

// Create implements `POST /operations-manager/automations`
func (svc *AutomationService) Create(in Automation) (*Automation, error) {
	logging.Trace()

	body := map[string]interface{}{
		"name":          in.Name,
		"description":   in.Description,
		"componentType": in.ComponentType,
	}

	type Response struct {
		Message  string                 `json:"message"`
		Data     *Automation            `json:"data"`
		Metadata map[string]interface{} `json:"metadata"`
	}

	var res Response

	if err := svc.PostRequest(&Request{
		uri:                "/operations-manager/automations",
		body:               &body,
		expectedStatusCode: http.StatusOK,
	}, &res); err != nil {
		return nil, err
	}

	logging.Info("%s", res.Message)

	return res.Data, nil
}

// Delete implements `DELETE /operations-manager/automations/{id}`
func (svc *AutomationService) Delete(id string) error {
	logging.Trace()
	return svc.BaseService.Delete(fmt.Sprintf("/operations-manager/automations/%s", id))
}

// GetAll implements `GET /operations-manager/automations`
func (svc *AutomationService) GetAll() ([]*Automation, error) {
	logging.Trace()

	type Response struct {
		Message  string        `json:"message"`
		Data     []*Automation `json:"data"`
		Metadata Metadata      `json:"metadata"`
	}

	var automations []*Automation

	var limit = 100
	var skip = 0

	for {
		// Declare res inside the loop so each page decodes into a freshly
		// allocated struct. Reusing a single res across pages lets
		// encoding/json merge map fields and reuse slice backing arrays,
		// bleeding fields from one page's elements into the next.
		var res Response
		if err := svc.GetRequest(&Request{
			uri:    "/operations-manager/automations",
			params: &QueryParams{Limit: limit, Skip: skip},
		}, &res); err != nil {
			return nil, err
		}

		automations = append(automations, res.Data...)

		if len(automations) == res.Metadata.Total {
			break
		}

		skip += limit
	}

	logging.Info("Found %v automations", len(automations))

	return automations, nil
}

// GetByName retrieves an automation by name using client-side filtering.
// DEPRECATED: Business logic method - prefer using resources.AutomationResource.GetByName
func (svc *AutomationService) GetByName(name string) (*Automation, error) {
	logging.Trace()

	automations, err := svc.GetAll()
	if err != nil {
		return nil, err
	}

	for _, automation := range automations {
		if automation.Name == name {
			return automation, nil
		}
	}

	return nil, errors.New("automation not found")
}

// Import imports an automation with business rule validation and data transformation.
// DEPRECATED: Business logic method - prefer using resources.AutomationResource.Import
func (svc *AutomationService) Import(in Automation) (*Automation, error) {
	logging.Trace()

	// Business rule validation: write group must be configured when read group is present
	if len(in.Gbac.Read) > 0 && len(in.Gbac.Write) == 0 {
		return nil, errors.New("write group must be configured, when read group present")
	}

	var automations []any

	// Transform automation data: ensure triggers array exists even if empty
	if len(in.Triggers) == 0 {
		b, err := json.Marshal(in)
		if err != nil {
			return nil, err
		}

		var item map[string]interface{}
		if err := json.Unmarshal(b, &item); err != nil {
			return nil, err
		}

		item["triggers"] = []any{}
		automations = append(automations, item)
	} else {
		automations = append(automations, in)
	}

	return svc.ImportTransformed(automations)
}

// Clear deletes all automations from the server.
// DEPRECATED: Business logic method - prefer using resources.AutomationResource.Clear
func (svc *AutomationService) Clear() error {
	logging.Trace()

	automations, err := svc.GetAll()
	if err != nil {
		return err
	}

	for _, automation := range automations {
		if err := svc.Delete(automation.Id); err != nil {
			return err
		}
	}

	return nil
}

// ImportTransformed imports pre-transformed automation data.
// The automations parameter should contain properly validated and transformed automation data.
func (svc *AutomationService) ImportTransformed(automations []any) (*Automation, error) {
	logging.Trace()

	body := map[string][]any{
		"automations": automations,
	}

	type Data struct {
		Success bool       `json:"success"`
		Data    Automation `json:"data"`
	}

	type Response struct {
		Data     []Data   `json:"data"`
		Message  string   `json:"message"`
		Metadata Metadata `json:"metadata"`
	}

	var res Response

	if err := svc.PutRequest(&Request{
		uri:  "/operations-manager/automations",
		body: &body,
	}, &res); err != nil {
		return nil, err
	}

	logging.Info("%s", res.Message)

	return &res.Data[0].Data, nil
}

// Export exports an automation by ID, including its triggers in raw format.
// Returns the automation data as received from the API without trigger transformation.
func (svc *AutomationService) Export(id string) (*Automation, error) {
	logging.Trace()

	type Response struct {
		Data     *Automation `json:"data"`
		Message  string      `json:"message"`
		Metadata Metadata    `json:"metadata"`
	}

	var res Response
	var uri = fmt.Sprintf("/operations-manager/automations/%s/export", id)

	if err := svc.BaseService.Get(uri, &res); err != nil {
		return nil, err
	}

	logging.Info("%s", res.Message)

	return res.Data, nil
}
