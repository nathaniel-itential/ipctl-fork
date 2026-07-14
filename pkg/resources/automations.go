// Copyright 2024 Itential Inc. All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential

package resources

import (
	"encoding/json"
	"fmt"

	"github.com/itential/ipctl/internal/logging"
	"github.com/itential/ipctl/pkg/services"
)

// AutomationResource provides business logic for automation operations.
type AutomationResource struct {
	BaseResource
	service services.AutomationServicer
}

// NewAutomationResource creates a new AutomationResource with the given service.
func NewAutomationResource(svc services.AutomationServicer) AutomationResourcer {
	return &AutomationResource{
		BaseResource: NewBaseResource(),
		service:      svc,
	}
}

// GetAll retrieves all automations from the API.
// This is a pass-through to the service layer for pure API access.
func (r *AutomationResource) GetAll() ([]*services.Automation, error) {
	return r.service.GetAll()
}

// Get retrieves a specific automation by ID from the API.
// This is a pass-through to the service layer for pure API access.
func (r *AutomationResource) Get(id string) (*services.Automation, error) {
	return r.service.Get(id)
}

// Create creates a new automation.
// This is a pass-through to the service layer for pure API access.
func (r *AutomationResource) Create(in services.Automation) (*services.Automation, error) {
	return r.service.Create(in)
}

// Delete removes an automation by its identifier.
// This is a pass-through to the service layer for pure API access.
func (r *AutomationResource) Delete(id string) error {
	return r.service.Delete(id)
}

// ImportTransformed imports automations using pre-transformed data.
// This is a pass-through to the service layer for pure API access.
func (r *AutomationResource) ImportTransformed(automations []any) (*services.Automation, error) {
	return r.service.ImportTransformed(automations)
}

// GetByName retrieves an automation by name using client-side filtering.
// It fetches all automations and searches for a matching name.
func (r *AutomationResource) GetByName(name string) (*services.Automation, error) {
	logging.Trace()

	automations, err := r.service.GetAll()
	if err != nil {
		return nil, err
	}

	for _, automation := range automations {
		if automation.Name == name {
			return automation, nil
		}
	}

	return nil, fmt.Errorf("automation not found")
}

// Clear deletes all automations from the server.
// This is a bulk operation that orchestrates multiple delete calls.
func (r *AutomationResource) Clear() error {
	logging.Trace()

	automations, err := r.service.GetAll()
	if err != nil {
		return err
	}

	return DeleteAll(automations, func(a *services.Automation) string {
		return a.Id
	}, r.service.Delete)
}

// Import imports an automation with business rule validation and data transformation.
// Validates GBAC rules and ensures triggers are properly formatted before import.
func (r *AutomationResource) Import(in services.Automation) (*services.Automation, error) {
	logging.Trace()

	// Business rule validation: write group must be configured when read group is present
	if err := ValidateGbacRules(in.Gbac.Read, in.Gbac.Write); err != nil {
		return nil, err
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

	return r.service.ImportTransformed(automations)
}

// Export exports an automation with trigger type transformation.
// Handles polymorphic trigger types and converts them to proper typed structures.
func (r *AutomationResource) Export(id string) (*services.Automation, error) {
	logging.Trace()

	automation, err := r.service.Export(id)
	if err != nil {
		return nil, err
	}

	// Transform polymorphic trigger data into typed structures
	triggers := automation.Triggers
	automation.Triggers = []services.Trigger{}

	for _, ele := range triggers {
		trigger, err := r.transformTrigger(ele)
		if err != nil {
			return nil, err
		}
		automation.Triggers = append(automation.Triggers, trigger)
	}

	return automation, nil
}

// transformTrigger converts a generic trigger interface to a specific typed trigger.
func (r *AutomationResource) transformTrigger(ele interface{}) (services.Trigger, error) {
	b, err := json.Marshal(ele.(map[string]interface{}))
	if err != nil {
		logging.Fatal(err, "error trying to marshal trigger data")
		return nil, err
	}

	triggerType := ele.(map[string]interface{})["type"].(string)

	switch triggerType {
	case "endpoint":
		var t services.EndpointTrigger
		if err := json.Unmarshal(b, &t); err != nil {
			logging.Fatal(err, "error trying to decode endpoint trigger")
			return nil, err
		}
		return t, nil
	case "eventSystem":
		var t services.EventTrigger
		if err := json.Unmarshal(b, &t); err != nil {
			logging.Fatal(err, "error trying to decode event trigger")
			return nil, err
		}
		return t, nil
	case "manual":
		var t services.ManualTrigger
		if err := json.Unmarshal(b, &t); err != nil {
			logging.Fatal(err, "error trying to decode manual trigger")
			return nil, err
		}
		return t, nil
	case "schedule":
		var t services.ScheduleTrigger
		if err := json.Unmarshal(b, &t); err != nil {
			logging.Fatal(err, "error trying to decode schedule trigger")
			return nil, err
		}
		return t, nil
	default:
		var t map[string]interface{}
		if err := json.Unmarshal(b, &t); err != nil {
			return nil, fmt.Errorf("error trying to decode trigger of type %s: %w", triggerType, err)
		}
		return t, nil
	}
}
