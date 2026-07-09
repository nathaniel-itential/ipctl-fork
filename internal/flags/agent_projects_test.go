// Copyright 2024 Itential Inc. All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential

package flags

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestAgentProjectImportOptions_Flags(t *testing.T) {
	checkFlags(t, &AgentProjectImportOptions{}, []string{"member"})
}

func TestAgentProjectExportOptions_Flags(t *testing.T) {
	opts := &AgentProjectExportOptions{}
	cmd := &cobra.Command{}
	opts.Flags(cmd)
	assert.NotNil(t, opts)
}
