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
	commandTemplateExportSuccess         = "mop/command-template.export.json"
	commandTemplateExportWithTagsSuccess = "mop/command-template.export.with-tags.json"
	commandTemplateImportSuccess         = "mop/command-template.import.json"
)

func setupCommandTemplateService() *CommandTemplateService {
	return NewCommandTemplateService(testlib.Setup())
}

// TestCommandTemplateExportPreservesTags verifies that tag objects returned by the
// export API are deserialised into the CommandTemplate struct and not silently
// dropped. Prior to the fix, Tags []string caused an unmarshal error when the API
// returned tag objects.
func TestCommandTemplateExportPreservesTags(t *testing.T) {
	svc := setupCommandTemplateService()
	defer testlib.Teardown()

	for _, ele := range fixtureSuites {
		response := testlib.Fixture(
			filepath.Join(fixtureRoot, ele, commandTemplateExportWithTagsSuccess),
		)

		testlib.AddPostResponseToMux("/mop/export", response, http.StatusOK)

		res, err := svc.Export("test")

		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, []Tag{{Id: "6800fb9af7a998b975bde8e5", Name: "Netbox-Utils", Description: "All assets related to managing Netbox integration"}}, res.Tags)
	}
}

// TestCommandTemplateExportEmptyTags verifies that a command template without tags
// exports with "tags": [] rather than omitting the field.
func TestCommandTemplateExportEmptyTags(t *testing.T) {
	svc := setupCommandTemplateService()
	defer testlib.Teardown()

	for _, ele := range fixtureSuites {
		response := testlib.Fixture(
			filepath.Join(fixtureRoot, ele, commandTemplateExportSuccess),
		)

		testlib.AddPostResponseToMux("/mop/export", response, http.StatusOK)

		res, err := svc.Export("test")

		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, []Tag{}, res.Tags)
	}
}

// TestCommandTemplateImportPreservesTags verifies that tags on a CommandTemplate
// are serialised and sent to the API on import.
func TestCommandTemplateImportPreservesTags(t *testing.T) {
	svc := setupCommandTemplateService()
	defer testlib.Teardown()

	for _, ele := range fixtureSuites {
		response := testlib.Fixture(
			filepath.Join(fixtureRoot, ele, commandTemplateImportSuccess),
		)

		testlib.AddPostResponseToMux("/mop/import", response, http.StatusOK)

		doc := NewCommandTemplate("test")
		doc.Tags = []Tag{{Id: "6800fb9af7a998b975bde8e5", Name: "Netbox-Utils", Description: "All assets related to managing Netbox integration"}}

		err := svc.Import(doc)

		assert.Nil(t, err)
	}
}

// TestNewCommandTemplateHasEmptyTags verifies that a newly created CommandTemplate
// initialises Tags as an empty slice, matching the platform API behaviour.
func TestNewCommandTemplateHasEmptyTags(t *testing.T) {
	ct := NewCommandTemplate("test")
	assert.Equal(t, []Tag{}, ct.Tags)
}
