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
	transformationsGetAllSuccess         = "transformations/getall.success.json"
	transformationsGetSuccess            = "transformations/get.success.json"
	transformationsGetNotFound           = "transformations/get.notfound.json"
	transformationsCreateSuccess         = "transformations/create.success.json"
	transformationsCreateError           = "transformations/create.error.json"
	transformationsCreateDuplicate       = "transformations/create.duplicate.json"
	transformationsDeleteSuccess         = "transformations/delete.success.json"
	transformationsDeleteNotFound        = "transformations/delete.notfound.json"
	transformationsImportSuccess         = "transformations/import.success.json"
	transformationsImportWithTagsSuccess = "transformations/export.with-tags.json"
)

func setupTransformationService() *TransformationService {
	return NewTransformationService(
		testlib.Setup(),
	)
}

func TestTransformationGetAll(t *testing.T) {
	svc := setupTransformationService()
	defer testlib.Teardown()

	for _, ele := range fixtureSuites {
		response := testlib.Fixture(
			filepath.Join(fixtureRoot, ele, transformationsGetAllSuccess),
		)

		testlib.AddGetResponseToMux("/transformations", response, 0)

		res, err := svc.GetAll()

		assert.Nil(t, err)
		assert.Equal(t, 1, len(res))
	}
}

func TestTransformationGet(t *testing.T) {
	svc := setupTransformationService()
	defer testlib.Teardown()

	for _, ele := range fixtureSuites {
		response := testlib.Fixture(
			filepath.Join(fixtureRoot, ele, transformationsGetSuccess),
		)

		data, err := fixtureDataToMap(response)
		if err != nil {
			t.FailNow()
		}

		id := data["_id"].(string)

		testlib.AddGetResponseToMux(fmt.Sprintf("/transformations/%s", id), response, 0)

		res, err := svc.Get(id)

		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, reflect.TypeOf((*Transformation)(nil)), reflect.TypeOf(res))
		assert.Equal(t, id, res.Id)
	}
}

func TestTransformationGetByName(t *testing.T) {
	svc := setupTransformationService()
	defer testlib.Teardown()

	for _, ele := range fixtureSuites {
		response := testlib.Fixture(
			filepath.Join(fixtureRoot, ele, transformationsGetAllSuccess),
		)

		data, err := fixtureDataToMap(response)
		if err != nil {
			t.FailNow()
		}

		results := data["results"].([]interface{})

		name := results[0].(map[string]interface{})["name"].(string)
		id := results[0].(map[string]interface{})["_id"].(string)

		testlib.AddGetResponseToMux("/transformations", response, 0)

		res, err := svc.GetByName(name)

		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, reflect.TypeOf((*Transformation)(nil)), reflect.TypeOf(res))
		assert.Equal(t, id, res.Id)
		assert.Equal(t, name, res.Name)
	}
}

func TestTransformationGetByNameNotFound(t *testing.T) {
	svc := setupTransformationService()
	defer testlib.Teardown()

	for _, ele := range fixtureSuites {
		response := testlib.Fixture(
			filepath.Join(fixtureRoot, ele, transformationsGetNotFound),
		)

		testlib.AddGetResponseToMux("/transformations", response, 0)

		res, err := svc.GetByName("abcdefghijklmnopqrstuvwxyz")

		assert.NotNil(t, err)
		assert.Nil(t, res)
		assert.Equal(t, err.Error(), "transformation not found")
	}
}

func TestTransformationCreate(t *testing.T) {
	svc := setupTransformationService()
	defer testlib.Teardown()

	for _, ele := range fixtureSuites {
		response := testlib.Fixture(
			filepath.Join(fixtureRoot, ele, transformationsCreateSuccess),
		)

		testlib.AddPostResponseToMux("/transformations", response, http.StatusOK)

		data, err := fixtureDataToMap(response)
		if err != nil {
			t.FailNow()
		}

		name := data["name"].(string)
		id := data["_id"].(string)

		res, err := svc.Create(NewTransformation(name, ""))

		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, reflect.TypeOf((*Transformation)(nil)), reflect.TypeOf(res))
		assert.Equal(t, id, res.Id)
	}
}

func TestTransformationCreateDuplicate(t *testing.T) {
	svc := setupTransformationService()
	defer testlib.Teardown()

	for _, ele := range fixtureSuites {
		response := testlib.Fixture(
			filepath.Join(fixtureRoot, ele, transformationsCreateDuplicate),
		)

		testlib.AddPostErrorToMux("/transformations", response, 0)

		res, err := svc.Create(
			NewTransformation("test", ""),
		)

		assert.NotNil(t, err)
		assert.Nil(t, res)
	}
}

func TestTransformationDelete(t *testing.T) {
	svc := setupTransformationService()
	defer testlib.Teardown()

	for _, ele := range fixtureSuites {
		response := testlib.Fixture(
			filepath.Join(fixtureRoot, ele, transformationsDeleteSuccess),
		)

		data, err := fixtureDataToMap(response)
		if err != nil {
			t.FailNow()
		}

		id := data["results"].([]interface{})[0].(map[string]interface{})["_id"].(string)

		testlib.AddDeleteResponseToMux(
			fmt.Sprintf("/transformations/%s", id),
			response,
			http.StatusNoContent,
		)

		assert.Nil(t, svc.Delete(id))
	}
}

func TestTransformationDeleteNotFound(t *testing.T) {
	svc := setupTransformationService()
	defer testlib.Teardown()

	for _, ele := range fixtureSuites {
		response := testlib.Fixture(
			filepath.Join(fixtureRoot, ele, transformationsDeleteNotFound),
		)
		testlib.AddDeleteErrorToMux("/transformations/test", response, 0)

		assert.NotNil(t, svc.Delete("test"))
	}
}

func TestTransformationImport(t *testing.T) {
	svc := setupTransformationService()
	defer testlib.Teardown()

	for _, ele := range fixtureSuites {
		response := testlib.Fixture(
			filepath.Join(fixtureRoot, ele, transformationsImportSuccess),
		)

		testlib.AddPostResponseToMux("/transformations/import", response, http.StatusOK)

		data, err := fixtureDataToMap(response)
		if err != nil {
			t.FailNow()
		}

		name := data["name"].(string)
		id := data["_id"].(string)

		res, err := svc.Import(NewTransformation(name, ""))

		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, reflect.TypeOf((*Transformation)(nil)), reflect.TypeOf(res))
		assert.Equal(t, id, res.Id)
		assert.Equal(t, name, res.Name)
	}
}

// TestTransformationImportPreservesTags verifies that tag objects returned by the
// import API are deserialised into the Transformation struct and not silently
// dropped. Prior to the fix, Tags []string caused an unmarshal error when the API
// returned tag objects.
func TestTransformationImportPreservesTags(t *testing.T) {
	svc := setupTransformationService()
	defer testlib.Teardown()

	for _, ele := range fixtureSuites {
		response := testlib.Fixture(
			filepath.Join(fixtureRoot, ele, transformationsImportWithTagsSuccess),
		)

		testlib.AddPostResponseToMux("/transformations/import", response, http.StatusOK)

		res, err := svc.Import(NewTransformation("test", ""))

		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, []Tag{{Id: "67e6a0f516f3c386c77fa706", Name: "GitLab-Utils", Description: ""}}, res.Tags)
	}
}

// TestNewTransformationHasEmptyTags verifies that a newly created Transformation
// initialises Tags as an empty slice, matching the platform API behaviour.
func TestNewTransformationHasEmptyTags(t *testing.T) {
	tr := NewTransformation("test", "")
	assert.Equal(t, []Tag{}, tr.Tags)
}
