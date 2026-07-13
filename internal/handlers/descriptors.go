// Copyright 2024 Itential Inc. All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential

package handlers

import (
	"embed"
	"fmt"
	"strings"

	"github.com/itential/ipctl/internal/cmdutils"
)

const descriptorsDir = "descriptors"

const (
	apiDescriptor = "api"

	accountsDescriptor     = "accounts"
	groupsDescriptor       = "groups"
	modelsDescriptor       = "models"
	rolesDescriptor        = "roles"
	roleTypesDescriptor    = "roletypes"
	adaptersDescriptor     = "adapters"
	methodsDescriptor      = "methods"
	viewsDescriptor        = "views"
	bundleDescriptor       = "bundle"
	prebuiltsDescriptor    = "prebuilts"
	profilesDescriptor     = "profiles"
	tagsDescriptor         = "tags"
	integrationModels      = "integration_models"
	integrations           = "integrations"
	adapterModels          = "adapter_models"
	applicationsDescriptor = "applications"

	automationsDescriptor = "automations"

	commandTemplatesDescriptor  = "command_templates"
	workflowsDescriptor         = "workflows"
	transformationsDescriptor   = "transformations"
	jsonformsDescriptor         = "jsonforms"
	projectsDescriptor          = "projects"
	agentProjectsDescriptor     = "agent_projects"
	analyticTemplatesDescriptor = "analytic_templates"
	templatesDescriptor         = "templates"

	devicesDescriptor              = "devices"
	deviceGroupsDescriptor         = "devicegroups"
	configurationParsersDescriptor = "configuration_parsers"
	gctreesDescriptor              = "gctrees"

	agentProjectsDescriptor = "agent_projects"

	serverDescriptor = "server"

	localAAADescriptor    = "localaaa"
	localClientDescriptor = "localclient"
)

//go:embed descriptors/*.yaml
var content embed.FS

type DescriptorMap map[string]cmdutils.Descriptor
type Descriptors map[string]DescriptorMap

func loadDescriptors() (Descriptors, error) {
	descriptors := map[string]DescriptorMap{}

	entries, err := content.ReadDir(descriptorsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read descriptors directory: %w", err)
	}

	for _, ele := range entries {
		name := strings.Split(ele.Name(), ".")[0]
		fn := strings.Join([]string{descriptorsDir, ele.Name()}, "/")

		data, err := content.ReadFile(fn)
		if err != nil {
			return nil, fmt.Errorf("failed to read descriptor %s: %w", name, err)
		}

		descriptors[name] = cmdutils.LoadDescriptor(data)
	}

	return descriptors, nil
}
