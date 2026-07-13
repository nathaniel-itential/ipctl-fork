// Copyright 2024 Itential Inc. All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential

package cli

import (
	"fmt"
	"strings"

	"github.com/itential/ipctl/internal/cmdutils"
	"github.com/itential/ipctl/internal/handlers"
	"github.com/itential/ipctl/internal/logging"
	"github.com/spf13/cobra"
)

// RootCommand defines the configuration for a top-level CLI command group.
// Each RootCommand generates a set of Cobra commands organized under a
// common parent command (e.g., "get", "create", "delete").
//
// The Run function is invoked to generate the actual Cobra commands, which
// are then organized using the descriptor from the YAML configuration files.
type RootCommand struct {
	Name       string                  // Name of the command (e.g., "get", "create")
	Group      string                  // Group ID for organizing in help output
	Run        func() []*cobra.Command // Function to generate child commands
	Descriptor string                  // Key for YAML descriptor lookup
}

// addRootCommand adds a new top level command to the application. Root
// commands typically do not implement functionality, rather provide a command
// tree for more specific commands.
func addRootCommand(cmd *cobra.Command, rt *handlers.Runtime, title string, f func(*handlers.Runtime, string) []*cobra.Command) {
	// Use deterministic ID based on title instead of random UUID
	id := strings.ToLower(strings.ReplaceAll(title, " ", "-"))
	children := f(rt, id)
	if len(children) > 0 {
		cmd.AddGroup(&cobra.Group{ID: id, Title: title})
		for _, ele := range children {
			cmd.AddCommand(ele)
		}
	}
}

// makeRootCommand will create a new root command for the application. Root
// commands are top level commands that implement additional subcommands and
// therefore do not directly perform any actions. The function accepts a
// single argument `rootCommands` which is an array of RootCommand instances.
// Returns an error if any required descriptor is missing.
func makeRootCommand(rootCommands []RootCommand) ([]*cobra.Command, error) {
	descriptors := cmdutils.LoadDescriptorsFromContent("descriptors", &descriptorFiles)

	commands := make([]*cobra.Command, 0, len(rootCommands))

	for _, ele := range rootCommands {
		desc, exists := descriptors[ele.Descriptor]
		if !exists {
			return nil, fmt.Errorf("failed to build %s command: missing descriptor '%s'", ele.Name, ele.Descriptor)
		}

		c := makeChildCommand(ele, desc)
		if c != nil {
			commands = append(commands, c)
		}
	}

	return commands, nil
}

// makeChildCommand creates a single command attached to a root command.  Child
// commands are typically handed off to a handler for further implementation of
// the command action.
func makeChildCommand(root RootCommand, desc map[string]cmdutils.Descriptor) *cobra.Command {
	var cmd = &cobra.Command{
		Use:     root.Name,
		GroupID: root.Group,

		Short: strings.Split(desc[root.Name].Description, "\n")[0],
		Long:  desc[root.Name].Description,

		Example: desc[root.Name].Example,

		Hidden: desc[root.Name].Hidden,
	}

	if desc[root.Name].IncludeGroups {
		cmd.AddGroup(
			&cobra.Group{ID: "admin-essentials", Title: "Admin Essentials Commands:"},
			&cobra.Group{ID: "automation-studio", Title: "Automation Studio Commands:"},
			&cobra.Group{ID: "agent-projects", Title: "Agent Project Commands:"},
			&cobra.Group{ID: "configuration-manager", Title: "Configuration Manager Commands:"},
			&cobra.Group{ID: "operations-manager", Title: "Operations Manager Commands:"},
			&cobra.Group{ID: "lifecycle-manager", Title: "Lifecycle Manager Commands:"},
			&cobra.Group{ID: "flow-agent", Title: "Flow Agent Commands:"},
		)
	}

	children := root.Run()

	if len(children) == 0 {
		return nil
	}

	cmd.AddCommand(children...)

	return cmd
}

// assetCommands define the aggregate set of commands for working with assets.
func assetCommands(rt *handlers.Runtime, id string) []*cobra.Command {
	h := handlers.NewHandler(rt)
	commands, err := makeRootCommand([]RootCommand{
		{Name: "get", Group: id, Run: h.GetCommands, Descriptor: "asset"},
		{Name: "describe", Group: id, Run: h.DescribeCommands, Descriptor: "asset"},
		{Name: "create", Group: id, Run: h.CreateCommands, Descriptor: "asset"},
		{Name: "delete", Group: id, Run: h.DeleteCommands, Descriptor: "asset"},
		{Name: "copy", Group: id, Run: h.CopyCommands, Descriptor: "asset"},
		{Name: "clear", Group: id, Run: h.ClearCommands, Descriptor: "asset"},
		{Name: "edit", Group: id, Run: h.EditCommands, Descriptor: "asset"},
		{Name: "import", Group: id, Run: h.ImportCommands, Descriptor: "asset"},
		{Name: "export", Group: id, Run: h.ExportCommands, Descriptor: "asset"},
	})
	if err != nil {
		logging.Error(err, "failed to create asset commands")
		return nil
	}
	return commands
}

// platformCommands define the set of commands that can be performed on a
// specific server instance.
func platformCommands(rt *handlers.Runtime, id string) []*cobra.Command {
	apiHandler := handlers.NewApiHandler(rt)
	h := handlers.NewHandler(rt)

	commands, err := makeRootCommand([]RootCommand{
		{Name: "api", Group: id, Run: apiHandler.Commands, Descriptor: "platform"},
		{Name: "inspect", Group: id, Run: h.InspectCommands, Descriptor: "platform"},
		{Name: "start", Group: id, Run: h.StartCommands, Descriptor: "platform"},
		{Name: "stop", Group: id, Run: h.StopCommands, Descriptor: "platform"},
		{Name: "restart", Group: id, Run: h.RestartCommands, Descriptor: "platform"},
	})
	if err != nil {
		logging.Error(err, "failed to create platform commands")
		return nil
	}
	return commands
}

// datasetCommands provide a set of commands for performing batch operations on
// specific asset types.
func datasetCommands(rt *handlers.Runtime, id string) []*cobra.Command {
	h := handlers.NewHandler(rt)
	commands, err := makeRootCommand([]RootCommand{
		{Name: "load", Group: id, Run: h.LoadCommands, Descriptor: "dataset"},
		{Name: "dump", Group: id, Run: h.DumpCommands, Descriptor: "dataset"},
	})
	if err != nil {
		logging.Error(err, "failed to create dataset commands")
		return nil
	}
	return commands
}

// pluginCommands are commands that extend the functionality of the
// application.
func pluginCommands(rt *handlers.Runtime, id string) []*cobra.Command {
	localAAAHandler := handlers.NewLocalAAAHandler(rt)
	localClientHandler := handlers.NewLocalClientHandler(rt)

	commands, err := makeRootCommand([]RootCommand{
		{Name: "local-aaa", Group: id, Run: localAAAHandler.Commands, Descriptor: "localaaa"},
		{Name: "client", Group: id, Run: localClientHandler.Commands, Descriptor: "localclient"},
	})
	if err != nil {
		logging.Error(err, "failed to create plugin commands")
		return nil
	}
	return commands
}
