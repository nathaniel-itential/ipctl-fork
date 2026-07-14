// Copyright 2024 Itential Inc. All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential

package runners

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/itential/ipctl/internal/config"
	"github.com/itential/ipctl/internal/flags"
	"github.com/itential/ipctl/internal/logging"
	"github.com/itential/ipctl/pkg/client"
	"github.com/itential/ipctl/pkg/resources"
	"github.com/itential/ipctl/pkg/services"
)

// AgentProjectRunner orchestrates CLI commands for agent project management.
// It implements the Reader, Importer, and Exporter interfaces.
type AgentProjectRunner struct {
	BaseRunner
	resource     resources.AgentProjectResourcer
	accounts     *services.AccountService
	groups       *services.GroupService
	userSettings *services.UserSettingsService
}

// NewAgentProjectRunner creates a new AgentProjectRunner with the provided client and config.
func NewAgentProjectRunner(c client.Client, cfg config.Provider) *AgentProjectRunner {
	return &AgentProjectRunner{
		BaseRunner:   NewBaseRunner(c, cfg),
		resource:     resources.NewAgentProjectResource(services.NewAgentProjectService(c)),
		accounts:     services.NewAccountService(c),
		groups:       services.NewGroupService(c),
		userSettings: services.NewUserSettingsService(c),
	}
}

//////////////////////////////////////////////////////////////////////////////
// Reader Interface
//

// Get retrieves all agent projects and returns them for display.
func (r *AgentProjectRunner) Get(in Request) (*Response, error) {
	logging.Trace()

	projects, err := r.resource.GetAll()
	if err != nil {
		return nil, err
	}

	return &Response{
		Keys:   []string{"name", "description"},
		Object: projects,
	}, nil
}

// Describe retrieves detailed information about a specific agent project.
func (r *AgentProjectRunner) Describe(in Request) (*Response, error) {
	logging.Trace()

	project, err := r.resource.GetByName(in.Args[0])
	if err != nil {
		return nil, err
	}

	createdBy := extractAgentProjectUsername(project.CreatedBy, "unknown")
	updatedBy := extractAgentProjectUsername(project.LastUpdatedBy, "unknown")

	output := []string{
		fmt.Sprintf("Name: %s (%s)", project.Name, project.Id),
		fmt.Sprintf("Description: %s", project.Description),
		fmt.Sprintf("Created: %s, by: %s", project.Created, createdBy),
		fmt.Sprintf("Updated: %s, by: %s", project.LastUpdated, updatedBy),
		fmt.Sprintf("Components: %d", len(project.Components)),
	}

	return &Response{
		Text:   strings.Join(output, "\n"),
		Object: project,
	}, nil
}

//////////////////////////////////////////////////////////////////////////////
// Importer Interface
//

// Import imports an agent project bundle from a local file or Git repository.
// Optionally adds members to the imported project if specified via the --member flag.
func (r *AgentProjectRunner) Import(in Request) (*Response, error) {
	logging.Trace()

	common := in.Common.(*flags.AssetImportCommon)
	options := in.Options.(*flags.AgentProjectImportOptions)

	path, err := importGetPathFromRequest(in)
	if err != nil {
		return nil, err
	}

	wd := filepath.Dir(path)

	if common.Repository != "" {
		defer os.RemoveAll(wd)
	}

	var bundle services.AgentProjectBundle

	if err := importLoadFromDisk(path, &bundle); err != nil {
		return nil, err
	}

	// conflictModeExplicit tracks whether the caller explicitly chose a --conflict-mode
	// (e.g. "keep-both") rather than relying on the --replace-derived default. When explicit,
	// the pre-import existence check below must be skipped so the server can perform its own
	// conflict handling (such as server-side duplication for keep-both) on a name collision.
	conflictModeExplicit := options.ConflictMode != ""

	if !common.Replace && !conflictModeExplicit {
		existing, err := r.resource.GetByName(bundle.Name)
		if err == nil && existing != nil {
			return nil, fmt.Errorf("agent project %q already exists, use --replace to overwrite", bundle.Name)
		}
	}

	conflictMode := options.ConflictMode
	if conflictMode == "" {
		// --conflict-mode wasn't explicitly set: --replace implies "replace", otherwise default to "keep-both".
		if common.Replace {
			conflictMode = "replace"
		} else {
			conflictMode = "keep-both"
		}
	}

	if conflictMode != "keep-both" && conflictMode != "replace" {
		return nil, fmt.Errorf("invalid --conflict-mode %q (must be 'keep-both' or 'replace')", conflictMode)
	}

	imported, err := r.resource.Import(bundle, conflictMode)
	if err != nil {
		return nil, err
	}

	if err := r.updateMembers(imported.Id, options.Members); err != nil {
		// Cleanup: delete the partially imported project. This is only safe when the import
		// created a brand-new project. In "replace" mode, the import overwrote an existing
		// project in place, so deleting it would destroy the user's only remaining copy of
		// their data rather than rolling back a fresh creation.
		if conflictMode != "replace" {
			if delErr := r.resource.Delete(imported.Id); delErr != nil {
				logging.Error(delErr, "failed to cleanup agent project %s after member update error", imported.Id)
			}
		} else {
			logging.Error(err, "member update failed after replacing agent project %s; skipping cleanup delete to avoid destroying the replaced project", imported.Id)
		}
		return nil, fmt.Errorf("failed to update agent project members: %w", err)
	}

	return &Response{
		Text: fmt.Sprintf("Successfully imported agent project `%s` (%s)", imported.Name, imported.Id),
	}, nil
}

//////////////////////////////////////////////////////////////////////////////
// Exporter Interface
//

// Export exports an agent project bundle to a local file or Git repository.
func (r *AgentProjectRunner) Export(in Request) (*Response, error) {
	logging.Trace()

	name := in.Args[0]

	project, err := r.resource.GetByName(name)
	if err != nil {
		return nil, err
	}

	bundle, err := r.resource.Export(project.Id)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(bundle)
	if err != nil {
		return nil, err
	}

	var exported map[string]interface{}
	if err := json.Unmarshal(b, &exported); err != nil {
		return nil, err
	}

	fn := fmt.Sprintf("%s.agent-project.json", normalizeFilename(name))

	if err := exportAssetFromRequest(in, exported, fn); err != nil {
		return nil, err
	}

	return &Response{
		Text: fmt.Sprintf("Successfully exported agent project `%s`", bundle.Name),
	}, nil
}

//////////////////////////////////////////////////////////////////////////////
// Private helpers
//

// extractAgentProjectUsername safely extracts a username from a user object.
func extractAgentProjectUsername(userObj any, fallback string) string {
	if userObj == nil {
		return fallback
	}

	userMap, ok := userObj.(map[string]interface{})
	if !ok {
		return fallback
	}

	username, ok := userMap["username"].(string)
	if !ok {
		return fallback
	}

	return username
}

// AgentProjectMemberSpec represents an agent project member specification parsed
// from CLI flags. It is used internally for parsing member specifications from
// CLI flags and constructing AgentProjectMember objects for API calls.
//
// Type must be either "account" or "group" (use constants services.MemberTypeAccount or services.MemberTypeGroup).
// Access must be one of "owner", "editor", "operator", or "viewer" (use constants services.MemberRole*).
// Name is the username (for accounts) or group name (for groups).
type AgentProjectMemberSpec struct {
	Type   string // "account" or "group"
	Name   string // Username or group name
	Access string // "owner", "editor", "operator", or "viewer"
}

// updateMembers adds members to an agent project after it has been created or imported.
//
// This helper method is used internally by Import to add members to an agent project.
// It parses member specifications, resolves accounts and groups, and adds them to the
// project while automatically excluding the active user.
//
// Parameters:
//   - projectId: The ID of the agent project to add members to
//   - projectMembers: Slice of member specification strings (format: "type=account,name=alice,access=editor")
//
// Returns:
//   - nil on success
//   - Error if member parsing, resolution, or addition fails
func (r *AgentProjectRunner) updateMembers(projectId string, projectMembers []string) error {
	logging.Trace()

	if len(projectMembers) == 0 {
		return nil // No members to update
	}

	activeUser, err := r.userSettings.Get()
	if err != nil {
		return fmt.Errorf("failed to get active user: %w", err)
	}

	members, err := r.buildAgentProjectMembers(projectMembers, activeUser.Username, r.accounts, r.groups)
	if err != nil {
		return err
	}

	if len(members) > 0 {
		if err := r.resource.AddMembers(projectId, members); err != nil {
			return fmt.Errorf("failed to add members to agent project: %w", err)
		}
	}

	return nil
}

// resolveAgentProjectMember resolves an AgentProjectMemberSpec into an AgentProjectMember
// by looking up the account or group and populating all required fields.
func (r *AgentProjectRunner) resolveAgentProjectMember(
	member *AgentProjectMemberSpec,
	accounts *services.AccountService,
	groups *services.GroupService,
) (services.AgentProjectMember, error) {
	switch member.Type {
	case services.MemberTypeAccount:
		account, err := accounts.GetByName(member.Name)
		if err != nil {
			return services.AgentProjectMember{}, fmt.Errorf("account %q not found: %w", member.Name, err)
		}
		return services.AgentProjectMember{
			Provenance: account.Provenance,
			Reference:  account.Id,
			Role:       member.Access,
			Type:       services.MemberTypeAccount,
			Username:   account.Username,
		}, nil

	case services.MemberTypeGroup:
		group, err := groups.GetByName(member.Name)
		if err != nil {
			return services.AgentProjectMember{}, fmt.Errorf("group %q not found: %w", member.Name, err)
		}
		return services.AgentProjectMember{
			Provenance: group.Provenance,
			Reference:  group.Id,
			Role:       member.Access,
			Type:       services.MemberTypeGroup,
			Name:       group.Name,
		}, nil

	default:
		return services.AgentProjectMember{}, fmt.Errorf("invalid member type %q (must be 'account' or 'group')", member.Type)
	}
}

// buildAgentProjectMembers converts member specifications into AgentProjectMember objects.
// It resolves accounts and groups by name, skips the active user, and validates
// member types and access levels.
//
// Parameters:
//   - memberSpecs: Slice of member specification strings (format: "type=account,name=alice,access=editor")
//   - activeUsername: Username of the currently authenticated user (will be skipped)
//   - accounts: Account service for resolving account names
//   - groups: Group service for resolving group names
//
// Returns:
//   - Slice of AgentProjectMember objects ready for API submission
//   - Error if member parsing fails, member not found, or resolution fails
func (r *AgentProjectRunner) buildAgentProjectMembers(
	memberSpecs []string,
	activeUsername string,
	accounts *services.AccountService,
	groups *services.GroupService,
) ([]services.AgentProjectMember, error) {
	logging.Trace()

	if len(memberSpecs) == 0 {
		return nil, nil
	}

	var members []services.AgentProjectMember

	for _, spec := range memberSpecs {
		member, err := parseAgentProjectMember(spec)
		if err != nil {
			return nil, fmt.Errorf("invalid member specification %q: %w", spec, err)
		}

		// Skip active user
		if member.Type == services.MemberTypeAccount && member.Name == activeUsername {
			logging.Info("skipping active user %q from member list", member.Name)
			continue
		}

		projectMember, err := r.resolveAgentProjectMember(member, accounts, groups)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve member %q: %w", member.Name, err)
		}

		members = append(members, projectMember)
	}

	return members, nil
}

// parseAgentProjectMember parses a member specification string into an AgentProjectMemberSpec.
// The format is: "type=<account|group>,name=<name>[,access=<role>]"
//
// Parameters:
//   - member: Member specification string
//
// Returns:
//   - Parsed AgentProjectMemberSpec with defaults applied
//   - Error if format is invalid or required fields are missing
//
// Example valid inputs:
//   - "type=account,name=alice"
//   - "type=account,name=alice,access=owner"
//   - "type=group,name=devops,access=editor"
func parseAgentProjectMember(member string) (*AgentProjectMemberSpec, error) {
	if member == "" {
		return nil, fmt.Errorf("member specification cannot be empty")
	}

	parts := strings.Split(member, ",")
	m := &AgentProjectMemberSpec{
		Access: services.MemberRoleEditor, // Default access level
	}

	seen := make(map[string]bool, 3)

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		tokens := strings.SplitN(part, "=", 2)
		if len(tokens) != 2 {
			return nil, fmt.Errorf("invalid key=value pair %q in member specification %q", part, member)
		}

		key := strings.TrimSpace(tokens[0])
		value := strings.TrimSpace(tokens[1])

		if value == "" {
			return nil, fmt.Errorf("empty value for key %q in member specification %q", key, member)
		}

		if seen[key] {
			return nil, fmt.Errorf("duplicate key %q in member specification %q", key, member)
		}
		seen[key] = true

		switch key {
		case "type":
			m.Type = value
		case "name":
			m.Name = value
		case "access":
			m.Access = value
		default:
			return nil, fmt.Errorf("unknown key %q in member specification %q", key, member)
		}
	}

	// Validate required fields
	if m.Type == "" {
		return nil, fmt.Errorf("missing required 'type' field in member specification %q", member)
	}
	if m.Name == "" {
		return nil, fmt.Errorf("missing required 'name' field in member specification %q", member)
	}

	// Validate type
	if m.Type != services.MemberTypeAccount && m.Type != services.MemberTypeGroup {
		return nil, fmt.Errorf("invalid type %q (must be 'account' or 'group') in member specification %q", m.Type, member)
	}

	// Validate access
	validAccess := []string{
		services.MemberRoleOwner,
		services.MemberRoleEditor,
		services.MemberRoleOperator,
		services.MemberRoleViewer,
	}
	if !slices.Contains(validAccess, m.Access) {
		return nil, fmt.Errorf("invalid access %q (must be one of: owner, editor, operator, viewer) in member specification %q", m.Access, member)
	}

	return m, nil
}
