// Copyright 2024 Itential Inc. All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential

package handlers

import (
	"github.com/itential/ipctl/internal/flags"
	"github.com/itential/ipctl/internal/runners"
)

func NewAgentProjectHandler(rt *Runtime, desc Descriptors) AssetHandler {
	return NewAssetHandler(
		runners.NewAgentProjectRunner(rt.GetClient(), rt.GetConfig()),
		desc[agentProjectsDescriptor],
		&AssetHandlerFlags{
			Import: &flags.AgentProjectImportOptions{},
			Export: &flags.AgentProjectExportOptions{},
		},
	)
}
