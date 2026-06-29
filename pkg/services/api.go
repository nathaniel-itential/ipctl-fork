// Copyright 2024 Itential Inc. All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential

package services

import (
	"net/http"
	"net/url"

	"github.com/itential/ipctl/internal/logging"
	"github.com/itential/ipctl/pkg/client"
)

type ApiService struct {
	client client.Client
}

func NewApiService(c client.Client) *ApiService {
	return &ApiService{client: c}
}

func (svc *ApiService) request(m string, uri string, body map[string]interface{}, expectedStatusCode int) (string, error) {
	logging.Trace()

	u, err := url.Parse(uri)
	if err != nil {
		return "", err
	}

	values, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return "", err
	}

	params := &RawParams{Values: values}

	req := &Request{
		client:             svc.client,
		method:             m,
		uri:                u.Path,
		params:             params,
		expectedStatusCode: expectedStatusCode,
	}

	if body != nil {
		req.body = &body
	}

	res, err := Do(req)

	if err != nil {
		return "", err
	}

	return string(res.Body), nil
}

func (svc *ApiService) Get(url string, expectedStatusCode int) (string, error) {
	logging.Trace()
	return svc.request(http.MethodGet, url, nil, expectedStatusCode)
}

func (svc *ApiService) Post(url string, body map[string]interface{}, expectedStatusCode int) (string, error) {
	logging.Trace()
	return svc.request(http.MethodPost, url, body, expectedStatusCode)
}

func (svc *ApiService) Put(url string, body map[string]interface{}, expectedStatusCode int) (string, error) {
	logging.Trace()
	return svc.request(http.MethodPut, url, body, expectedStatusCode)
}

func (svc *ApiService) Delete(url string, expectedStatusCode int) (string, error) {
	logging.Trace()
	return svc.request(http.MethodDelete, url, nil, expectedStatusCode)
}

func (svc *ApiService) Patch(url string, body map[string]interface{}, expectedStatusCode int) (string, error) {
	logging.Trace()
	return svc.request(http.MethodPatch, url, nil, expectedStatusCode)
}
