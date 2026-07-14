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
	"github.com/mitchellh/mapstructure"
)

type Trigger interface {
}

type ScheduleTrigger struct {
	Id                string                 `json:"_id"`
	Name              string                 `json:"name"`
	Description       string                 `json:"description"`
	Type              string                 `json:"type"`
	Enabled           bool                   `json:"enabled"`
	ActionType        string                 `json:"actionType"`
	ActionId          string                 `json:"actionId"`
	FormId            string                 `json:"formId"`
	FormData          map[string]interface{} `json:"formData"`
	Created           string                 `json:"created"`
	CreatedBy         string                 `json:"createdBy"`
	FormSchemaHash    string                 `json:"formSchemaHash"`
	LastUpdated       string                 `json:"lastUpdated"`
	LastUpdatedBy     string                 `json:"lastUpdatedBy"`
	LegacyWrapper     bool                   `json:"legacyWrapper"`
	Migrationversion  int                    `json:"migrationVersion"`
	FirstRunAt        int                    `json:"firstRunAt"`
	ProcessMissedRuns string                 `json:"processMissedRuns"`
	RepeatUnit        string                 `json:"repeatUnit"`
	RepeatFrequency   int                    `json:"repeatFrequency"`
	RepeatInterval    int                    `json:"repeatInterval"`
	Options           map[string]interface{} `json:"options,omitempty"`
}

type EventTrigger struct {
	Id               string                 `json:"_id"`
	Name             string                 `json:"name"`
	Description      string                 `json:"description"`
	Type             string                 `json:"type"`
	Enabled          bool                   `json:"enabled"`
	ActionType       string                 `json:"actionType"`
	ActionId         string                 `json:"actionId"`
	Created          string                 `json:"created"`
	CreatedBy        string                 `json:"createdBy"`
	LastUpdated      string                 `json:"lastUpdated"`
	LastUpdatedBy    string                 `json:"lastUpdatedBy"`
	LegacyWrapper    bool                   `json:"legacyWrapper"`
	Migrationversion int                    `json:"migrationVersion"`
	Source           string                 `json:"source"`
	Topic            string                 `json:"topic"`
	Schema           map[string]interface{} `json:"schema"`
	Jst              map[string]interface{} `json:"jst"`
	Options          map[string]interface{} `json:"options,omitempty"`
}

type EndpointTrigger struct {
	Id               string                 `json:"_id"`
	Name             string                 `json:"name"`
	Description      string                 `json:"description"`
	Type             string                 `json:"type"`
	Enabled          bool                   `json:"enabled"`
	ActionType       string                 `json:"actionType"`
	ActionId         string                 `json:"actionId"`
	Created          string                 `json:"created"`
	CreatedBy        string                 `json:"createdBy"`
	LastUpdated      string                 `json:"lastUpdated"`
	LastUpdatedBy    string                 `json:"lastUpdatedBy"`
	Migrationversion int                    `json:"migrationVersion"`
	Schema           map[string]interface{} `json:"schema"`
	Jst              map[string]interface{} `json:"jst"`
	RouteName        string                 `json:"routeName"`
	Verb             string                 `json:"verb"`
	Options          map[string]interface{} `json:"options,omitempty"`
}

type ManualTrigger struct {
	Id               string                 `json:"_id"`
	Name             string                 `json:"name"`
	Description      string                 `json:"description"`
	Type             string                 `json:"type"`
	Enabled          bool                   `json:"enabled"`
	ActionType       string                 `json:"actionType"`
	ActionId         string                 `json:"actionId"`
	FormId           string                 `json:"formId"`
	FormData         map[string]interface{} `json:"formData"`
	Created          string                 `json:"created"`
	CreatedBy        string                 `json:"createdBy"`
	FormSchemaHash   string                 `json:"formSchemaHash"`
	LastUpdated      string                 `json:"lastUpdated"`
	LastUpdatedBy    string                 `json:"lastUpdatedBy"`
	LegacyWrapper    bool                   `json:"legacyWrapper"`
	MigrationVersion int                    `json:"migrationVersion"`
	Options          map[string]interface{} `json:"options,omitempty"`
}

func (t ManualTrigger) MarshalJSON() ([]byte, error) {
	res := map[string]interface{}{
		"_id":              t.Id,
		"name":             t.Name,
		"description":      t.Description,
		"type":             t.Type,
		"enabled":          t.Enabled,
		"actionType":       t.ActionType,
		"actionId":         t.ActionId,
		"created":          t.Created,
		"createdBy":        t.CreatedBy,
		"lastUpdated":      t.LastUpdated,
		"lastUpdatedBy":    t.LastUpdatedBy,
		"legacyWrapper":    t.LegacyWrapper,
		"migrationVersion": t.MigrationVersion,
		"formData":         t.FormData,
		"formSchemaHash":   t.FormSchemaHash,
	}

	if t.FormId == "" {
		res["formId"] = nil
	} else {
		res["formId"] = t.FormId
	}

	if t.FormSchemaHash == "" {
		res["formSchemaHash"] = nil
	} else {
		res["formSchemaHash"] = t.FormSchemaHash
	}

	if t.Options != nil {
		res["options"] = t.Options
	}

	return json.Marshal(res)
}

type TriggerService struct {
	BaseService
}

func NewTriggerService(c client.Client) *TriggerService {
	return &TriggerService{BaseService: NewBaseService(c)}
}

func NewEndpointTrigger(name, desc, route, action string) Trigger {
	logging.Trace()
	return EndpointTrigger{
		Name:        name,
		Description: desc,
		ActionId:    action,
		ActionType:  "automations",
		RouteName:   route,
	}
}

func (svc *TriggerService) Create(in Trigger) (Trigger, error) {
	logging.Trace()

	type Response struct {
		Message  string                 `json:"message"`
		Data     map[string]interface{} `json:"data"`
		Metadata map[string]interface{} `json:"metadata"`
	}

	var res Response

	if err := svc.PostRequest(&Request{
		uri:                "/operations-manager/triggers",
		body:               &in,
		expectedStatusCode: http.StatusOK,
	}, &res); err != nil {
		return nil, err
	}

	var trigger Trigger

	b, err := json.Marshal(res.Data)
	if err != nil {
		return nil, err
	}

	switch res.Data["type"].(string) {
	case "endpoint":
		var t EndpointTrigger
		if err := json.Unmarshal(b, &t); err != nil {
			return nil, err
		}
		trigger = t
	}

	return trigger, nil
}

func (svc *TriggerService) DeleteAction(id string) error {
	logging.Trace()
	return svc.BaseService.Delete(
		fmt.Sprintf("/operations-manager/triggers/action/%s", id),
	)
}

func (svc *TriggerService) Import(in Trigger) (*Trigger, error) {
	logging.Trace()

	body := map[string]interface{}{
		"triggers": []Trigger{in},
	}

	type TriggerResponse struct {
		Success bool                   `json:"success"`
		Data    map[string]interface{} `json:"data"`
	}

	type Response struct {
		Message  string                 `json:"message"`
		Metadata map[string]interface{} `json:"metadata"`
		Data     []TriggerResponse      `json:"data"`
	}

	var res Response

	if err := svc.PutRequest(&Request{
		uri:                "/operations-manager/triggers",
		body:               &body,
		expectedStatusCode: http.StatusOK,
	}, &res); err != nil {
		return nil, err
	}

	logging.Info("%s", res.Message)

	data := res.Data[0].Data

	var trigger Trigger

	switch data["type"].(string) {
	case "schedule":
		var t ScheduleTrigger
		if err := mapstructure.Decode(data, &t); err != nil {
			return nil, err
		}
		trigger = t
	case "manual":
		var t ManualTrigger
		if err := mapstructure.Decode(data, &t); err != nil {
			return nil, err
		}
		trigger = t
	case "eventSystem":
		var t EventTrigger
		if err := mapstructure.Decode(data, &t); err != nil {
			return nil, err
		}
		trigger = t
	case "endpoint":
		var t EndpointTrigger
		if err := mapstructure.Decode(data, &t); err != nil {
			return nil, err
		}
		trigger = t
	}

	if trigger == nil {
		return nil, errors.New("error trying to import trigger")
	}

	return &trigger, nil
}
