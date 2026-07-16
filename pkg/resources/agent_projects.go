// Copyright 2024 Itential Inc. All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential

package resources

import (
	"github.com/itential/ipctl/internal/logging"
	"github.com/itential/ipctl/pkg/services"
)

// AgentProjectResource provides business logic for agent project operations.
type AgentProjectResource struct {
	BaseResource
	service services.AgentProjectServicer
}

// NewAgentProjectResource creates a new AgentProjectResource with the given service.
func NewAgentProjectResource(svc services.AgentProjectServicer) AgentProjectResourcer {
	return &AgentProjectResource{
		BaseResource: NewBaseResource(),
		service:      svc,
	}
}

// GetAll retrieves all agent projects.
func (r *AgentProjectResource) GetAll() ([]services.AgentProject, error) {
	return r.service.GetAll()
}

// Get retrieves an agent project by ID.
func (r *AgentProjectResource) Get(id string) (*services.AgentProject, error) {
	return r.service.Get(id)
}

// GetByName retrieves an agent project by name.
func (r *AgentProjectResource) GetByName(name string) (*services.AgentProject, error) {
	logging.Trace()
	return r.service.GetByName(name)
}

// Export exports an agent project bundle by project ID.
func (r *AgentProjectResource) Export(id string) (*services.AgentProjectBundle, error) {
	return r.service.Export(id)
}

// Import imports an agent project bundle.
func (r *AgentProjectResource) Import(bundle services.AgentProjectBundle, conflictMode string) (*services.AgentProjectBundle, error) {
	return r.service.Import(bundle, conflictMode)
}

// Create creates a new agent project with the specified name and description.
func (r *AgentProjectResource) Create(name string, description string) (*services.AgentProject, error) {
	return r.service.Create(name, description)
}

// Delete removes an agent project by its identifier.
func (r *AgentProjectResource) Delete(id string) error {
	return r.service.Delete(id)
}

// AddMembers adds new members to an existing agent project.
// This method implements the business logic of fetching current members,
// merging with new members, and updating the project.
//
// The PATCH API only accepts "type", "reference", and "role" per member —
// it rejects members with any additional properties (e.g. username, name,
// provenance), so members are serialized down to that minimal shape here.
// It also rejects duplicate (type, reference) pairs, so if a newly specified
// member matches one the platform already assigned (e.g. the project creator),
// the new member's role wins.
func (r *AgentProjectResource) AddMembers(projectId string, members []services.AgentProjectMember) error {
	logging.Trace()

	project, err := r.service.Get(projectId)
	if err != nil {
		return err
	}

	type memberKey struct {
		Type      string
		Reference string
	}

	merged := make(map[memberKey]services.AgentProjectMember, len(members)+len(project.Members))

	for _, m := range project.Members {
		merged[memberKey{m.Type, m.Reference}] = m
	}

	for _, m := range members {
		merged[memberKey{m.Type, m.Reference}] = m
	}

	minimalMembers := make([]map[string]interface{}, 0, len(merged))
	for _, m := range merged {
		minimalMembers = append(minimalMembers, map[string]interface{}{
			"type":      m.Type,
			"reference": m.Reference,
			"role":      m.Role,
		})
	}

	data := map[string]interface{}{
		"members": minimalMembers,
	}
	return r.service.UpdateProject(projectId, data)
}
