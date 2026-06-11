// Copyright 2024 Itential Inc. All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential

package runners

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/itential/ipctl/internal/testlib"
	"github.com/itential/ipctl/pkg/services"
	"github.com/stretchr/testify/assert"
)

// sampleExpandableProject builds a project with two components carrying inline
// documents, suitable for exercising the expand/import round-trip.
func sampleExpandableProject() *services.Project {
	return &services.Project{
		Id:                "proj-id",
		Name:              "RoundTripProject",
		ComponentIidIndex: 42,
		Folders: []services.ProjectFolder{
			{Iid: 1, Name: "Workflows", NodeType: "folder", Children: []services.ProjectFolder{}},
		},
		Members: []services.ProjectMember{{Reference: "abc", Role: "owner", Type: "account"}},
		Components: []services.ProjectComponent{
			{
				Iid:       1,
				Type:      "workflow",
				Folder:    "/Workflows",
				Reference: "ref-wf",
				Document:  map[string]interface{}{"name": "Sample Workflow", "tasks": map[string]interface{}{}},
			},
			{
				Iid:       2,
				Type:      "transformation",
				Folder:    "/",
				Reference: "ref-tf",
				Document:  map[string]interface{}{"name": "Sample Transformation"},
			},
		},
	}
}

// TestExpandProjectWritesComponentFiles verifies that the expanded export
// writes each component document to its own file, references it by filename in
// the main project file, and strips server-managed fields (members,
// accessControl, componentIidIndex).
func TestExpandProjectWritesComponentFiles(t *testing.T) {
	dir := t.TempDir()
	project := sampleExpandableProject()

	err := expandProject(Request{}, project, dir)
	assert.Nil(t, err)

	// Component documents are written to their own files.
	assert.FileExists(t, filepath.Join(dir, "Workflows", "Sample Workflow.workflow.json"))
	assert.FileExists(t, filepath.Join(dir, "Sample Transformation.transformation.json"))

	// The main project file references documents by filename, not inline.
	b, err := os.ReadFile(filepath.Join(dir, "RoundTripProject.project.json"))
	assert.Nil(t, err)

	var main map[string]interface{}
	assert.Nil(t, json.Unmarshal(b, &main))

	_, hasMembers := main["members"]
	_, hasAccessControl := main["accessControl"]
	_, hasComponentIidIndex := main["componentIidIndex"]
	assert.False(t, hasMembers)
	assert.False(t, hasAccessControl)
	assert.False(t, hasComponentIidIndex)

	components := main["components"].([]interface{})
	assert.Equal(t, 2, len(components))
	for _, c := range components {
		comp := c.(map[string]interface{})
		_, hasDocument := comp["document"]
		assert.False(t, hasDocument, "expanded component should not embed its document")
		assert.NotEmpty(t, comp["filename"])
	}
}

// TestImportProjectMissingComponentFile verifies that importing an expanded
// project whose component files are absent returns a clear error instead of
// silently importing empty components or terminating the process.
func TestImportProjectMissingComponentFile(t *testing.T) {
	runner := NewProjectRunner(testlib.Setup(), testlib.DefaultConfig())
	defer testlib.Teardown()

	dir := t.TempDir()
	mainPath := filepath.Join(dir, "Broken.project.json")

	main := map[string]interface{}{
		"name": "Broken",
		"components": []map[string]interface{}{
			{
				"iid":       1,
				"type":      "workflow",
				"folder":    "/",
				"reference": "ref-missing",
				"filename":  "does-not-exist.workflow.json",
			},
		},
		"folders": []interface{}{},
	}
	b, _ := json.MarshalIndent(main, "", "  ")
	assert.Nil(t, os.WriteFile(mainPath, b, 0o644))

	var project services.Project
	assert.Nil(t, importLoadFromDisk(mainPath, &project))

	res, err := runner.importProject(project, mainPath, false)

	assert.Nil(t, res)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "ref-missing")
	assert.Contains(t, err.Error(), "component files")
}

// TestExpandImportRoundTrip verifies that a project exported in expanded form
// can be imported back: the component documents are reconstructed from their
// files and the import succeeds.
func TestExpandImportRoundTrip(t *testing.T) {
	runner := NewProjectRunner(testlib.Setup(), testlib.DefaultConfig())
	defer testlib.Teardown()

	dir := t.TempDir()
	project := sampleExpandableProject()

	// Export in expanded form (writes main file plus component files).
	assert.Nil(t, expandProject(Request{}, project, dir))

	// GetByName finds no existing project, so import proceeds without a delete.
	testlib.AddGetResponseToMux("/automation-studio/projects", `{"data":[],"metadata":{"total":0}}`, 0)

	// The import endpoint accepts the reconstructed project.
	testlib.AddPostResponseToMux(
		"/automation-studio/projects/import",
		`{"message":"imported","data":{"_id":"new-id","name":"RoundTripProject"},"metadata":{}}`,
		http.StatusOK,
	)

	mainPath := filepath.Join(dir, "RoundTripProject.project.json")

	var loaded services.Project
	assert.Nil(t, importLoadFromDisk(mainPath, &loaded))

	res, err := runner.importProject(loaded, mainPath, false)

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "new-id", res.Id)
}
