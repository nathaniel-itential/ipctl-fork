// Copyright 2024 Itential Inc. All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential

package resources

import (
	"encoding/json"

	"github.com/itential/ipctl/internal/logging"
	"github.com/itential/ipctl/pkg/services"
)

// ProjectResource provides business logic for project operations.
type ProjectResource struct {
	BaseResource
	service services.ProjectServicer
}

// NewProjectResource creates a new ProjectResource with the given service.
func NewProjectResource(svc services.ProjectServicer) ProjectResourcer {
	return &ProjectResource{
		BaseResource: NewBaseResource(),
		service:      svc,
	}
}

// GetAll retrieves all projects from the API.
// This is a pass-through to the service layer for pure API access.
func (r *ProjectResource) GetAll() ([]services.Project, error) {
	return r.service.GetAll()
}

// Get retrieves a specific project by ID from the API.
// This is a pass-through to the service layer for pure API access.
func (r *ProjectResource) Get(id string) (*services.Project, error) {
	return r.service.Get(id)
}

// Create creates a new project with the specified name.
// This is a pass-through to the service layer for pure API access.
func (r *ProjectResource) Create(name string) (*services.Project, error) {
	return r.service.Create(name)
}

// Delete removes a project by its identifier.
// This is a pass-through to the service layer for pure API access.
func (r *ProjectResource) Delete(id string) error {
	return r.service.Delete(id)
}

// Export retrieves a project in export format by its identifier.
// This is a pass-through to the service layer for pure API access.
func (r *ProjectResource) Export(id string) (*services.Project, error) {
	return r.service.Export(id)
}

// ImportTransformed imports a project using pre-transformed data.
// This is a pass-through to the service layer for pure API access.
func (r *ProjectResource) ImportTransformed(data map[string]interface{}) (*services.Project, error) {
	return r.service.Import(data)
}

// UpdateMembers updates the members of a project via PATCH request.
// This is a pass-through to the service layer for pure API access.
func (r *ProjectResource) UpdateMembers(projectId string, members []services.ProjectMember) error {
	data := map[string]interface{}{
		"members": members,
	}
	return r.service.UpdateProject(projectId, data)
}

// GetByName retrieves a project by name using client-side filtering.
// It fetches all projects and searches for a matching name.
func (r *ProjectResource) GetByName(name string) (*services.Project, error) {
	logging.Trace()

	projects, err := r.service.GetAll()
	if err != nil {
		return nil, err
	}

	return FindByName(projects, "project", name, func(p services.Project) string {
		return p.Name
	})
}

// transformImport recursively iterates over folders in a project schema and
// removes keys to prepare the body for server acceptance during import operations.
func (r *ProjectResource) transformImport(in map[string]interface{}) {
	if in["nodeType"].(string) == "folder" {
		delete(in, "iid")
	}

	if in["nodeType"].(string) == "component" {
		delete(in, "name")
	}

	if in["children"] != nil {
		for _, ele := range in["children"].([]interface{}) {
			r.transformImport(ele.(map[string]interface{}))
		}
	} else if in["children"] == nil {
		delete(in, "children")
	}
}

// Import imports a project with data transformation for server compatibility.
// This method handles the business logic of transforming folder structures
// before sending to the API.
func (r *ProjectResource) Import(in services.Project, cfg services.ProjectImportConfig) (*services.Project, error) {
	logging.Trace()

	body := map[string]interface{}{
		"skipReferenceValidation": cfg.SkipReferenceValidation,
		"assignNewReferences":     cfg.AssignNewReferences,
		"project":                 in.Import(),
	}

	if cfg.ConflictMode != "" {
		body["conflictMode"] = cfg.ConflictMode
	}

	b, _ := json.Marshal(body)

	var data map[string]interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, err
	}

	project := data["project"].(map[string]interface{})

	if foldersRaw, exists := project["folders"]; exists && foldersRaw != nil {
		if folders, ok := foldersRaw.([]interface{}); ok && folders != nil {
			for _, ele := range folders {
				r.transformImport(ele.(map[string]interface{}))
			}
		}
	}

	return r.service.Import(data)
}

// AddMembers adds new members to an existing project.
// This method implements the business logic of fetching current members,
// merging with new members, and updating the project.
// Existing members are deduplicated by reference to avoid the server
// rejecting the request with "may be specified only once".
func (r *ProjectResource) AddMembers(projectId string, members []services.ProjectMember) error {
	logging.Trace()

	project, err := r.service.Get(projectId)
	if err != nil {
		return err
	}

	// Index new members by reference so existing members with the same
	// reference are skipped rather than duplicated.
	newRefs := make(map[string]bool, len(members))
	for _, m := range members {
		newRefs[m.Reference] = true
	}

	for _, m := range project.Members {
		if !newRefs[m.Reference] {
			members = append(members, m)
		}
	}

	data := map[string]interface{}{
		"members": members,
	}
	return r.service.UpdateProject(projectId, data)
}
