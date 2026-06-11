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
	"github.com/itential/ipctl/internal/utils"
	"github.com/itential/ipctl/pkg/client"
	"github.com/itential/ipctl/pkg/resources"
	"github.com/itential/ipctl/pkg/services"
)

// ProjectRunner orchestrates CLI commands for project management operations.
// It handles CRUD operations, import/export with Git repositories, and member
// management across Itential Platform servers.
//
// ProjectRunner implements the Reader, Writer, Copier, Importer, and Exporter
// interfaces, providing comprehensive project lifecycle management through
// the CLI.
//
// Example usage:
//
//	runner := NewProjectRunner(client, config)
//	response, err := runner.Get(Request{})
type ProjectRunner struct {
	BaseRunner
	resource     resources.ProjectResourcer
	accounts     *services.AccountService
	groups       *services.GroupService
	userSettings *services.UserSettingsService
}

// NewProjectRunner creates a new ProjectRunner instance with the provided client and configuration.
//
// The runner is initialized with all necessary service dependencies for project management,
// including project resources, account services, group services, and user settings.
//
// Parameters:
//   - client: HTTP client for making API requests to the Itential Platform
//   - cfg: Configuration provider for accessing application settings
//
// Returns:
//   - A fully initialized ProjectRunner ready to handle CLI commands
func NewProjectRunner(client client.Client, cfg config.Provider) *ProjectRunner {
	return &ProjectRunner{
		BaseRunner:   NewBaseRunner(client, cfg),
		resource:     resources.NewProjectResource(services.NewProjectService(client)),
		accounts:     services.NewAccountService(client),
		groups:       services.NewGroupService(client),
		userSettings: services.NewUserSettingsService(client),
	}
}

//////////////////////////////////////////////////////////////////////////////
// Reader Interface
//

// Get retrieves all projects from the Itential Platform and returns them in a format
// suitable for display in the CLI.
//
// This method implements the Reader interface and handles the `get projects` command.
// It fetches all projects and returns them with keys for name and description columns.
//
// Parameters:
//   - in: Request containing CLI arguments and options (unused for this operation)
//
// Returns:
//   - Response with project list and display keys for name and description
//   - Error if the API request fails
func (r *ProjectRunner) Get(in Request) (*Response, error) {
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

// extractUsername safely extracts a username from a user object, returning a fallback if extraction fails.
func extractUsername(userObj any, fallback string) string {
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

// Describe retrieves detailed information about a specific project and formats it
// for display in the CLI.
//
// This method implements the Reader interface and handles the `describe project <name>` command.
// It fetches the project by name and returns formatted details including name, ID, description,
// creation and update timestamps with user information.
//
// Parameters:
//   - in: Request with Args[0] containing the project name to describe
//
// Returns:
//   - Response with formatted text output and the complete project object
//   - Error if the project is not found or the API request fails
func (r *ProjectRunner) Describe(in Request) (*Response, error) {
	logging.Trace()

	var res *services.Project

	res, err := r.resource.GetByName(in.Args[0])
	if err != nil {
		return nil, err
	}

	createdBy := extractUsername(res.CreatedBy, "unknown")
	updatedBy := extractUsername(res.LastUpdatedBy, "unknown")

	output := []string{
		fmt.Sprintf("Name: %s (%s)", res.Name, res.Id),
		fmt.Sprintf("Description: %s", res.Description),
		fmt.Sprintf("Created: %s, by: %s", res.Created, createdBy),
		fmt.Sprintf("Updated: %s, by: %s", res.LastUpdated, updatedBy),
	}

	return &Response{
		Text:   strings.Join(output, "\n"),
		Object: res,
	}, nil
}

//////////////////////////////////////////////////////////////////////////////
// Writer Interface
//

// Create creates a new project with the specified name on the Itential Platform.
//
// This method implements the Writer interface and handles the `create project <name>` command.
// It first checks if a project with the same name already exists and returns an error if found.
// On success, it creates the project and returns confirmation with the project ID.
//
// Parameters:
//   - in: Request with Args[0] containing the name for the new project
//
// Returns:
//   - Response with success message and the created project object
//   - Error if a project with the same name already exists or the API request fails
func (r *ProjectRunner) Create(in Request) (*Response, error) {
	logging.Trace()

	name := in.Args[0]

	existing, err := r.resource.GetByName(name)
	if existing != nil {
		return nil, fmt.Errorf("project %q already exists", name)
	}

	project, err := r.resource.Create(name)
	if err != nil {
		return nil, err
	}

	return &Response{
		Text:   fmt.Sprintf("Successfully created project `%s` (%s)", project.Name, project.Id),
		Object: project,
	}, nil
}

// Delete removes a project from the Itential Platform by name.
//
// This method implements the Writer interface and handles the `delete project <name>` command.
// It first looks up the project by name to get its ID, then performs the deletion.
//
// Parameters:
//   - in: Request with Args[0] containing the name of the project to delete
//
// Returns:
//   - Response with success message including the project name and ID
//   - Error if the project is not found or the deletion fails
func (r *ProjectRunner) Delete(in Request) (*Response, error) {
	logging.Trace()

	project, err := r.resource.GetByName(in.Args[0])
	if err != nil {
		return nil, err
	}

	if err := r.resource.Delete(project.Id); err != nil {
		return nil, err
	}

	return &Response{
		Text: fmt.Sprintf("Successfully deleted project `%s` (%s)", project.Name, project.Id),
	}, nil
}

// Clear removes all projects from the Itential Platform.
//
// This method implements the Writer interface and handles the `clear projects` command.
// It retrieves all projects and deletes them one by one. If any deletion fails, the
// operation stops and returns an error.
//
// Parameters:
//   - in: Request containing CLI arguments and options (unused for this operation)
//
// Returns:
//   - Response with count of deleted projects
//   - Error if fetching projects fails or any deletion fails
func (r *ProjectRunner) Clear(in Request) (*Response, error) {
	logging.Trace()

	projects, err := r.resource.GetAll()
	if err != nil {
		return nil, err
	}

	for _, ele := range projects {
		if err := r.resource.Delete(ele.Id); err != nil {
			logging.Debug("failed to delete project `%s` (%s)", ele.Name, ele.Id)
			return nil, err
		}
	}

	return &Response{
		Text: fmt.Sprintf("Deleted %v project(s)", len(projects)),
	}, nil
}

//////////////////////////////////////////////////////////////////////////////
// Copier Interface
//

// Copy copies a project from the source profile to a destination profile, including
// project members if specified.
//
// This method implements the Copier interface and handles the `copy project <name> <dst>` command.
// It exports the project from the source, imports it to the destination, and optionally adds
// members to the copied project. The active user is automatically excluded from the member list.
//
// Parameters:
//   - in: Request containing the project name, destination profile, and optional member specifications
//
// Returns:
//   - Response with success message indicating the copy operation completed
//   - Error if the copy fails, member resolution fails, or member addition fails
func (r *ProjectRunner) Copy(in Request) (*Response, error) {
	logging.Trace()

	res, err := Copy(CopyRequest{Request: in, Type: "project"}, r)
	if err != nil {
		return nil, err
	}

	client, cancel, err := NewClient(in.Common.(*flags.AssetCopyCommon).To, r.config)
	if err != nil {
		return nil, err
	}
	defer cancel()

	projectsvc := services.NewProjectService(client)
	accounts := services.NewAccountService(client)
	groups := services.NewGroupService(client)
	userSettings := services.NewUserSettingsService(client)

	activeUser, err := userSettings.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get active user: %w", err)
	}

	members, err := r.buildProjectMembers(
		in.Options.(*flags.ProjectCopyOptions).Members,
		activeUser.Username,
		accounts,
		groups,
	)
	if err != nil {
		return nil, err
	}

	if len(members) > 0 {
		projectRes := resources.NewProjectResource(projectsvc)
		if err := projectRes.AddMembers(res.CopyToData.(services.Project).Id, members); err != nil {
			return nil, fmt.Errorf("failed to add members: %w", err)
		}
	}

	return &Response{
		Text: fmt.Sprintf("Successfully copied project `%s` from `%s` to `%s`", res.Name, res.From, res.To),
	}, nil
}

// CopyFrom exports a project from the specified profile for copying to another profile.
//
// This method is part of the Copier interface implementation and is called internally
// by the Copy method. It exports the project in a format suitable for import.
//
// Parameters:
//   - profile: Name of the profile to export from
//   - name: Name of the project to export
//
// Returns:
//   - The exported project data as an interface{}
//   - Error if the client creation, project lookup, or export fails
func (r *ProjectRunner) CopyFrom(profile, name string) (any, error) {
	logging.Trace()

	client, cancel, err := NewClient(profile, r.config)
	if err != nil {
		return nil, err
	}
	defer cancel()

	projectRes := resources.NewProjectResource(services.NewProjectService(client))

	project, err := projectRes.GetByName(name)
	if err != nil {
		return nil, err
	}

	res, err := projectRes.Export(project.Id)
	if err != nil {
		return nil, err
	}

	return *res, err
}

// CopyTo imports a project to the specified profile as part of the copy operation.
//
// This method is part of the Copier interface implementation and is called internally
// by the Copy method. It imports the exported project data to the destination profile.
// If a project with the same name already exists, it returns an error unless replace is true.
//
// Parameters:
//   - profile: Name of the profile to import to
//   - in: Project data to import (must be services.Project type)
//   - replace: If true, allows overwriting an existing project with the same name
//
// Returns:
//   - The imported project data
//   - Error if type assertion fails, project exists and replace is false, or import fails
func (r *ProjectRunner) CopyTo(profile string, in any, replace bool) (any, error) {
	logging.Trace()

	client, cancel, err := NewClient(profile, r.config)
	if err != nil {
		return nil, err
	}
	defer cancel()

	projectRes := resources.NewProjectResource(services.NewProjectService(client))

	project, ok := in.(services.Project)
	if !ok {
		return nil, fmt.Errorf("expected services.Project, got %T", in)
	}

	if exists, err := projectRes.GetByName(project.Name); exists != nil {
		if !replace {
			return nil, fmt.Errorf("project %q exists on the destination server, use --replace to overwrite", project.Name)
		} else if err != nil {
			return nil, err
		}
	}

	return projectRes.Import(project)
}

//////////////////////////////////////////////////////////////////////////////
// Importer Interface
//

// Import imports a project from a local file or Git repository into the Itential Platform.
//
// This method implements the Importer interface and handles the `import project <path>` command.
// It supports importing from local files, Git repositories (via --repository flag), and
// can handle both standard and expanded project formats. Optionally adds members to the imported
// project if specified via the --members flag.
//
// Parameters:
//   - in: Request containing the file path, repository URL (optional), and member specifications (optional)
//
// Returns:
//   - Response with success message including the project name and ID
//   - Error if file reading fails, import fails, or member addition fails
//
// Side effects:
//   - If importing from a Git repository, clones to a temporary directory that is cleaned up
//   - If member addition fails, attempts to delete the partially imported project
func (r *ProjectRunner) Import(in Request) (*Response, error) {
	logging.Trace()

	common := in.Common.(*flags.AssetImportCommon)
	options := in.Options.(*flags.ProjectImportOptions)

	path, err := importGetPathFromRequest(in)
	if err != nil {
		return nil, err
	}

	wd := filepath.Dir(path)

	if common.Repository != "" {
		defer os.RemoveAll(wd)
	}

	var project services.Project

	if err := importLoadFromDisk(path, &project); err != nil {
		return nil, err
	}

	imported, err := r.importProject(project, path, common.Replace)
	if err != nil {
		return nil, err
	}

	if err := r.updateMembers(imported.Id, options.Members); err != nil {
		// Cleanup: delete the partially imported project
		if delErr := r.resource.Delete(imported.Id); delErr != nil {
			logging.Error(delErr, "failed to cleanup project %s after member update error", imported.Id)
		}
		return nil, fmt.Errorf("failed to update project members: %w", err)
	}

	return &Response{
		Text: fmt.Sprintf("Successfully imported project `%s` (%s)", project.Name, project.Id),
	}, nil
}

/*
*******************************************************************************
Exporter interface
*******************************************************************************
*/

// Export exports a project from the Itential Platform to a local file or Git repository.
//
// This method implements the Exporter interface and handles the `export project <name>` command.
// It supports exporting to local files, Git repositories (via --repository flag), and
// can export in either standard (single file) or expanded format (multiple files, one per component).
//
// Parameters:
//   - in: Request containing the project name, export path, repository URL (optional), and expand flag (optional)
//
// Returns:
//   - Response with success message indicating the export completed
//   - Error if project lookup fails, export fails, file writing fails, or Git operations fail
//
// Side effects:
//   - Writes files to disk at the specified path
//   - If exporting to a Git repository, clones, commits, and pushes changes
//   - If using --expand, creates multiple files in a folder structure
func (r *ProjectRunner) Export(in Request) (*Response, error) {
	logging.Trace()

	common := in.Common.(*flags.AssetExportCommon)
	options := in.Options.(*flags.ProjectExportOptions)

	name := in.Args[0]

	p, err := r.resource.GetByName(name)
	if err != nil {
		return nil, err
	}

	project, err := r.resource.Export(p.Id)
	if err != nil {
		return nil, err
	}

	if options.Expand {
		path := common.Path

		var repo *Repository
		var repoPath string

		if common.Repository != "" {
			repo, err = exportNewRepositoryFromRequest(in)
			if err != nil {
				return nil, err
			}

			var e error

			repoPath, e = repo.Clone(&FileReaderImpl{}, &ClonerImpl{})
			if e != nil {
				return nil, e
			}
			defer os.RemoveAll(repoPath)

			path = filepath.Join(repoPath, common.Path)
		}

		if err := expandProject(in, project, path); err != nil {
			return nil, err
		}

		if common.Repository != "" {
			logging.Info("committing changes to repository at %s", repoPath)
			if err := repo.CommitAndPush(repoPath, common.Message); err != nil {
				return nil, err
			}
		}

	} else {
		b, err := json.Marshal(project)
		if err != nil {
			return nil, err
		}

		var exported map[string]interface{}
		if err := json.Unmarshal(b, &exported); err != nil {
			return nil, err
		}

		// Remove server-managed fields so the exported file matches the format
		// produced by the platform's UI export and imports cleanly: members and
		// accessControl are managed separately, and componentIidIndex is
		// re-derived by the server on import.
		delete(exported, "members")
		delete(exported, "accessControl")
		delete(exported, "componentIidIndex")

		fn := fmt.Sprintf("%s.project.json", strings.Replace(name, "/", "_", -1))

		if err := exportAssetFromRequest(in, exported, fn); err != nil {
			return nil, err
		}
	}

	return &Response{
		Text: fmt.Sprintf("Successfully exported project `%s`", project.Name),
	}, nil
}

//////////////////////////////////////////////////////////////////////////////
// Private functions
//

// Member represents a project member specification parsed from CLI flags.
// It is used internally for parsing member specifications from CLI flags
// and constructing ProjectMember objects for API calls.
//
// Type must be either "account" or "group" (use constants services.MemberTypeAccount or services.MemberTypeGroup).
// Access must be one of "owner", "editor", "operator", or "viewer" (use constants services.MemberRole*).
// Name is the username (for accounts) or group name (for groups).
type Member struct {
	Type   string // "account" or "group"
	Name   string // Username or group name
	Access string // "owner", "editor", "operator", or "viewer"
}

// importProject imports a project from disk into the Itential Platform.
//
// When a project is exported with the --expand flag, components are written to
// separate files referenced by their folder structure. This method reconstructs
// the complete project structure by:
//  1. Reading the project metadata from the specified path
//  2. Loading each component's document from its separate file
//  3. Assembling the complete project structure
//  4. Optionally deleting an existing project with the same name if replace=true
//  5. Importing the reconstructed project via the API
//
// Parameters:
//   - project: The project metadata loaded from the main project file
//   - path: Absolute path to the project file on disk
//   - replace: If true, deletes any existing project with the same name
//
// Returns:
//   - The imported project with server-assigned ID
//   - Error if path doesn't exist, file reading fails, or API import fails
//
// Side effects:
//   - May delete existing projects if replace=true
//   - Reads multiple files from disk in the project's folder structure
func (r *ProjectRunner) importProject(project services.Project, path string, replace bool) (*services.Project, error) {
	logging.Trace()

	var projectMap map[string]interface{}
	if err := importLoadFromDisk(path, &projectMap); err != nil {
		return nil, err
	}

	componentsRaw, ok := projectMap["components"]
	if !ok {
		return nil, fmt.Errorf("project missing 'components' field")
	}

	components, ok := componentsRaw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("project 'components' field has unexpected type %T", componentsRaw)
	}

	basepath := filepath.Dir(path)

	for idx, ele := range project.Components {
		if idx >= len(components) {
			continue
		}

		componentMap, ok := components[idx].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("component %d has unexpected type %T", idx, components[idx])
		}

		filenameVal, exists := componentMap["filename"]
		if !exists {
			continue
		}

		filename, ok := filenameVal.(string)
		if !ok {
			return nil, fmt.Errorf("component %d filename has unexpected type %T", idx, filenameVal)
		}

		fp := filepath.Join(basepath, normalizeProjectPath(ele.Folder), filename)

		doc, err := os.ReadFile(fp)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to read component file %q for component %q: an expanded project must be imported together with its component files: %w",
				fp, ele.Reference, err,
			)
		}

		var document map[string]interface{}
		if err := utils.UnmarshalData(doc, &document); err != nil {
			return nil, fmt.Errorf("failed to parse component file %q: %w", fp, err)
		}

		project.Components[idx].Document = document
	}

	existing, err := r.resource.GetByName(project.Name)
	if err == nil && existing != nil {
		if !replace {
			return nil, fmt.Errorf("project %q already exists, use --replace to overwrite it", project.Name)
		}
		if err := r.resource.Delete(existing.Id); err != nil {
			return nil, fmt.Errorf("failed to delete existing project: %w", err)
		}
	}

	return r.resource.Import(project)
}

// updateMembers adds members to a project after it has been created or imported.
//
// This helper method is used internally by Import and other operations to add members
// to a project. It parses member specifications, resolves accounts and groups, and adds
// them to the project while automatically excluding the active user.
//
// Parameters:
//   - projectId: The ID of the project to add members to
//   - projectMembers: Slice of member specification strings (format: "type=account,name=alice,access=editor")
//
// Returns:
//   - nil on success
//   - Error if member parsing, resolution, or addition fails
func (r *ProjectRunner) updateMembers(projectId string, projectMembers []string) error {
	logging.Trace()

	if len(projectMembers) == 0 {
		return nil // No members to update
	}

	activeUser, err := r.userSettings.Get()
	if err != nil {
		return fmt.Errorf("failed to get active user: %w", err)
	}

	members, err := r.buildProjectMembers(projectMembers, activeUser.Username, r.accounts, r.groups)
	if err != nil {
		return err
	}

	if len(members) > 0 {
		if err := r.resource.AddMembers(projectId, members); err != nil {
			return fmt.Errorf("failed to add members to project: %w", err)
		}
	}

	return nil
}

// normalizeProjectPath removes the leading slash from a project folder path.
// This ensures the path can be safely used with filepath.Join.
func normalizeProjectPath(folder string) string {
	return strings.TrimPrefix(folder, "/")
}

// makeFolder recursively creates the folder structure for a project on disk.
//
// This helper function is used during project export with the --expand flag to create
// the directory hierarchy that matches the project's folder structure. It creates each
// folder and recursively creates all child folders.
//
// Parameters:
//   - p: Parent directory path where the folder should be created
//   - f: ProjectFolder containing the folder name and children to create
//
// Side effects:
//   - Creates directories on disk if they don't already exist
func makeFolder(p string, f services.ProjectFolder) {
	path := filepath.Join(p, f.Name)
	if !utils.PathExists(path) {
		utils.EnsurePathExists(path)
	}
	for _, ele := range f.Children {
		makeFolder(path, ele)
	}
}

// resolveMember resolves a Member specification into a ProjectMember by
// looking up the account or group and populating all required fields.
func (r *ProjectRunner) resolveMember(
	member *Member,
	accounts *services.AccountService,
	groups *services.GroupService,
) (services.ProjectMember, error) {
	switch member.Type {
	case services.MemberTypeAccount:
		account, err := accounts.GetByName(member.Name)
		if err != nil {
			return services.ProjectMember{}, fmt.Errorf("account %q not found: %w", member.Name, err)
		}
		return services.ProjectMember{
			Provenance: account.Provenance,
			Reference:  account.Id,
			Role:       member.Access,
			Type:       services.MemberTypeAccount,
			Username:   account.Username,
		}, nil

	case services.MemberTypeGroup:
		group, err := groups.GetByName(member.Name)
		if err != nil {
			return services.ProjectMember{}, fmt.Errorf("group %q not found: %w", member.Name, err)
		}
		return services.ProjectMember{
			Provenance: group.Provenance,
			Reference:  group.Id,
			Role:       member.Access,
			Type:       services.MemberTypeGroup,
			Name:       group.Name,
		}, nil

	default:
		return services.ProjectMember{}, fmt.Errorf("invalid member type %q (must be 'account' or 'group')", member.Type)
	}
}

// buildProjectMembers converts member specifications into ProjectMember objects.
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
//   - Slice of ProjectMember objects ready for API submission
//   - Error if member parsing fails, member not found, or resolution fails
func (r *ProjectRunner) buildProjectMembers(
	memberSpecs []string,
	activeUsername string,
	accounts *services.AccountService,
	groups *services.GroupService,
) ([]services.ProjectMember, error) {
	logging.Trace()

	if len(memberSpecs) == 0 {
		return nil, nil
	}

	var members []services.ProjectMember

	for _, spec := range memberSpecs {
		member, err := parseMember(spec)
		if err != nil {
			return nil, fmt.Errorf("invalid member specification %q: %w", spec, err)
		}

		// Skip active user
		if member.Type == services.MemberTypeAccount && member.Name == activeUsername {
			logging.Info("skipping active user %q from member list", member.Name)
			continue
		}

		projectMember, err := r.resolveMember(member, accounts, groups)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve member %q: %w", member.Name, err)
		}

		members = append(members, projectMember)
	}

	return members, nil
}

// parseMember parses a member specification string into a Member struct.
// The format is: "type=<account|group>,name=<name>[,access=<role>]"
//
// Parameters:
//   - member: Member specification string
//
// Returns:
//   - Parsed Member with defaults applied
//   - Error if format is invalid or required fields are missing
//
// Example valid inputs:
//   - "type=account,name=alice"
//   - "type=account,name=alice,access=owner"
//   - "type=group,name=devops,access=editor"
func parseMember(member string) (*Member, error) {
	if member == "" {
		return nil, fmt.Errorf("member specification cannot be empty")
	}

	parts := strings.Split(member, ",")
	m := &Member{
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

// expandProject writes a project to disk in expanded format with separate files
// for each component.
//
// This function is used during export with the --expand flag. It creates a folder
// structure matching the project's hierarchy and writes each component to a separate
// JSON file. The main project file references these component files by filename instead
// of including the full component data.
//
// Parameters:
//   - in: Request object (unused, kept for consistency with command pattern)
//   - project: The project to expand to disk
//   - path: Base directory path where the expanded project will be written
//
// Returns:
//   - nil on success
//   - Error if directory creation, file writing, or JSON marshaling fails
//
// Side effects:
//   - Creates folder structure on disk matching project.Folders
//   - Writes multiple JSON files (one per component plus main project file)
//   - Removes "members", "accessControl", and "componentIidIndex" fields
//
// The resulting main project file references component documents by filename
// rather than embedding them, so it is only importable via `ipctl import`
// (which reconstructs the documents), not by the platform's manual import.
//
// TODO: This function should be moved to the resource layer (pkg/resources/projects.go)
// as it contains business logic about project structure and serialization format.
func expandProject(in Request, project *services.Project, path string) error {
	logging.Trace()

	for _, ele := range project.Folders {
		if ele.NodeType == "folder" {
			makeFolder(path, ele)
		}
	}

	var projectMap map[string]interface{}
	if err := utils.ToMap(project, &projectMap); err != nil {
		return err
	}

	// Remove server-managed fields, mirroring the standard export. Unlike the
	// standard export, the expanded layout is an ipctl-specific format: each
	// component's document is written to its own file and referenced by
	// filename, so this main project file is NOT directly importable by the
	// platform. It must be re-imported with `ipctl import`, which reconstructs
	// the documents from the component files.
	delete(projectMap, "members")
	delete(projectMap, "accessControl")
	delete(projectMap, "componentIidIndex")

	components := projectMap["components"].([]interface{})

	for idx, ele := range project.Components {
		p := path
		if ele.Folder != "/" {
			p = filepath.Join(path, normalizeProjectPath(ele.Folder))
		}

		docName := strings.Replace(ele.Document["name"].(string), "/", "_", -1)
		fn := fmt.Sprintf("%s.%s.json", docName, strings.ToLower(ele.Type))
		if err := utils.WriteJsonToDisk(ele.Document, fn, p); err != nil {
			return err
		}

		delete(components[idx].(map[string]interface{}), "document")
		components[idx].(map[string]interface{})["filename"] = fn
	}

	projectMap["components"] = components

	fn := fmt.Sprintf("%s.project.json", strings.Replace(project.Name, "/", "_", -1))

	return utils.WriteJsonToDisk(projectMap, fn, path)
}
