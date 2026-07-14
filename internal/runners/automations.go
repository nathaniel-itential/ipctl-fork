// Copyright 2024 Itential Inc. All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential

package runners

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/itential/ipctl/internal/config"
	"github.com/itential/ipctl/internal/flags"
	"github.com/itential/ipctl/internal/logging"
	"github.com/itential/ipctl/internal/utils"
	"github.com/itential/ipctl/pkg/client"
	"github.com/itential/ipctl/pkg/resources"
	"github.com/itential/ipctl/pkg/services"
	"github.com/itential/ipctl/pkg/validators"
)

type AutomationRunner struct {
	BaseRunner
	client    client.Client
	resource  resources.AutomationResourcer
	workflows *services.AutomationService
	triggers  *services.TriggerService
}

func NewAutomationRunner(c client.Client, cfg config.Provider) *AutomationRunner {
	return &AutomationRunner{
		BaseRunner: NewBaseRunner(c, cfg),
		client:     c,
		resource:   resources.NewAutomationResource(services.NewAutomationService(c)),
		workflows:  services.NewAutomationService(c),
		triggers:   services.NewTriggerService(c),
	}
}

//////////////////////////////////////////////////////////////////////////////
// Reader Interface
//

// Get is the implementation of the command `get automations`
func (r *AutomationRunner) Get(in Request) (*Response, error) {
	logging.Trace()

	automations, err := r.resource.GetAll()
	if err != nil {
		return nil, err
	}

	var display = []string{"NAME\tDESCRIPTION"}
	for _, ele := range automations {
		desc := strings.Replace(ele.Description, "\n", " ", -1)
		display = append(display, fmt.Sprintf("%s\t%s", ele.Name, desc))
	}

	return &Response{
		Keys:   []string{"name", "description"},
		Object: automations,
	}, nil
}

func (r *AutomationRunner) Describe(in Request) (*Response, error) {
	logging.Trace()

	name := in.Args[0]

	automation, err := r.resource.GetByName(name)
	if err != nil {
		return nil, err
	}

	res, err := r.resource.Export(automation.Id)
	if err != nil {
		return nil, err
	}

	var triggers []string

	for _, ele := range res.Triggers {
		m, err := toMap(ele)
		if err != nil {
			return nil, err
		}
		triggers = append(triggers,
			fmt.Sprintf("- Name: %s, Type: %s", m["name"].(string), m["type"].(string)),
		)
	}

	if len(triggers) == 0 {
		triggers = append(triggers, "No triggers configured")
	}

	desc := []string{"\nDescription:"}
	if res.Description != "" {
		desc = append(desc, fmt.Sprintf("%s\n", res.Description))
	}

	output := []string{
		fmt.Sprintf("Name: %s (%s)", res.Name, res.Id),
		strings.Join(desc, "\n"),
		fmt.Sprintf("\nComponent Name: %s (%s)", res.ComponentName, automation.ComponentId),
		fmt.Sprintf("Component Type: %s", res.ComponentType),
		fmt.Sprintf("\nTiggers"),
		strings.Join(triggers, "\n"),
		fmt.Sprintf("\nCreated: %s, By: %s", res.Created, res.CreatedBy),
		fmt.Sprintf("Updated: %s, By: %s", res.LastUpdated, res.LastUpdatedBy),
	}

	return &Response{
		Text:   strings.Join(output, "\n"),
		Object: res,
	}, nil
}

//////////////////////////////////////////////////////////////////////////////
// Writer Interface
//

func (r *AutomationRunner) Create(in Request) (*Response, error) {
	logging.Trace()

	name := in.Args[0]

	var options flags.AutomationCreateOptions
	utils.LoadObject(in.Options, &options)

	if options.Replace {
		existing, err := r.resource.GetByName(name)

		if existing != nil {
			if err := r.resource.Delete(existing.Id); err != nil {
				return nil, err
			}
		} else if err != nil {
			if err.Error() != "automation not found" {
				return nil, err
			}
		}
	}

	res, err := r.resource.Create(services.NewAutomation(name, options.Description))
	if err != nil {
		return nil, err
	}

	return &Response{
		Text:   fmt.Sprintf("Successfully created automation `%s`", res.Name),
		Object: res,
	}, nil
}

func (r *AutomationRunner) Delete(in Request) (*Response, error) {
	logging.Trace()

	name := in.Args[0]

	automations, err := r.resource.GetAll()
	if err != nil {
		return nil, err
	}

	var selected *services.Automation

	for _, ele := range automations {
		if ele.Name == name {
			selected = ele
			break
		}
	}

	if selected != nil {
		if err := r.resource.Delete(selected.Id); err != nil {
			return nil, err
		}
	}

	return &Response{
		Text: fmt.Sprintf("Successfully deleted automation `%s`", name),
	}, nil
}

// Clear implements the `clear automations` command
func (r *AutomationRunner) Clear(in Request) (*Response, error) {
	logging.Trace()

	automations, err := r.resource.GetAll()
	if err != nil {
		return nil, err
	}

	for _, ele := range automations {
		if err := r.resource.Delete(ele.Id); err != nil {
			return nil, err
		}
	}

	return &Response{
		Text: fmt.Sprintf("Deleted %v automations(s)", len(automations)),
	}, nil
}

//////////////////////////////////////////////////////////////////////////////
// Copier Interface
//

func (r *AutomationRunner) Copy(in Request) (*Response, error) {
	logging.Trace()

	res, err := Copy(CopyRequest{Request: in, Type: "automation"}, r)
	if err != nil {
		return nil, err
	}

	return &Response{
		Text: fmt.Sprintf("Successfully copied automation `%s` from `%s` to `%s`", res.Name, res.From, res.To),
	}, nil
}

func (r *AutomationRunner) CopyFrom(profile, name string) (any, error) {
	logging.Trace()

	client, cancel, err := NewClient(profile, r.config)
	if err != nil {
		return nil, err
	}
	defer cancel()

	svc := services.NewAutomationService(client)
	res := resources.NewAutomationResource(svc)

	automation, err := res.GetByName(name)
	if err != nil {
		return nil, err
	}

	exported, err := res.Export(automation.Id)
	if err != nil {
		return nil, err
	}

	return *exported, err
}

func (r *AutomationRunner) CopyTo(profile string, in any, replace bool) (any, error) {
	logging.Trace()

	client, cancel, err := NewClient(profile, r.config)
	if err != nil {
		return nil, err
	}
	defer cancel()

	svc := services.NewAutomationService(client)

	automation := in.(services.Automation)

	if err := validators.NewAutomationValidator(r.client).CanImport(automation); err != nil {
		if err := r.checkImportValidationError(err, automation.Name, replace); err != nil {
			return nil, err
		}
	}

	automationRes := resources.NewAutomationResource(svc)
	res, err := automationRes.Import(automation)

	if err != nil {
		return nil, errors.New(r.formatImportErrorMessage(err))
	}

	return res, err
}

//////////////////////////////////////////////////////////////////////////////
// Importer Interface
//

// Import implements the `import automation <name>` command
func (r *AutomationRunner) Import(in Request) (*Response, error) {
	logging.Trace()

	var automation services.Automation

	if err := importUnmarshalFromRequest(in, &automation); err != nil {
		return nil, err
	}

	common := in.Common.(*flags.AssetImportCommon)

	res, err := r.importAutomation(automation, common.Replace)
	if err != nil {
		return nil, err
	}

	return &Response{
		Text:   fmt.Sprintf("Successfully imported automation `%s` with %v trigger(s)", res.Name, len(automation.Triggers)),
		Object: res,
	}, nil
}

//////////////////////////////////////////////////////////////////////////////
// Exporter Interface
//

// Export implements the `export automation <name>` command
func (r *AutomationRunner) Export(in Request) (*Response, error) {
	logging.Trace()

	name := in.Args[0]

	automation, err := r.resource.GetByName(name)
	if err != nil {
		return nil, err
	}

	res, err := r.resource.Export(automation.Id)
	if err != nil {
		return nil, err
	}

	fn := fmt.Sprintf("%s.automation.json", name)

	if err := exportAssetFromRequest(in, res, fn); err != nil {
		return nil, err
	}

	return &Response{
		Text: fmt.Sprintf("Successfully exported automation `%s`", name),
	}, nil

}

//////////////////////////////////////////////////////////////////////////////
// Dumper Interface
//

// Dump implements the `dump automations...` command
func (r *AutomationRunner) Dump(in Request) (*Response, error) {
	logging.Trace()

	res, err := r.resource.GetAll()
	if err != nil {
		return nil, err
	}

	var assets = map[string]interface{}{}

	for _, ele := range res {
		automation, err := r.resource.Export(ele.Id)
		if err != nil {
			return nil, err
		}

		key := fmt.Sprintf("%s.automation.json", automation.Name)
		assets[key] = automation
	}

	if err := dumpAssets(in, assets); err != nil {
		return nil, err
	}

	return &Response{
		Text: fmt.Sprintf("Dumped %v automation(s)", len(assets)),
	}, nil
}

//////////////////////////////////////////////////////////////////////////////
// Loader Interface
//

// Load implements the `load automations ...` command
func (r *AutomationRunner) Load(in Request) (*Response, error) {
	logging.Trace()

	elements, err := loadAssets(in)
	if err != nil {
		return nil, err
	}

	var loaded int
	var skipped int

	var output []string

	for fn, ele := range elements {
		var automation services.Automation

		if err := loadUnmarshalAsset(ele, &automation); err != nil {
			output = append(output, fmt.Sprintf("Failed to load automation from `%s`, skipping", fn))
			skipped++
		} else {
			_, err := r.importAutomation(automation, false)
			if err != nil {
				if !strings.HasSuffix(err.Error(), "already exists on the server") {
					return nil, err
				}
				output = append(output, fmt.Sprintf("Skipping `%s`, automation `%s` already exists", fn, automation.Name))
				skipped++
			} else {
				output = append(output, fmt.Sprintf("Loaded automation `%s` successfully from `%s`", automation.Name, fn))
				loaded++
			}
		}
	}

	output = append(output, fmt.Sprintf(
		"\nSuccessfully loaded %v and skipped %v files from `%s`", loaded, skipped, in.Args[0],
	))

	return &Response{
		Text: strings.Join(output, "\n"),
	}, nil
}

//////////////////////////////////////////////////////////////////////////////
// Private functions
//

func (r *AutomationRunner) importAutomation(in services.Automation, replace bool) (*services.Automation, error) {
	logging.Trace()

	if err := validators.NewAutomationValidator(r.client).CanImport(in); err != nil {
		if err := r.checkImportValidationError(err, in.Name, replace); err != nil {
			return nil, err
		}
	}

	triggers, err := r.updateTriggers(in)
	if err != nil {
		return nil, err
	}
	in.Triggers = triggers

	return r.resource.Import(in)
}

func (r *AutomationRunner) formatImportErrorMessage(e error) string {
	logging.Trace()

	type ResponseError struct {
		Message  string `json:"message"`
		Data     any    `json:"data"`
		Metadata struct {
			Errors []struct {
				Success bool                   `json:"success"`
				Reason  string                 `json:"reason"`
				Data    map[string]interface{} `json:"data"`
			} `json:"errors"`
		} `json:"metadata"`
	}

	var res ResponseError

	if err := json.Unmarshal([]byte(e.Error()), &res); err != nil {
		logging.Fatal(err, "failed to unmarshal error message")
	}

	var output = []string{
		fmt.Sprintf("%s (See details below)", res.Message),
	}

	for _, ele := range res.Metadata.Errors {
		output = append(output, fmt.Sprintf("- %s", ele.Reason))
	}

	return strings.Join(output, "\n")

}

func (r *AutomationRunner) updateTriggers(in services.Automation) ([]services.Trigger, error) {
	logging.Trace()

	data, err := toMap(in)
	if err != nil {
		return nil, err
	}

	var triggers []services.Trigger

	if value, exists := data["triggers"]; exists {
		if value != nil {
			for _, ele := range value.([]interface{}) {
				b, err := json.Marshal(ele)
				if err != nil {
					return nil, err
				}

				item := ele.(map[string]interface{})

				switch item["type"].(string) {
				case "endpoint":
					var t services.EndpointTrigger
					if err := json.Unmarshal(b, &t); err != nil {
						return nil, err
					}
					triggers = append(triggers, t)
				case "eventSystem":
					var t services.EventTrigger
					if err := json.Unmarshal(b, &t); err != nil {
						return nil, err
					}
					triggers = append(triggers, t)
				case "manual":
					var t services.ManualTrigger
					if err := json.Unmarshal(b, &t); err != nil {
						return nil, err
					}
					triggers = append(triggers, t)
				case "schedule":
					var t services.ScheduleTrigger
					if err := json.Unmarshal(b, &t); err != nil {
						return nil, err
					}
					triggers = append(triggers, t)
				default:
					var t map[string]interface{}
					if err := json.Unmarshal(b, &t); err != nil {
						return nil, err
					}
					triggers = append(triggers, t)
				}
			}
		}
	}
	return triggers, nil
}

func (r *AutomationRunner) checkImportValidationError(e error, name string, replace bool) error {
	if e.Error() == "automation already exists" {
		if replace {
			existing, err := r.resource.GetByName(name)
			if err != nil {
				return err
			}
			if err := r.resource.Delete(existing.Id); err != nil {
				return err
			}
		} else {
			return errors.New(
				fmt.Sprintf("automation `%s` already exists on the server", name),
			)
		}
	}
	return nil
}
