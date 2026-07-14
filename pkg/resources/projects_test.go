// Copyright 2024 Itential Inc. All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential

package resources

import (
	"testing"

	"github.com/itential/ipctl/pkg/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockProjectService is a minimal ProjectServicer that captures the data
// passed to Import and UpdateProject for inspection in tests.
type mockProjectService struct {
	services.ProjectServicer
	importedData    map[string]interface{}
	updatedData     map[string]interface{}
	existingProject *services.Project
	returnErr       error
}

func (m *mockProjectService) Import(data map[string]interface{}) (*services.Project, error) {
	m.importedData = data
	if m.returnErr != nil {
		return nil, m.returnErr
	}
	return &services.Project{Id: "test-id", Name: "test"}, nil
}

func (m *mockProjectService) Get(id string) (*services.Project, error) {
	if m.existingProject != nil {
		return m.existingProject, nil
	}
	return &services.Project{Id: id}, nil
}

func (m *mockProjectService) UpdateProject(projectId string, data map[string]interface{}) error {
	m.updatedData = data
	return m.returnErr
}

func (m *mockProjectService) GetAll() ([]services.Project, error) { return nil, nil }
func (m *mockProjectService) GetByName(name string) (*services.Project, error) {
	return nil, nil
}

func newProjectResourceWithMock(svc *mockProjectService) ProjectResourcer {
	return &ProjectResource{
		BaseResource: NewBaseResource(),
		service:      svc,
	}
}

func sampleProject() services.Project {
	return services.Project{
		Id:   "proj-123",
		Name: "My Project",
	}
}

func TestProjectImport_NoConflictModeOmitsField(t *testing.T) {
	mock := &mockProjectService{}
	r := newProjectResourceWithMock(mock)

	_, err := r.Import(sampleProject(), services.ProjectImportConfig{})

	require.NoError(t, err)
	_, present := mock.importedData["conflictMode"]
	assert.False(t, present, "conflictMode should be omitted when not specified")
}

func TestProjectImport_ConflictModeOverwrite(t *testing.T) {
	mock := &mockProjectService{}
	r := newProjectResourceWithMock(mock)

	_, err := r.Import(sampleProject(), services.ProjectImportConfig{
		ConflictMode: services.ConflictModeOverwrite,
	})

	require.NoError(t, err)
	assert.Equal(t, services.ConflictModeOverwrite, mock.importedData["conflictMode"])
}

func TestProjectImport_ConflictModeInsertNew(t *testing.T) {
	mock := &mockProjectService{}
	r := newProjectResourceWithMock(mock)

	_, err := r.Import(sampleProject(), services.ProjectImportConfig{
		ConflictMode: services.ConflictModeInsertNew,
	})

	require.NoError(t, err)
	assert.Equal(t, services.ConflictModeInsertNew, mock.importedData["conflictMode"])
}

func TestProjectImport_SkipReferenceValidation(t *testing.T) {
	mock := &mockProjectService{}
	r := newProjectResourceWithMock(mock)

	_, err := r.Import(sampleProject(), services.ProjectImportConfig{
		SkipReferenceValidation: true,
	})

	require.NoError(t, err)
	assert.Equal(t, true, mock.importedData["skipReferenceValidation"])
}

func TestProjectImport_SkipReferenceValidationDefaultsFalse(t *testing.T) {
	mock := &mockProjectService{}
	r := newProjectResourceWithMock(mock)

	_, err := r.Import(sampleProject(), services.ProjectImportConfig{})

	require.NoError(t, err)
	assert.Equal(t, false, mock.importedData["skipReferenceValidation"])
}

func TestProjectImport_AssignNewReferences(t *testing.T) {
	mock := &mockProjectService{}
	r := newProjectResourceWithMock(mock)

	_, err := r.Import(sampleProject(), services.ProjectImportConfig{
		AssignNewReferences: true,
	})

	require.NoError(t, err)
	assert.Equal(t, true, mock.importedData["assignNewReferences"])
}

func TestProjectImport_AssignNewReferencesDefaultsFalse(t *testing.T) {
	mock := &mockProjectService{}
	r := newProjectResourceWithMock(mock)

	_, err := r.Import(sampleProject(), services.ProjectImportConfig{})

	require.NoError(t, err)
	assert.Equal(t, false, mock.importedData["assignNewReferences"])
}

func TestProjectImport_AllOptionsSet(t *testing.T) {
	mock := &mockProjectService{}
	r := newProjectResourceWithMock(mock)

	_, err := r.Import(sampleProject(), services.ProjectImportConfig{
		ConflictMode:            services.ConflictModeOverwrite,
		SkipReferenceValidation: true,
		AssignNewReferences:     true,
	})

	require.NoError(t, err)
	assert.Equal(t, services.ConflictModeOverwrite, mock.importedData["conflictMode"])
	assert.Equal(t, true, mock.importedData["skipReferenceValidation"])
	assert.Equal(t, true, mock.importedData["assignNewReferences"])
}

func TestProjectImport_ProjectFieldPresent(t *testing.T) {
	mock := &mockProjectService{}
	r := newProjectResourceWithMock(mock)

	_, err := r.Import(sampleProject(), services.ProjectImportConfig{})

	require.NoError(t, err)
	assert.NotNil(t, mock.importedData["project"])
}

func TestAddMembers_DeduplicatesExistingMembers(t *testing.T) {
	mock := &mockProjectService{
		existingProject: &services.Project{
			Id: "proj-123",
			Members: []services.ProjectMember{
				{Reference: "ref-alice", Role: "owner", Type: "account"},
				{Reference: "ref-bob", Role: "editor", Type: "account"},
			},
		},
	}
	r := newProjectResourceWithMock(mock)

	// alice is already on the project — should not be duplicated
	err := r.AddMembers("proj-123", []services.ProjectMember{
		{Reference: "ref-alice", Role: "owner", Type: "account"},
	})

	require.NoError(t, err)
	members := mock.updatedData["members"].([]services.ProjectMember)
	assert.Len(t, members, 2, "should have alice (new) + bob (existing), not three entries")

	refs := make([]string, len(members))
	for i, m := range members {
		refs[i] = m.Reference
	}
	assert.Contains(t, refs, "ref-alice")
	assert.Contains(t, refs, "ref-bob")
}

func TestAddMembers_AddsNewMembersToExisting(t *testing.T) {
	mock := &mockProjectService{
		existingProject: &services.Project{
			Id: "proj-123",
			Members: []services.ProjectMember{
				{Reference: "ref-alice", Role: "owner", Type: "account"},
			},
		},
	}
	r := newProjectResourceWithMock(mock)

	err := r.AddMembers("proj-123", []services.ProjectMember{
		{Reference: "ref-carol", Role: "editor", Type: "account"},
	})

	require.NoError(t, err)
	members := mock.updatedData["members"].([]services.ProjectMember)
	assert.Len(t, members, 2)
}

func TestAddMembers_NoExistingMembers(t *testing.T) {
	mock := &mockProjectService{
		existingProject: &services.Project{Id: "proj-123", Members: nil},
	}
	r := newProjectResourceWithMock(mock)

	err := r.AddMembers("proj-123", []services.ProjectMember{
		{Reference: "ref-alice", Role: "owner", Type: "account"},
	})

	require.NoError(t, err)
	members := mock.updatedData["members"].([]services.ProjectMember)
	assert.Len(t, members, 1)
}
