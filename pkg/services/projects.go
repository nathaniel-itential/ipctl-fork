// Copyright 2024 Itential Inc. All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential

package services

import (
	"fmt"
	"net/http"

	"github.com/itential/ipctl/internal/logging"
	"github.com/itential/ipctl/pkg/client"
)

const (
	// defaultProjectPageSize is the number of projects to retrieve per API call
	// during pagination in GetAll operations.
	defaultProjectPageSize = 100

	// projectsBasePath is the base URI path for all project-related API endpoints.
	projectsBasePath = "/automation-studio/projects"

	// Member type constants
	MemberTypeAccount = "account"
	MemberTypeGroup   = "group"

	// Member role constants
	MemberRoleOwner    = "owner"
	MemberRoleEditor   = "editor"
	MemberRoleOperator = "operator"
	MemberRoleViewer   = "viewer"
)

// ProjectComponent represents a single component within an Itential project.
// Components are the building blocks of a project and can include workflows,
// transformations, templates, and other automation assets.
type ProjectComponent struct {
	// Iid is the internal integer identifier for the component within the project.
	Iid int `json:"iid"`

	// Type identifies the component type (e.g., "workflow", "transformation").
	Type string `json:"type"`

	// Folder is the path to the folder containing this component within the project structure.
	Folder string `json:"folder"`

	// Reference is the unique identifier or name reference for the component.
	Reference string `json:"reference"`

	// Document contains the component's configuration and metadata as key-value pairs.
	Document map[string]interface{} `json:"document"`
}

// ProjectFolder represents a folder node in the project's hierarchical structure.
// Folders organize components and can contain nested folders, forming a tree structure.
type ProjectFolder struct {
	// Iid is the internal integer identifier for the folder.
	Iid int `json:"iid"`

	// Name is the display name of the folder.
	Name string `json:"name"`

	// NodeType identifies the type of node (typically "folder").
	NodeType string `json:"nodeType"`

	// Children contains nested folders within this folder, forming a recursive structure.
	Children []ProjectFolder `json:"children"`
}

// ProjectOperation represents the response structure for project API operations
// that return a single project with message and metadata.
type ProjectOperation struct {
	// Message contains a human-readable status message from the API.
	Message string `json:"message"`

	// Data contains the project data returned by the operation.
	Data Project `json:"data"`

	// Metadata contains additional information about the operation.
	Metadata Metadata `json:"metadata"`
}

// ProjectMember represents a user or group member assigned to a project.
// Members have specific roles that determine their permissions within the project.
type ProjectMember struct {
	// Provenance indicates the source system for this member (e.g., "local", "ldap").
	Provenance string `json:"provenance"`

	// Reference is the unique identifier for the member in the source system.
	Reference string `json:"reference"`

	// Role defines the member's permissions (e.g., "viewer", "editor", "admin").
	Role string `json:"role"`

	// Type identifies whether this is a "user" or "group" member.
	Type string `json:"type"`

	// Username is the member's username. Only populated for user-type members.
	Username string `json:"username,omitempty"`

	// Name is the display name of the member. Only populated for user-type members.
	Name string `json:"name,omitempty"`
}

// ProjectAccessControl defines fine-grained access control for project operations.
// Each permission level contains a list of user/group identifiers.
type ProjectAccessControl struct {
	// Manage contains identifiers of users/groups who can manage project settings and members.
	Manage []string `json:"manage"`

	// Write contains identifiers of users/groups who can modify project content.
	Write []string `json:"write"`

	// Execute contains identifiers of users/groups who can execute project automations.
	Execute []string `json:"execute"`

	// Read contains identifiers of users/groups who can view project content.
	Read []string `json:"read"`
}

// Project represents an Itential Automation Platform project.
// A project is a container for automation components, providing organization,
// versioning, and access control for related automation assets.
type Project struct {
	// Id is the unique MongoDB identifier for the project.
	Id string `json:"_id"`

	// Name is the human-readable name of the project.
	Name string `json:"name"`

	// BackgroundColor is the hex color code used for the project's UI representation.
	BackgroundColor string `json:"backgroundColor"`

	// Components contains all automation assets included in the project.
	Components []ProjectComponent `json:"components"`

	// Created is the ISO 8601 timestamp when the project was created.
	Created string `json:"created"`

	// CreatedBy contains information about the user who created the project.
	// Can be a string username or a map with user details.
	CreatedBy any `json:"createdBy"`

	// Description provides detailed information about the project's purpose.
	Description string `json:"description"`

	// Folders defines the hierarchical folder structure for organizing components.
	Folders []ProjectFolder `json:"folders"`

	// Iid is the internal integer identifier for the project.
	Iid int `json:"iid"`

	// ComponentIidIndex tracks the next available Iid for new components.
	// Not supported for import operations.
	ComponentIidIndex int `json:"componentIidIndex"`

	// LastUpdated is the ISO 8601 timestamp of the last modification.
	LastUpdated string `json:"lastUpdated"`

	// LastUpdatedBy contains information about the user who last modified the project.
	// Can be a string username or a map with user details.
	LastUpdatedBy any `json:"lastUpdatedBy"`

	// Thumbnail is the base64-encoded image data for the project icon.
	Thumbnail string `json:"thumbnail,omitempty"`

	// Members lists all users and groups with access to the project.
	// Not supported for import operations.
	Members []ProjectMember `json:"members"`

	// AccessControl defines fine-grained permission levels for the project.
	// Not supported for import operations.
	AccessControl ProjectAccessControl `json:"accessControl"`

	// ReferencedComponentHashes lists components that the project references
	// but does not contain, each paired with a content hash the platform uses
	// to validate the reference during import. It is modeled as a slice of maps
	// rather than a typed struct so that every platform-provided field is
	// preserved on round-trip. Omitted from the payload when the project has no
	// referenced components, matching the platform's UI export.
	ReferencedComponentHashes []map[string]interface{} `json:"referencedComponentHashes,omitempty"`
}

// Import returns a map representation of the Project suitable for importing.
// This method excludes server-managed fields that should not be included when
// importing a project: componentIidIndex (auto-incremented), members (managed separately),
// and accessControl (managed separately).
//
// Use this method when preparing a project for import via the Import API,
// or when serializing a project for export to another environment.
func (p Project) Import() map[string]interface{} {
	logging.Trace()

	// Pre-allocate map with exact capacity to avoid reallocations
	result := make(map[string]interface{}, 13)

	result["_id"] = p.Id
	result["name"] = p.Name
	result["backgroundColor"] = p.BackgroundColor
	result["components"] = p.Components
	result["created"] = p.Created
	result["createdBy"] = p.CreatedBy
	result["description"] = p.Description
	result["folders"] = p.Folders
	result["iid"] = p.Iid
	result["lastUpdated"] = p.LastUpdated
	result["lastUpdatedBy"] = p.LastUpdatedBy
	result["thumbnail"] = p.Thumbnail

	// Preserve referenced component hashes so the platform can validate
	// externally referenced components on import. Only include the key when
	// present to avoid sending a null value for projects without references.
	if len(p.ReferencedComponentHashes) > 0 {
		result["referencedComponentHashes"] = p.ReferencedComponentHashes
	}

	return result
}

// ProjectService provides operations for managing Itential projects.
// It handles project CRUD operations, import/export, and member management
// through the Automation Studio API.
type ProjectService struct {
	BaseService
}

// NewProjectService creates a new ProjectService instance with the provided HTTP client.
// The client should be configured with appropriate authentication credentials and base URL.
func NewProjectService(c client.Client) *ProjectService {
	return &ProjectService{
		BaseService: NewBaseService(c),
	}
}

// GetAll retrieves all projects from the Automation Studio API.
// It automatically handles pagination, fetching all pages until all projects are retrieved.
//
// Returns an empty slice if no projects exist on the server.
// Returns an error if any API call fails during pagination.
func (svc *ProjectService) GetAll() ([]Project, error) {
	logging.Trace()

	type getAllResponse struct {
		Message  string    `json:"message"`
		Data     []Project `json:"data"`
		Metadata Metadata  `json:"metadata"`
	}

	var res getAllResponse

	// Pre-allocate slice with a reasonable initial capacity to reduce allocations
	projects := make([]Project, 0, defaultProjectPageSize)

	limit := defaultProjectPageSize
	skip := 0

	for {
		if err := svc.GetRequest(&Request{
			uri:    projectsBasePath,
			params: &QueryParams{Limit: limit, Skip: skip},
		}, &res); err != nil {
			return nil, fmt.Errorf("failed to retrieve projects (skip=%d, limit=%d): %w", skip, limit, err)
		}

		// Append results efficiently
		projects = append(projects, res.Data...)

		// Check if we've retrieved all projects
		if len(projects) >= res.Metadata.Total {
			break
		}

		skip += limit
	}

	logging.Info("Found %d project(s)", len(projects))

	return projects, nil
}

// Get retrieves a single project by its unique identifier.
//
// The id parameter must be a valid MongoDB ObjectId string.
// Returns an error if the project does not exist or if the API call fails.
func (svc *ProjectService) Get(id string) (*Project, error) {
	logging.Trace()

	if id == "" {
		return nil, fmt.Errorf("project id cannot be empty")
	}

	type getResponse struct {
		Message  string   `json:"message"`
		Data     *Project `json:"data"`
		Metadata Metadata `json:"metadata"`
	}

	var res getResponse

	uri := fmt.Sprintf("%s/%s", projectsBasePath, id)

	if err := svc.BaseService.Get(uri, &res); err != nil {
		return nil, fmt.Errorf("failed to get project %s: %w", id, err)
	}

	logging.Info("%s", res.Message)

	return res.Data, nil
}

// Create creates a new project with the specified name.
//
// The name parameter must not be empty. The created project will have
// default settings and an empty components list.
//
// Returns the newly created project with server-assigned fields populated,
// or an error if creation fails.
func (svc *ProjectService) Create(name string) (*Project, error) {
	logging.Trace()

	if name == "" {
		return nil, fmt.Errorf("project name cannot be empty")
	}

	body := map[string]interface{}{
		"name":       name,
		"components": []string{},
	}

	type createResponse struct {
		Message  string   `json:"message"`
		Data     *Project `json:"data"`
		Metadata any      `json:"metadata"`
	}

	var res createResponse

	if err := svc.PostRequest(&Request{
		uri:                projectsBasePath,
		body:               body,
		expectedStatusCode: http.StatusOK,
	}, &res); err != nil {
		return nil, fmt.Errorf("failed to create project '%s': %w", name, err)
	}

	logging.Info("%s", res.Message)

	return res.Data, nil
}

// Delete removes a project by its unique identifier.
//
// This is a destructive operation that cannot be undone. All components
// and configuration within the project will be deleted.
//
// The id parameter must be a valid MongoDB ObjectId string.
// Returns an error if the project does not exist or if the deletion fails.
func (svc *ProjectService) Delete(id string) error {
	logging.Trace()

	if id == "" {
		return fmt.Errorf("project id cannot be empty")
	}

	uri := fmt.Sprintf("%s/%s", projectsBasePath, id)

	if err := svc.BaseService.Delete(uri); err != nil {
		return fmt.Errorf("failed to delete project %s: %w", id, err)
	}

	return nil
}

// Import imports a project using pre-transformed data.
//
// The data parameter should be prepared with the proper structure for import,
// typically by using the Project.Import() method to exclude non-importable fields
// like componentIidIndex, members, and accessControl.
//
// This method is useful for migrating projects between environments or
// restoring projects from backups. Components referenced in the project
// must already exist on the target server.
//
// Returns the imported project with server-assigned fields populated,
// or an error if the import fails.
func (svc *ProjectService) Import(data map[string]interface{}) (*Project, error) {
	logging.Trace()

	if data == nil {
		return nil, fmt.Errorf("import data cannot be nil")
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("import data cannot be empty")
	}

	type importResponse struct {
		Message  string                 `json:"message"`
		Data     *Project               `json:"data"`
		Metadata map[string]interface{} `json:"metadata"`
	}

	var res importResponse

	uri := fmt.Sprintf("%s/import", projectsBasePath)

	if err := svc.PostRequest(&Request{
		uri:                uri,
		body:               data,
		expectedStatusCode: http.StatusOK,
	}, &res); err != nil {
		return nil, fmt.Errorf("failed to import project: %w", err)
	}

	logging.Info("%s", res.Message)

	return res.Data, nil
}

// Export retrieves a project in export format by its identifier.
//
// The exported format includes all project components and configuration
// suitable for transfer to another environment or for backup purposes.
// The export may include additional metadata not present in standard Get operations.
//
// The id parameter must be a valid MongoDB ObjectId string.
// Returns the project data in export format, or an error if the export fails.
func (svc *ProjectService) Export(id string) (*Project, error) {
	logging.Trace()

	if id == "" {
		return nil, fmt.Errorf("project id cannot be empty")
	}

	type exportResponse struct {
		Message  string   `json:"message"`
		Data     *Project `json:"data"`
		Metadata Metadata `json:"metadata"`
	}

	var res exportResponse

	uri := fmt.Sprintf("%s/%s/export", projectsBasePath, id)

	if err := svc.BaseService.Get(uri, &res); err != nil {
		return nil, fmt.Errorf("failed to export project %s: %w", id, err)
	}

	logging.Info("%s", res.Message)

	return res.Data, nil
}

// UpdateProject updates a project via PATCH request.
//
// This method accepts a map of fields to update. Common fields include:
//   - members: []ProjectMember - replaces the entire members list
//   - name: string - updates the project name
//   - description: string - updates the project description
//   - backgroundColor: string - updates the background color
//
// To update only members, use a map with the "members" key:
//
//	data := map[string]interface{}{"members": members}
//
// The projectId parameter must be a valid MongoDB ObjectId string.
// The data parameter should contain the fields to update.
//
// Returns an error if the update fails or if the project does not exist.
func (svc *ProjectService) UpdateProject(projectId string, data map[string]interface{}) error {
	logging.Trace()

	if projectId == "" {
		return fmt.Errorf("project id cannot be empty")
	}

	if data == nil || len(data) == 0 {
		return fmt.Errorf("update data cannot be nil or empty")
	}

	uri := fmt.Sprintf("%s/%s", projectsBasePath, projectId)

	if err := svc.Patch(uri, data, nil); err != nil {
		return fmt.Errorf("failed to update project %s: %w", projectId, err)
	}

	return nil
}
