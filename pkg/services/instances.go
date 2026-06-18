// Copyright 2024 Itential Inc. All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential

package services

import (
	"fmt"

	"github.com/itential/ipctl/internal/logging"
	"github.com/itential/ipctl/pkg/client"
)

type InstanceOperation struct {
	Message  string     `json:"message"`
	Data     []Instance `json:"data"`
	Metadata Metadata   `json:"metadata"`
}

type LastAction struct {
	Id          string `json:"_id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	ExecutionId string `json:"executionId"`
}

type Instance struct {
	Id           string                 `json:"_id"`
	ModelId      string                 `json:"modelId"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	InstanceData map[string]interface{} `json:"instanceData"`
	LastAction   LastAction             `json:"lastAction"`
	/*
		Created       string                 `json:"created"`
		CreatedBy     string                 `json:"createdBy"`
		LastUpdated   string                 `json:"lastUpdated"`
		LastUpdatedBy string                 `json:"lastUpdatedBy"`
	*/
}

type InstanceService struct {
	BaseService
}

func NewInstanceService(c client.Client) *InstanceService {
	return &InstanceService{BaseService: NewBaseService(c)}
}

/*
func (svc *InstanceService) Get(modelId, instanceId string) (*Instance, error) {
	logging.Trace()

	var response map[string]interface{}
	resp, err := Do(&Request{
		client:   svc.client,
		method:   http.MethodGet,
		uri:      fmt.Sprintf("/lifecycle-manager/resources/%s/instances/%s", modelId, instanceId),
		response: &response,
	})

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New(string(resp.Body))
	}

	var instance *Instance
	if err := Unmarshal(response["data"].(map[string]interface{}), &instance); err != nil {
		return nil, err
	}

	return instance, nil
}
*/

func (svc *InstanceService) GetAll(modelId string) ([]Instance, error) {
	logging.Trace()

	type Response struct {
		Message  string     `json:"message"`
		Metadata Metadata   `json:"metadata"`
		Data     []Instance `json:"data"`
	}

	var uri = fmt.Sprintf("/lifecycle-manager/resources/%s/instances", modelId)

	var results []Instance
	var message string

	var limit = 100
	var skip = 0

	for {
		// Declare res inside the loop so each page decodes into a freshly
		// allocated struct. Reusing a single res across pages lets
		// encoding/json merge map fields and reuse slice backing arrays,
		// bleeding fields from one page's elements into the next.
		var res Response
		if err := svc.GetRequest(&Request{
			uri:    uri,
			params: &QueryParams{Limit: limit, Skip: skip, Raw: map[string]string{"include-deleted": "true"}},
		}, &res); err != nil {
			return nil, err
		}

		results = append(results, res.Data...)
		message = res.Message

		if len(results) == res.Metadata.Total {
			break
		}

		skip += limit
	}

	logging.Info("%s", message)

	return results, nil
}
