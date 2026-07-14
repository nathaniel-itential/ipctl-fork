// Copyright 2024 Itential Inc. All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential

package flags

import (
	"github.com/spf13/cobra"
)

// Command line options for `import project ...`
type ProjectImportOptions struct {
	Members                 []string
	SkipReferenceValidation bool
	ConflictMode            string
	AssignNewReferences     bool
}

func (o *ProjectImportOptions) Flags(cmd *cobra.Command) {
	cmd.Flags().StringArrayVar(&o.Members, "member", o.Members, "Configure one or more project members")
	cmd.Flags().BoolVar(&o.SkipReferenceValidation, "skip-reference-validation", false, "Skip reference validation on import")
	cmd.Flags().StringVar(&o.ConflictMode, "conflict-mode", "", `Conflict resolution strategy when a component already exists: "insert-new" or "overwrite"`)
	cmd.Flags().BoolVar(&o.AssignNewReferences, "assign-new-references", false, "Assign new references to imported components")
}

// Command line options for `export project ...`
type ProjectExportOptions struct {
	Expand bool
}

func (o *ProjectExportOptions) Flags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&o.Expand, "expand", o.Expand, "Expand the project assets")
}

// Command line options for `copy project ...`
type ProjectCopyOptions struct {
	Members []string
}

func (o *ProjectCopyOptions) Flags(cmd *cobra.Command) {
	cmd.Flags().StringArrayVar(&o.Members, "member", o.Members, "Configure one or more project members")
}
