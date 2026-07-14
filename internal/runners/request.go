// Copyright 2024 Itential Inc. All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential

package runners

import "github.com/itential/ipctl/internal/config"

// Request encapsulates the input parameters for a runner operation.
// It provides access to command arguments, flags, and configuration.
type Request struct {
	Args    []string
	Common  any
	Options any
	Runner  Runner
	// Config provides access to application configuration via the Provider interface.
	// This allows runners to access profiles, features, and other configuration
	// without depending on the concrete Config type.
	Config  config.Provider
	Verbose bool
}
