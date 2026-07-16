// Copyright 2024 Itential Inc. All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential

package flags

import "github.com/spf13/cobra"

// AgentProjectCreateOptions holds options for the `create agent-project` command.
type AgentProjectCreateOptions struct {
	Description string
}

func (o *AgentProjectCreateOptions) Flags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.Description, "description", o.Description, "Description for the agent project (optional)")
}

// AgentProjectImportOptions holds options for the `import agent-project` command.
type AgentProjectImportOptions struct {
	Members      []string
	ConflictMode string
}

func (o *AgentProjectImportOptions) Flags(cmd *cobra.Command) {
	cmd.Flags().StringArrayVar(&o.Members, "member", o.Members, "Configure one or more project members")
	// Left unset by default (empty string) so the runner can distinguish "not specified" from
	// an explicit choice: --replace implies conflict-mode=replace unless --conflict-mode overrides it.
	cmd.Flags().StringVar(&o.ConflictMode, "conflict-mode", "", "How to handle a collision with an existing project (keep-both or replace); defaults to \"replace\" if --replace is set, otherwise \"keep-both\"")
}

// AgentProjectExportOptions holds options for the `export agent-project` command.
type AgentProjectExportOptions struct{}

func (o *AgentProjectExportOptions) Flags(_ *cobra.Command) {}

// AgentProjectCopyOptions holds options for the `copy agent-project` command.
type AgentProjectCopyOptions struct {
	Members []string
}

func (o *AgentProjectCopyOptions) Flags(cmd *cobra.Command) {
	cmd.Flags().StringArrayVar(&o.Members, "member", o.Members, "Configure one or more agent project members")
}
