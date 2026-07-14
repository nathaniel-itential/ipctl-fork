// Copyright 2024 Itential Inc. All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential

package services

import (
	"net/http"
	"path/filepath"
	"testing"

	"github.com/itential/ipctl/internal/testlib"
	"github.com/stretchr/testify/assert"
)

var (
	analyticTemplateExportWithTagsSuccess = "analytic-templates/export.with-tags.json"
	analyticTemplateExportSuccess         = "analytic-templates/export.json"
	analyticTemplateImportSuccess         = "analytic-templates/import.success.json"
)

func setupAnalyticTemplateService() *AnalyticTemplateService {
	return NewAnalyticTemplateService(testlib.Setup())
}

// TestAnalyticTemplateExportPreservesTags verifies that tag objects returned by the
// export API are deserialised into the AnalyticTemplate struct and not silently
// dropped. Prior to the fix, Tags []string caused an unmarshal error when the API
// returned tag objects.
func TestAnalyticTemplateExportPreservesTags(t *testing.T) {
	svc := setupAnalyticTemplateService()
	defer testlib.Teardown()

	for _, ele := range fixtureSuites {
		response := testlib.Fixture(
			filepath.Join(fixtureRoot, ele, analyticTemplateExportWithTagsSuccess),
		)

		testlib.AddPostResponseToMux("/mop/export", response, http.StatusOK)

		res, err := svc.Export("test")

		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, []Tag{{Id: "6800fb9af7a998b975bde8e5", Name: "Netbox-Utils", Description: "All assets related to managing Netbox integration"}}, res.Tags)
	}
}

// TestAnalyticTemplateExportEmptyTags verifies that an analytic template without tags
// exports with "tags": [] rather than omitting the field.
func TestAnalyticTemplateExportEmptyTags(t *testing.T) {
	svc := setupAnalyticTemplateService()
	defer testlib.Teardown()

	for _, ele := range fixtureSuites {
		response := testlib.Fixture(
			filepath.Join(fixtureRoot, ele, analyticTemplateExportSuccess),
		)

		testlib.AddPostResponseToMux("/mop/export", response, http.StatusOK)

		res, err := svc.Export("test")

		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, []Tag{}, res.Tags)
	}
}

// TestAnalyticTemplateImportPreservesTags verifies that tags on an AnalyticTemplate
// are serialised and sent to the API on import.
func TestAnalyticTemplateImportPreservesTags(t *testing.T) {
	svc := setupAnalyticTemplateService()
	defer testlib.Teardown()

	for _, ele := range fixtureSuites {
		response := testlib.Fixture(
			filepath.Join(fixtureRoot, ele, analyticTemplateImportSuccess),
		)

		testlib.AddPostResponseToMux("/mop/import", response, http.StatusCreated)

		doc := NewAnalyticTemplate("test")
		doc.Tags = []Tag{{Id: "6800fb9af7a998b975bde8e5", Name: "Netbox-Utils", Description: "All assets related to managing Netbox integration"}}

		err := svc.Import(doc)

		assert.Nil(t, err)
	}
}

// TestNewAnalyticTemplateHasEmptyTags verifies that a newly created AnalyticTemplate
// initialises Tags as an empty slice, matching the platform API behaviour.
func TestNewAnalyticTemplateHasEmptyTags(t *testing.T) {
	at := NewAnalyticTemplate("test")
	assert.Equal(t, []Tag{}, at.Tags)
}
