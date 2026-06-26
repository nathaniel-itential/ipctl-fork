// Copyright 2024 Itential Inc. All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential

package services

import (
	"fmt"
	"net/http"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/itential/ipctl/internal/testlib"
	"github.com/stretchr/testify/assert"
)

var (
	workflowsGetSuccess          = "automation-studio/workflows/get.success.json"
	workflowsGetAllSuccess       = "automation-studio/workflows/getall.success.json"
	workflowsDeleteSuccess       = "automation-studio/workflows/delete.success.json"
	workflowsDeleteNotFound      = "automation-studio/workflows/delete.notfound.json"
	workflowsCreateSuccess       = "automation-studio/automations/create.success.json"
	workflowsExportSuccess       = "workflow_builder/export/export.success.json"
	workflowsExportWithTags      = "workflow_builder/export/export.with-tags.success.json"
	workflowsImportSuccess       = "automation-studio/automations/import.success.json"
	workflowsImportWithTagsSuccess = "automation-studio/automations/import.with-tags.success.json"
	workflowsImportError         = "automation-studio/automations/import.error.json"
)

func setupWorkflowService() *WorkflowService {
	return NewWorkflowService(
		testlib.Setup(),
	)
}

func TestWorkflowGetAll(t *testing.T) {
	svc := setupWorkflowService()
	defer testlib.Teardown()

	for _, ele := range fixtureSuites {
		response := testlib.Fixture(
			filepath.Join(fixtureRoot, ele, workflowsGetAllSuccess),
		)

		testlib.AddGetResponseToMux("/automation-studio/workflows", response, 0)

		res, err := svc.GetAll()

		assert.Nil(t, err)
		assert.Equal(t, 5, len(res))
	}
}

func TestWorkflowGet(t *testing.T) {
	svc := setupWorkflowService()
	defer testlib.Teardown()

	for _, ele := range fixtureSuites {
		response := testlib.Fixture(
			filepath.Join(fixtureRoot, ele, workflowsGetSuccess),
		)

		testlib.AddGetResponseToMux("/automation-studio/workflows", response, 0)

		data, err := fixtureDataToMap(response)
		if err != nil {
			t.FailNow()
		}

		items := data["items"].([]interface{})

		name := items[0].(map[string]interface{})["name"].(string)
		id := items[0].(map[string]interface{})["_id"].(string)

		res, err := svc.Get(name)

		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, reflect.TypeOf((*Workflow)(nil)), reflect.TypeOf(res))
		assert.Equal(t, id, res.Id)
		assert.Equal(t, name, res.Name)
	}

}

func TestWorkflowGetError(t *testing.T) {
	svc := setupWorkflowService()
	defer testlib.Teardown()

	testlib.AddGetErrorToMux("/automation-studio/workflows", "", 0)

	res, err := svc.Get("test")

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestWorkflowCreate(t *testing.T) {
	svc := setupWorkflowService()
	defer testlib.Teardown()

	for _, ele := range fixtureSuites {
		response := testlib.Fixture(
			filepath.Join(fixtureRoot, ele, workflowsCreateSuccess),
		)

		testlib.AddPostResponseToMux("/automation-studio/automations", response, http.StatusOK)

		data, err := fixtureDataToMap(response)
		if err != nil {
			t.FailNow()
		}

		name := data["created"].(map[string]interface{})["name"].(string)
		id := data["created"].(map[string]interface{})["_id"].(string)

		doc := NewWorkflow(name)

		res, err := svc.Create(doc)

		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, reflect.TypeOf((*Workflow)(nil)), reflect.TypeOf(res))
		assert.Equal(t, name, res.Name)
		assert.Equal(t, id, res.Id)
	}
}

func TestWorkflowCreateError(t *testing.T) {
	svc := setupWorkflowService()
	defer testlib.Teardown()

	testlib.AddPostErrorToMux("/automation-studio/workflows", "", 0)

	doc := NewWorkflow("TEST")

	res, err := svc.Create(doc)

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestWorkflowDelete(t *testing.T) {
	svc := setupWorkflowService()
	defer testlib.Teardown()

	for _, ele := range fixtureSuites {
		response := testlib.Fixture(
			filepath.Join(fixtureRoot, ele, workflowsDeleteSuccess),
		)

		data, err := fixtureDataToMap(response)
		if err != nil {
			t.FailNow()
		}

		name := data["value"].(map[string]interface{})["name"].(string)

		testlib.AddDeleteResponseToMux(
			fmt.Sprintf("/workflow_builder/workflows/delete/%s", name), response, http.StatusOK,
		)

		err = svc.Delete(name)

		assert.Nil(t, err)
	}
}

func TestWorkflowDeleteNotFound(t *testing.T) {
	svc := setupWorkflowService()
	defer testlib.Teardown()

	for _, ele := range fixtureSuites {
		response := testlib.Fixture(
			filepath.Join(fixtureRoot, ele, workflowsDeleteNotFound),
		)

		testlib.AddDeleteErrorToMux("/workflow_builder/workflows/delete/test", response, 0)

		err := svc.Delete("test")

		assert.NotNil(t, err)
	}
}

func TestWorkflowExport(t *testing.T) {
	svc := setupWorkflowService()
	defer testlib.Teardown()

	for _, ele := range fixtureSuites {
		response := testlib.Fixture(
			filepath.Join(fixtureRoot, ele, workflowsExportSuccess),
		)

		data, err := fixtureDataToMap(response)
		if err != nil {
			t.FailNow()
		}

		name := data["name"].(string)

		testlib.AddPostErrorToMux("/workflow_builder/export", response, http.StatusOK)

		res, err := svc.Export(name)

		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, reflect.TypeOf((*Workflow)(nil)), reflect.TypeOf(res))
		assert.Equal(t, name, res.Name)
	}
}

func TestWorkflowImport(t *testing.T) {
	svc := setupWorkflowService()
	defer testlib.Teardown()

	for _, ele := range fixtureSuites {
		response := testlib.Fixture(
			filepath.Join(fixtureRoot, ele, workflowsImportSuccess),
		)

		data, err := fixtureDataToMap(response)
		if err != nil {
			t.FailNow()
		}

		name := data["name"].(string)

		testlib.AddPostResponseToMux("/automation-studio/automations/import", response, http.StatusOK)

		doc := NewWorkflow(name)

		res, err := svc.Import(doc)

		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, reflect.TypeOf((*Workflow)(nil)), reflect.TypeOf(res))
		assert.Equal(t, name, res.Name)
	}
}

func TestWorkflowImportError(t *testing.T) {
	svc := setupWorkflowService()
	defer testlib.Teardown()

	for _, ele := range fixtureSuites {
		response := testlib.Fixture(
			filepath.Join(fixtureRoot, ele, workflowsImportError),
		)

		testlib.AddPostResponseToMux("/automation-studio/automations/import", response, http.StatusInternalServerError)

		doc := NewWorkflow("test")

		res, err := svc.Import(doc)

		assert.NotNil(t, err)
		assert.Nil(t, res)
		assert.Equal(t, reflect.TypeOf((*Workflow)(nil)), reflect.TypeOf(res))
	}
}

func TestNewWorkflow(t *testing.T) {
	name := "test-workflow"
	wf := NewWorkflow(name)

	assert.Equal(t, name, wf.Name)
	assert.Equal(t, defaultWorkflowType, wf.Type)
	assert.Equal(t, float64(defaultWorkflowCanvasVersion), wf.CanvasVersion)
	assert.Equal(t, defaultWorkflowFontSize, wf.FontSize)
	assert.NotNil(t, wf.InputSchema)
	assert.NotNil(t, wf.OutputSchema)
	assert.NotNil(t, wf.Tasks)
	assert.NotNil(t, wf.Transitions)
	assert.Contains(t, wf.Tasks, "workflow_start")
	assert.Contains(t, wf.Tasks, "workflow_end")
}

func TestWorkflowGetById(t *testing.T) {
	svc := setupWorkflowService()
	defer testlib.Teardown()

	for _, ele := range fixtureSuites {
		response := testlib.Fixture(
			filepath.Join(fixtureRoot, ele, workflowsGetAllSuccess),
		)

		testlib.AddGetResponseToMux("/automation-studio/workflows", response, 0)

		data, err := fixtureDataToMap(response)
		if err != nil {
			t.FailNow()
		}

		items := data["items"].([]interface{})
		testId := items[0].(map[string]interface{})["_id"].(string)
		testName := items[0].(map[string]interface{})["name"].(string)

		res, err := svc.GetById(testId)

		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, testId, res.Id)
		assert.Equal(t, testName, res.Name)
	}
}

func TestWorkflowGetByIdNotFound(t *testing.T) {
	svc := setupWorkflowService()
	defer testlib.Teardown()

	for _, ele := range fixtureSuites {
		response := testlib.Fixture(
			filepath.Join(fixtureRoot, ele, workflowsGetAllSuccess),
		)

		testlib.AddGetResponseToMux("/automation-studio/workflows", response, 0)

		res, err := svc.GetById("nonexistent-id")

		assert.NotNil(t, err)
		assert.Nil(t, res)
		assert.Equal(t, "workflow not found", err.Error())
	}
}

func TestWorkflowGetByIdError(t *testing.T) {
	svc := setupWorkflowService()
	defer testlib.Teardown()

	testlib.AddGetErrorToMux("/automation-studio/workflows", "", 0)

	res, err := svc.GetById("test-id")

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestWorkflowGetNotFound(t *testing.T) {
	svc := setupWorkflowService()
	defer testlib.Teardown()

	// Mock empty response
	emptyResponse := `{"items": [], "count": 0, "total": 0}`
	testlib.AddGetResponseToMux("/automation-studio/workflows", emptyResponse, 0)

	res, err := svc.Get("nonexistent-workflow")

	assert.NotNil(t, err)
	assert.Nil(t, res)
	assert.Equal(t, "workflow not found", err.Error())
}

func TestWorkflowGetMultipleFound(t *testing.T) {
	svc := setupWorkflowService()
	defer testlib.Teardown()

	// Mock response with multiple workflows
	multipleResponse := `{"items": [{"_id": "1", "name": "test"}, {"_id": "2", "name": "test"}], "count": 2, "total": 2}`
	testlib.AddGetResponseToMux("/automation-studio/workflows", multipleResponse, 0)

	res, err := svc.Get("test")

	assert.NotNil(t, err)
	assert.Nil(t, res)
	assert.Equal(t, "unable to find workflow", err.Error())
}

func TestWorkflowGetAllError(t *testing.T) {
	svc := setupWorkflowService()
	defer testlib.Teardown()

	testlib.AddGetErrorToMux("/automation-studio/workflows", "", 0)

	res, err := svc.GetAll()

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestWorkflowExportError(t *testing.T) {
	svc := setupWorkflowService()
	defer testlib.Teardown()

	testlib.AddPostErrorToMux("/workflow_builder/export", "", 0)

	res, err := svc.Export("test")

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestWorkflowExportById(t *testing.T) {
	svc := setupWorkflowService()
	defer testlib.Teardown()

	for _, ele := range fixtureSuites {
		response := testlib.Fixture(
			filepath.Join(fixtureRoot, ele, workflowsExportSuccess),
		)

		data, err := fixtureDataToMap(response)
		if err != nil {
			t.FailNow()
		}

		name := data["name"].(string)
		testId := "test-workflow-id"

		testlib.AddPostResponseToMux("/workflow_builder/export", response, http.StatusOK)

		res, err := svc.ExportById(testId)

		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, reflect.TypeOf((*Workflow)(nil)), reflect.TypeOf(res))
		assert.Equal(t, name, res.Name)
	}
}

func TestWorkflowExportByIdError(t *testing.T) {
	svc := setupWorkflowService()
	defer testlib.Teardown()

	testlib.AddPostErrorToMux("/workflow_builder/export", "", 0)

	res, err := svc.ExportById("test-id")

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestWorkflowUpdate(t *testing.T) {
	svc := setupWorkflowService()
	defer testlib.Teardown()

	for _, ele := range fixtureSuites {
		response := testlib.Fixture(
			filepath.Join(fixtureRoot, ele, workflowsGetSuccess),
		)

		data, err := fixtureDataToMap(response)
		if err != nil {
			t.FailNow()
		}

		items := data["items"].([]interface{})
		testId := items[0].(map[string]interface{})["_id"].(string)
		testName := items[0].(map[string]interface{})["name"].(string)

		testlib.AddPutResponseToMux(fmt.Sprintf("/automation-studio/automations/%s", testId), string(response), http.StatusOK)

		wf := NewWorkflow(testName)
		wf.Id = testId

		res, err := svc.Update(wf)

		assert.Nil(t, err)
		assert.NotNil(t, res)
	}
}

func TestWorkflowUpdateError(t *testing.T) {
	svc := setupWorkflowService()
	defer testlib.Teardown()

	wf := NewWorkflow("test")
	wf.Id = "test-id"

	testlib.AddPutErrorToMux("/automation-studio/automations/test-id", "", 0)

	res, err := svc.Update(wf)

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestWorkflowClear(t *testing.T) {
	svc := setupWorkflowService()
	defer testlib.Teardown()

	for _, ele := range fixtureSuites {
		getAllResponse := testlib.Fixture(
			filepath.Join(fixtureRoot, ele, workflowsGetAllSuccess),
		)
		deleteResponse := testlib.Fixture(
			filepath.Join(fixtureRoot, ele, workflowsDeleteSuccess),
		)

		testlib.AddGetResponseToMux("/automation-studio/workflows", getAllResponse, 0)

		data, err := fixtureDataToMap(getAllResponse)
		if err != nil {
			t.FailNow()
		}

		items := data["items"].([]interface{})
		for _, item := range items {
			workflowData := item.(map[string]interface{})
			name := workflowData["name"].(string)
			testlib.AddDeleteResponseToMux(
				fmt.Sprintf("/workflow_builder/workflows/delete/%s", name),
				deleteResponse, http.StatusOK,
			)
		}

		err = svc.Clear()

		assert.Nil(t, err)
	}
}

func TestWorkflowClearGetAllError(t *testing.T) {
	svc := setupWorkflowService()
	defer testlib.Teardown()

	testlib.AddGetErrorToMux("/automation-studio/workflows", "", 0)

	err := svc.Clear()

	assert.NotNil(t, err)
}

// TestWorkflowExportPreservesTags verifies that tags returned by the export API
// are deserialised into the Workflow struct and not silently dropped.
func TestWorkflowExportPreservesTags(t *testing.T) {
	svc := setupWorkflowService()
	defer testlib.Teardown()

	for _, ele := range fixtureSuites {
		response := testlib.Fixture(
			filepath.Join(fixtureRoot, ele, workflowsExportWithTags),
		)

		testlib.AddPostErrorToMux("/workflow_builder/export", response, http.StatusOK)

		res, err := svc.Export("test")

		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, []Tag{{Id: "67e6a0f516f3c386c77fa706", Name: "GitLab-Utils", Description: ""}}, res.Tags)
	}
}

// TestWorkflowImportPreservesTags verifies that tags present in an imported workflow
// JSON are serialised and sent to the API, and that the returned workflow carries them.
func TestWorkflowImportPreservesTags(t *testing.T) {
	svc := setupWorkflowService()
	defer testlib.Teardown()

	for _, ele := range fixtureSuites {
		response := testlib.Fixture(
			filepath.Join(fixtureRoot, ele, workflowsImportWithTagsSuccess),
		)

		testlib.AddPostResponseToMux("/automation-studio/automations/import", response, http.StatusOK)

		doc := NewWorkflow("test")
		doc.Tags = []Tag{{Id: "67e6a0f516f3c386c77fa706", Name: "GitLab-Utils", Description: ""}}

		res, err := svc.Import(doc)

		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, []Tag{{Id: "67e6a0f516f3c386c77fa706", Name: "GitLab-Utils", Description: ""}}, res.Tags)
	}
}

func TestWorkflowClearDeleteError(t *testing.T) {
	svc := setupWorkflowService()
	defer testlib.Teardown()

	for _, ele := range fixtureSuites {
		getAllResponse := testlib.Fixture(
			filepath.Join(fixtureRoot, ele, workflowsGetAllSuccess),
		)

		testlib.AddGetResponseToMux("/automation-studio/workflows", getAllResponse, 0)

		data, err := fixtureDataToMap(getAllResponse)
		if err != nil {
			t.FailNow()
		}

		items := data["items"].([]interface{})
		if len(items) > 0 {
			workflowData := items[0].(map[string]interface{})
			name := workflowData["name"].(string)
			testlib.AddDeleteErrorToMux(
				fmt.Sprintf("/workflow_builder/workflows/delete/%s", name),
				"", 0,
			)

			err = svc.Clear()

			assert.NotNil(t, err)
		}
	}
}
