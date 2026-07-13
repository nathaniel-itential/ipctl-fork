// Copyright 2024 Itential Inc. All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential

package handlers

import (
	"strings"

	"github.com/itential/ipctl/internal/config"
	"github.com/itential/ipctl/internal/terminal"
	"github.com/itential/ipctl/pkg/client"
	"github.com/spf13/cobra"
)

// RuntimeContext defines the interface for what handlers need from runtime.
// This enables dependency injection and makes handlers testable.
// By using config.Provider instead of *config.Config, handlers can work with
// any configuration implementation, making testing easier.
type RuntimeContext interface {
	GetClient() client.Client
	GetConfig() config.Provider
	GetDescriptors() Descriptors
	GetTerminalConfig() *terminal.Config
	IsVerbose() bool
}

// Runtime provides the execution context for handlers and commands.
// It implements RuntimeContext interface and holds configuration via the Provider interface,
// allowing for flexible configuration implementations.
type Runtime struct {
	client         client.Client
	config         config.Provider
	terminalConfig *terminal.Config
	descriptors    Descriptors
	// Verbose is exported to allow flag binding
	Verbose bool
}

// GetClient returns the HTTP client for API communication.
func (rt *Runtime) GetClient() client.Client {
	return rt.client
}

// GetConfig returns the application configuration provider.
// This returns the Provider interface, allowing handlers to access configuration
// without depending on the concrete Config type.
func (rt *Runtime) GetConfig() config.Provider {
	return rt.config
}

// GetTerminalConfig returns the terminal configuration.
func (rt *Runtime) GetTerminalConfig() *terminal.Config {
	return rt.terminalConfig
}

// GetDescriptors returns the command descriptors.
func (rt *Runtime) GetDescriptors() Descriptors {
	return rt.descriptors
}

// IsVerbose returns whether verbose output is enabled.
func (rt *Runtime) IsVerbose() bool {
	return rt.Verbose
}

// Handler coordinates command creation and manages the handler registry.
type Handler struct {
	runtime     *Runtime
	registry    *Registry
	descriptors Descriptors
}

// NewRuntime creates a new Runtime with the given client, configuration, and terminal config.
// Returns a pointer to enable sharing the runtime across handlers and allowing
// flag updates to be visible to all commands.
//
// The configuration is accepted as a Provider interface rather than *config.Config,
// which enables dependency injection and makes testing easier.
//
// Returns an error if descriptor loading fails, which can occur if:
// - The embedded descriptor directory cannot be read
// - A descriptor file is malformed or cannot be parsed
func NewRuntime(c client.Client, cfg config.Provider, termCfg *terminal.Config) (*Runtime, error) {
	descriptors, err := loadDescriptors()
	if err != nil {
		return nil, err
	}

	return &Runtime{
		client:         c,
		config:         cfg,
		terminalConfig: termCfg,
		descriptors:    descriptors,
	}, nil
}

// NewHandler creates a new Handler with the given runtime.
// It initializes all resource handlers and registers them in the handler registry.
func NewHandler(rt *Runtime) Handler {
	// Reuse descriptors from runtime instead of reloading
	descriptors := rt.descriptors

	// Create all handlers
	handlerInstances := []any{
		// Automation Studio handlers
		NewProjectHandler(rt, descriptors),
		NewAgentProjectHandler(rt, descriptors),
		NewWorkflowHandler(rt, descriptors),
		NewTransformationHandler(rt, descriptors),
		NewJsonFormHandler(rt, descriptors),
		NewCommandTemplateHandler(rt, descriptors),
		NewAnalyticTemplateHandler(rt, descriptors),
		NewTemplateHandler(rt, descriptors),

		// Operations Manager Handlers
		NewAutomationHandler(rt, descriptors),

		// Admin Essentials handlers
		NewAccountHandler(rt, descriptors),
		NewProfileHandler(rt, descriptors),
		NewRoleHandler(rt, descriptors),
		NewRoleTypesHandler(rt, descriptors),
		NewGroupHandler(rt, descriptors),
		NewMethodHandler(rt, descriptors),
		NewViewHandler(rt, descriptors),
		NewPrebuiltHandler(rt, descriptors),
		NewIntegrationModelHandler(rt, descriptors),
		NewIntegrationHandler(rt, descriptors),
		NewAdapterHandler(rt, descriptors),
		NewAdapterModelHandler(rt, descriptors),
		NewTagHandler(rt, descriptors),
		NewApplicationHandler(rt, descriptors),

		// Configuration Manager handlers
		NewDeviceHandler(rt, descriptors),
		NewDeviceGroupHandler(rt, descriptors),
		NewConfigurationParserHandler(rt, descriptors),
		NewGoldenConfigHandler(rt, descriptors),

		// Lifecycle Manager handlers
		NewModelHandler(rt, descriptors),

		// Flow Agent handlers
		NewAgentProjectHandler(rt, descriptors),

		NewServerHandler(rt, descriptors),
	}

	// Create instance-based registry
	registry := NewRegistry(handlerInstances)

	return Handler{
		runtime:     rt,
		registry:    registry,
		descriptors: descriptors,
	}
}

// GetCommands returns all 'get' commands from registered handlers.
func (h Handler) GetCommands() []*cobra.Command {
	var commands []*cobra.Command
	for _, ele := range h.registry.Readers() {
		cmd := ele.Get(h.runtime)
		if cmd != nil {
			commands = append(commands, cmd)
		}
	}
	return commands
}

// DescribeCommands returns all 'describe' commands from registered handlers.
func (h Handler) DescribeCommands() []*cobra.Command {
	var commands []*cobra.Command
	for _, ele := range h.registry.Readers() {
		cmd := ele.Describe(h.runtime)
		if cmd != nil {
			commands = append(commands, cmd)
		}
	}
	return commands
}

// AddCommandGroup adds a command group with a deterministic ID based on the title.
func (h Handler) AddCommandGroup(cmd *cobra.Command, title string, f func(Handler, string) []*cobra.Command) {
	// Use deterministic ID based on title instead of random UUID
	id := strings.ToLower(strings.ReplaceAll(title, " ", "-"))
	cmd.AddGroup(&cobra.Group{ID: id, Title: title})
	for _, ele := range f(h, id) {
		cmd.AddCommand(ele)
	}
}

// CreateCommands returns all 'create' commands from registered handlers.
func (h Handler) CreateCommands() []*cobra.Command {
	var commands []*cobra.Command
	for _, ele := range h.registry.Writers() {
		cmd := ele.Create(h.runtime)
		if cmd != nil {
			commands = append(commands, cmd)
		}
	}
	return commands
}

// DeleteCommands returns all 'delete' commands from registered handlers.
func (h Handler) DeleteCommands() []*cobra.Command {
	var commands []*cobra.Command
	for _, ele := range h.registry.Writers() {
		cmd := ele.Delete(h.runtime)
		if cmd != nil {
			commands = append(commands, cmd)
		}
	}
	return commands
}

// CopyCommands returns all 'copy' commands from registered handlers.
func (h Handler) CopyCommands() []*cobra.Command {
	var commands []*cobra.Command
	for _, ele := range h.registry.Copiers() {
		cmd := ele.Copy(h.runtime)
		if cmd != nil {
			commands = append(commands, cmd)
		}
	}
	return commands
}

// ClearCommands returns all 'clear' commands from registered handlers.
func (h Handler) ClearCommands() []*cobra.Command {
	var commands []*cobra.Command
	for _, ele := range h.registry.Writers() {
		cmd := ele.Clear(h.runtime)
		if cmd != nil {
			commands = append(commands, cmd)
		}
	}
	return commands
}

// ImportCommands returns all 'import' commands from registered handlers.
func (h Handler) ImportCommands() []*cobra.Command {
	var commands []*cobra.Command
	for _, ele := range h.registry.Importers() {
		cmd := ele.Import(h.runtime)
		if cmd != nil {
			commands = append(commands, cmd)
		}
	}
	return commands
}

// ExportCommands returns all 'export' commands from registered handlers.
func (h Handler) ExportCommands() []*cobra.Command {
	var commands []*cobra.Command
	for _, ele := range h.registry.Exporters() {
		cmd := ele.Export(h.runtime)
		if cmd != nil {
			commands = append(commands, cmd)
		}
	}
	return commands
}

// StartCommands returns all 'start' commands from registered handlers.
func (h Handler) StartCommands() []*cobra.Command {
	var commands []*cobra.Command
	for _, ele := range h.registry.Controllers() {
		cmd := ele.Start(h.runtime)
		if cmd != nil {
			commands = append(commands, cmd)
		}
	}
	return commands
}

// StopCommands returns all 'stop' commands from registered handlers.
func (h Handler) StopCommands() []*cobra.Command {
	var commands []*cobra.Command
	for _, ele := range h.registry.Controllers() {
		cmd := ele.Stop(h.runtime)
		if cmd != nil {
			commands = append(commands, cmd)
		}
	}
	return commands
}

// RestartCommands returns all 'restart' commands from registered handlers.
func (h Handler) RestartCommands() []*cobra.Command {
	var commands []*cobra.Command
	for _, ele := range h.registry.Controllers() {
		cmd := ele.Restart(h.runtime)
		if cmd != nil {
			commands = append(commands, cmd)
		}
	}
	return commands
}

// InspectCommands returns all 'inspect' commands from registered handlers.
func (h Handler) InspectCommands() []*cobra.Command {
	var commands []*cobra.Command
	for _, ele := range h.registry.Inspectors() {
		cmd := ele.Inspect(h.runtime)
		if cmd != nil {
			commands = append(commands, cmd)
		}
	}
	return commands
}

// EditCommands returns all 'edit' commands from registered handlers.
func (h Handler) EditCommands() []*cobra.Command {
	var commands []*cobra.Command
	for _, ele := range h.registry.Editors() {
		cmd := ele.Edit(h.runtime)
		if cmd != nil {
			commands = append(commands, cmd)
		}
	}
	return commands
}

// DumpCommands returns all 'dump' commands from registered handlers.
func (h Handler) DumpCommands() []*cobra.Command {
	var commands []*cobra.Command
	for _, ele := range h.registry.Dumpers() {
		cmd := ele.Dump(h.runtime)
		if cmd != nil {
			commands = append(commands, cmd)
		}
	}
	return commands
}

// LoadCommands returns all 'load' commands from registered handlers.
func (h Handler) LoadCommands() []*cobra.Command {
	var commands []*cobra.Command
	for _, ele := range h.registry.Loaders() {
		cmd := ele.Load(h.runtime)
		if cmd != nil {
			commands = append(commands, cmd)
		}
	}
	return commands
}
