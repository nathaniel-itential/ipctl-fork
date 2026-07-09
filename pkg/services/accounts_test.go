// Copyright 2024 Itential Inc. All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential

package services

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/itential/ipctl/internal/testlib"
	"github.com/stretchr/testify/assert"
)

var (
	accountsGetSuccess    = "authorization/accounts/get.success.json"
	accountsGetAllSuccess = "authorization/accounts/getall.success.json"
)

func setupAccountService() *AccountService {
	return NewAccountService(
		testlib.Setup(),
	)
}

func TestAccountGetAll(t *testing.T) {
	svc := setupAccountService()
	defer testlib.Teardown()

	for _, ele := range fixtureSuites {
		response := testlib.Fixture(
			filepath.Join(fixtureRoot, ele, accountsGetAllSuccess),
		)
		// The new pagination implementation uses GetRequest, so we need to handle query parameters
		testlib.AddGetResponseToMux("/authorization/accounts", response, 0)

		res, err := svc.GetAll()

		assert.Nil(t, err)
		assert.Equal(t, 2, len(res))
		if len(res) > 0 {
			assert.NotEmpty(t, res[0].Id)
			assert.NotEmpty(t, res[0].Username)
		}
		if len(res) > 1 {
			assert.NotEmpty(t, res[1].Id)
			assert.NotEmpty(t, res[1].Username)
		}
	}
}

func TestAccountGetAllError(t *testing.T) {
	svc := setupAccountService()
	defer testlib.Teardown()

	testlib.AddGetErrorToMux("/authorization/accounts", "", 0)

	res, err := svc.GetAll()

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestAccountGet(t *testing.T) {
	svc := setupAccountService()
	defer testlib.Teardown()

	for _, ele := range fixtureSuites {
		response := testlib.Fixture(
			filepath.Join(fixtureRoot, ele, accountsGetSuccess),
		)

		// Use specific path that matches the Get method implementation
		testlib.AddGetResponseToMux("/authorization/accounts/ID", response, 0)

		res, err := svc.Get("ID")

		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, reflect.TypeOf((*Account)(nil)), reflect.TypeOf(res))
		assert.NotEmpty(t, res.Id)
	}
}

func TestAccountGetError(t *testing.T) {
	svc := setupAccountService()
	defer testlib.Teardown()

	testlib.AddGetErrorToMux("/authorization/accounts/TEST", "", 0)

	res, err := svc.Get("TEST")

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestAccountGetByName(t *testing.T) {
	svc := setupAccountService()
	defer testlib.Teardown()

	for _, ele := range fixtureSuites {
		response := testlib.Fixture(
			filepath.Join(fixtureRoot, ele, accountsGetAllSuccess),
		)
		testlib.AddGetResponseToMux("/authorization/accounts", response, 0)

		res, err := svc.GetByName("admin@pronghorn")

		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, reflect.TypeOf((*Account)(nil)), reflect.TypeOf(res))
		assert.NotEmpty(t, res.Id)
		assert.Equal(t, "admin@pronghorn", res.Username)
		assert.Equal(t, "admin", res.FirstName)
		assert.Equal(t, "local_aaa", res.Provenance)
		assert.False(t, res.Inactive)
		assert.True(t, res.LoggedIn)

	}
}

func TestAccountGetByNameError(t *testing.T) {
	svc := setupAccountService()
	defer testlib.Teardown()

	testlib.AddGetErrorToMux("/authorization/accounts", "", 0)

	res, err := svc.GetByName("TEST")

	assert.NotNil(t, err)
	assert.Nil(t, res)
	assert.Equal(t, reflect.TypeOf((*Account)(nil)), reflect.TypeOf(res))
}

func TestAccountGetByNamePrefersActive(t *testing.T) {
	svc := setupAccountService()
	defer testlib.Teardown()

	mockResponse := `{
		"results": [
			{
				"_id": "inactive-id-1",
				"email": "joksan.flores@itential.com",
				"firstname": "Joksan",
				"inactive": true,
				"loggedIn": false,
				"provenance": "Okta SAML",
				"username": "joksan.flores@itential.com"
			},
			{
				"_id": "active-id",
				"email": "joksan.flores@itential.com",
				"firstname": "Joksan",
				"inactive": false,
				"loggedIn": true,
				"provenance": "CloudAAA",
				"username": "joksan.flores@itential.com"
			}
		],
		"total": 2
	}`

	testlib.AddGetResponseToMux("/authorization/accounts", mockResponse, 0)

	res, err := svc.GetByName("joksan.flores@itential.com")

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "active-id", res.Id)
	assert.False(t, res.Inactive)
}

func TestAccountGetByNameNotFound(t *testing.T) {
	svc := setupAccountService()
	defer testlib.Teardown()

	for _, ele := range fixtureSuites {
		response := testlib.Fixture(
			filepath.Join(fixtureRoot, ele, accountsGetAllSuccess),
		)
		testlib.AddGetResponseToMux("/authorization/accounts", response, 0)

		res, err := svc.GetByName("nonexistent@user")

		assert.NotNil(t, err)
		assert.Nil(t, res)
		assert.Equal(t, "account not found", err.Error())
	}
}

func TestAccountActivate(t *testing.T) {
	svc := setupAccountService()
	defer testlib.Teardown()

	testlib.AddPatchResponseToMux("/authorization/accounts/test-id", "", 0)

	err := svc.Activate("test-id")

	assert.Nil(t, err)
}

func TestAccountActivateError(t *testing.T) {
	svc := setupAccountService()
	defer testlib.Teardown()

	testlib.AddPatchErrorToMux("/authorization/accounts/test-id", "", 0)

	err := svc.Activate("test-id")

	assert.NotNil(t, err)
}

func TestAccountDeactivate(t *testing.T) {
	svc := setupAccountService()
	defer testlib.Teardown()

	testlib.AddPatchResponseToMux("/authorization/accounts/test-id", "", 0)

	err := svc.Deactivate("test-id")

	assert.Nil(t, err)
}

func TestAccountDeactivateError(t *testing.T) {
	svc := setupAccountService()
	defer testlib.Teardown()

	testlib.AddPatchErrorToMux("/authorization/accounts/test-id", "", 0)

	err := svc.Deactivate("test-id")

	assert.NotNil(t, err)
}

func TestNewAccountService(t *testing.T) {
	client := testlib.Setup()
	defer testlib.Teardown()

	svc := NewAccountService(client)

	assert.NotNil(t, svc)
	assert.NotNil(t, svc.client)
	assert.Equal(t, reflect.TypeOf((*AccountService)(nil)), reflect.TypeOf(svc))
}
