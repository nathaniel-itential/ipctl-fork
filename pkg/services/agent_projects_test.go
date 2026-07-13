// Copyright 2024 Itential Inc. All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential

package services

import (
	"fmt"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/itential/ipctl/internal/testlib"
	"github.com/stretchr/testify/assert"
)

var (
	agentProjectsGetAllSuccess = "agent-project-service/getall.success.json"
	agentProjectsGetSuccess    = "agent-project-service/get.success.json"
	agentProjectsCreateSuccess = "agent-projects/create.success.json"
	agentProjectsExportSuccess = "agent-project-service/export.success.json"
	agentProjectsImportSuccess = "agent-project-service/import.success.json"
)

func setupAgentProjectService() *AgentProjectService {
	return NewAgentProjectService(testlib.Setup())
}

func TestAgentProjectService_GetAll(t *testing.T) {
	svc := setupAgentProjectService()
	defer testlib.Teardown()

	for _, suite := range fixtureSuites {
		response := testlib.Fixture(filepath.Join(fixtureRoot, suite, agentProjectsGetAllSuccess))
		testlib.AddGetResponseToMux("/agent-project-service/projects", response, 0)

		res, err := svc.GetAll()

		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, 2, len(res))
		assert.Equal(t, "test-agent-project-1", res[0].Name)
		assert.Equal(t, "test-agent-project-2", res[1].Name)
	}
}

func TestAgentProjectService_GetAll_Error(t *testing.T) {
	svc := setupAgentProjectService()
	defer testlib.Teardown()

	testlib.AddGetErrorToMux("/agent-project-service/projects", "", 0)

	res, err := svc.GetAll()

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestAgentProjectService_Get(t *testing.T) {
	svc := setupAgentProjectService()
	defer testlib.Teardown()

	id := "6a1bc2d3e4f0123456789001"

	for _, suite := range fixtureSuites {
		response := testlib.Fixture(filepath.Join(fixtureRoot, suite, agentProjectsGetSuccess))
		testlib.AddGetResponseToMux(fmt.Sprintf("/agent-project-service/projects/%s", id), response, 0)

		res, err := svc.Get(id)

		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, id, res.Id)
		assert.Equal(t, "test-agent-project-1", res.Name)
	}
}

func TestAgentProjectService_Get_EmptyID(t *testing.T) {
	svc := setupAgentProjectService()
	defer testlib.Teardown()

	res, err := svc.Get("")

	assert.NotNil(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "id cannot be empty")
}

func TestAgentProjectService_Get_Error(t *testing.T) {
	svc := setupAgentProjectService()
	defer testlib.Teardown()

	testlib.AddGetErrorToMux("/agent-project-service/projects/bad-id", "", 0)

	res, err := svc.Get("bad-id")

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestAgentProjectService_GetByName(t *testing.T) {
	svc := setupAgentProjectService()
	defer testlib.Teardown()

	for _, suite := range fixtureSuites {
		response := testlib.Fixture(filepath.Join(fixtureRoot, suite, agentProjectsGetAllSuccess))
		testlib.AddGetResponseToMux("/agent-project-service/projects", response, 0)

		res, err := svc.GetByName("test-agent-project-1")

		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "test-agent-project-1", res.Name)
		assert.Equal(t, "6a1bc2d3e4f0123456789001", res.Id)
	}
}

func TestAgentProjectService_GetByName_NotFound(t *testing.T) {
	svc := setupAgentProjectService()
	defer testlib.Teardown()

	for _, suite := range fixtureSuites {
		response := testlib.Fixture(filepath.Join(fixtureRoot, suite, agentProjectsGetAllSuccess))
		testlib.AddGetResponseToMux("/agent-project-service/projects", response, 0)

		res, err := svc.GetByName("nonexistent-project")

		assert.NotNil(t, err)
		assert.Nil(t, res)
		assert.Equal(t, "agent project not found", err.Error())
	}
}

func TestAgentProjectService_GetByName_Error(t *testing.T) {
	svc := setupAgentProjectService()
	defer testlib.Teardown()

	testlib.AddGetErrorToMux("/agent-project-service/projects", "", 0)

	res, err := svc.GetByName("test-agent-project-1")

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestAgentProjectService_Export(t *testing.T) {
	svc := setupAgentProjectService()
	defer testlib.Teardown()

	id := "6a1bc2d3e4f0123456789001"

	for _, suite := range fixtureSuites {
		response := testlib.Fixture(filepath.Join(fixtureRoot, suite, agentProjectsExportSuccess))
		testlib.AddGetResponseToMux(
			fmt.Sprintf("/agent-project-service/project-bundles/%s/export", id),
			response, 0,
		)

		res, err := svc.Export(id)

		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, id, res.Id)
		assert.Equal(t, "test-agent-project-1", res.Name)
		assert.Equal(t, 1, len(res.Agents))
	}
}

func TestAgentProjectService_Export_EmptyID(t *testing.T) {
	svc := setupAgentProjectService()
	defer testlib.Teardown()

	res, err := svc.Export("")

	assert.NotNil(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "id cannot be empty")
}

func TestAgentProjectService_Export_Error(t *testing.T) {
	svc := setupAgentProjectService()
	defer testlib.Teardown()

	testlib.AddGetErrorToMux("/agent-project-service/project-bundles/bad-id/export", "", 0)

	res, err := svc.Export("bad-id")

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestAgentProjectService_Import(t *testing.T) {
	svc := setupAgentProjectService()
	defer testlib.Teardown()

	for _, suite := range fixtureSuites {
		response := testlib.Fixture(filepath.Join(fixtureRoot, suite, agentProjectsImportSuccess))
		testlib.AddPostResponseToMux("/agent-project-service/project-bundles/import", response, http.StatusOK)

		bundle := AgentProjectBundle{
			Name:        "imported-agent-project",
			Description: "An imported agent project",
			Agents:      []map[string]interface{}{},
		}

		res, err := svc.Import(bundle, "keep-both")

		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "imported-agent-project", res.Name)
	}
}

func TestAgentProjectService_Import_Error(t *testing.T) {
	svc := setupAgentProjectService()
	defer testlib.Teardown()

	testlib.AddPostErrorToMux("/agent-project-service/project-bundles/import", "", 0)

	bundle := AgentProjectBundle{Name: "test"}

	res, err := svc.Import(bundle, "keep-both")

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestAgentProjectService_Create(t *testing.T) {
	svc := setupAgentProjectService()
	defer testlib.Teardown()

	for _, suite := range fixtureSuites {
		response := testlib.Fixture(filepath.Join(fixtureRoot, suite, agentProjectsCreateSuccess))
		testlib.AddPostResponseToMux("/agent-project-service/projects", response, http.StatusOK)

		res, err := svc.Create("Test Agent Project", "A test agent project")

		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "Test Agent Project", res.Name)
	}
}

func TestAgentProjectService_Create_Error(t *testing.T) {
	svc := setupAgentProjectService()
	defer testlib.Teardown()

	testlib.AddPostErrorToMux("/agent-project-service/projects", "", 0)

	res, err := svc.Create("Test Agent Project", "A test agent project")

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestAgentProjectService_Create_EmptyName(t *testing.T) {
	svc := setupAgentProjectService()
	defer testlib.Teardown()

	res, err := svc.Create("", "A test agent project")

	assert.NotNil(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "name cannot be empty")
}
