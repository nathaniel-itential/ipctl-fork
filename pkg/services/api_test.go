// Copyright 2024 Itential Inc. All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential

package services

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/itential/ipctl/internal/testlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupApiService() *ApiService {
	return NewApiService(testlib.Setup())
}

func TestApiService_Get_NoParams(t *testing.T) {
	svc := setupApiService()
	defer testlib.Teardown()

	testlib.AddGetResponseToMux("/automation-studio/projects", `{"data":[]}`, http.StatusOK)

	result, err := svc.Get("/automation-studio/projects", http.StatusOK)

	require.NoError(t, err)
	assert.Contains(t, result, "data")
}

func TestApiService_Get_WithQueryParamsInURI(t *testing.T) {
	svc := setupApiService()
	defer testlib.Teardown()

	var received *http.Request
	testlib.AddHandlerToMux("/automation-studio/projects", func(w http.ResponseWriter, r *http.Request) {
		received = r
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"data":[]}`)
	})

	_, err := svc.Get("/automation-studio/projects?limit=10", http.StatusOK)

	require.NoError(t, err)
	require.NotNil(t, received, "handler was not called — path was likely malformed")
	assert.Equal(t, "/automation-studio/projects", received.URL.Path, "path should not contain encoded query string")
	assert.Equal(t, "10", received.URL.Query().Get("limit"), "limit param should arrive as a query string parameter")
}

func TestApiService_Get_WithMultipleQueryParamsInURI(t *testing.T) {
	svc := setupApiService()
	defer testlib.Teardown()

	var received *http.Request
	testlib.AddHandlerToMux("/automation-studio/workflows", func(w http.ResponseWriter, r *http.Request) {
		received = r
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"data":[]}`)
	})

	_, err := svc.Get("/automation-studio/workflows?limit=25&offset=50", http.StatusOK)

	require.NoError(t, err)
	require.NotNil(t, received, "handler was not called — path was likely malformed")
	assert.Equal(t, "/automation-studio/workflows", received.URL.Path)
	assert.Equal(t, "25", received.URL.Query().Get("limit"))
	assert.Equal(t, "50", received.URL.Query().Get("offset"))
}

func TestApiService_Get_StatusCodeError(t *testing.T) {
	svc := setupApiService()
	defer testlib.Teardown()

	testlib.AddGetErrorToMux("/automation-studio/projects", `{"message":"not found"}`, http.StatusNotFound)

	_, err := svc.Get("/automation-studio/projects", http.StatusOK)

	assert.Error(t, err)
}

func TestApiService_Delete_WithQueryParamsInURI(t *testing.T) {
	svc := setupApiService()
	defer testlib.Teardown()

	var received *http.Request
	testlib.AddHandlerToMux("/operations-manager/triggers", func(w http.ResponseWriter, r *http.Request) {
		received = r
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"data":{}}`)
	})

	_, err := svc.Delete("/operations-manager/triggers?id=abc123", http.StatusOK)

	require.NoError(t, err)
	require.NotNil(t, received, "handler was not called — path was likely malformed")
	assert.Equal(t, "/operations-manager/triggers", received.URL.Path)
	assert.Equal(t, "abc123", received.URL.Query().Get("id"))
}
