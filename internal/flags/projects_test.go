// Copyright 2024 Itential Inc. All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential

package flags

import "testing"

func TestProjectImportOptions(t *testing.T) {
	checkFlags(t, &ProjectImportOptions{}, []string{
		"member",
		"skip-reference-validation",
		"conflict-mode",
		"assign-new-references",
	})
}

func TestProjectExportOptions(t *testing.T) {
	checkFlags(t, &ProjectExportOptions{}, []string{"expand"})
}

func TestProjectCopyOptions(t *testing.T) {
	checkFlags(t, &ProjectCopyOptions{}, []string{"member"})
}
