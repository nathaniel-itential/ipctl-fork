// Copyright 2024 Itential Inc. All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential

package handlers

import (
	"fmt"
	"strings"

	"github.com/itential/ipctl/internal/flags"
	"github.com/itential/ipctl/internal/output"
	"github.com/itential/ipctl/internal/runners"
	"github.com/spf13/cobra"
)

type RunCommand func(*cobra.Command, []string)

type CommandOptions func(*cobra.Command)

type CommandRunnerOption func(*CommandRunner)

type CommandRunner struct {
	Key         string
	Descriptors DescriptorMap
	Run         runners.RunnerFunc
	Common      flags.Flagger
	Options     flags.Flagger
	Runtime     *Runtime
	Runner      runners.Runner
	Flags       *AssetHandlerFlags
}

func NewCommandRunner(
	key string,
	desc DescriptorMap,
	run runners.RunnerFunc,
	runtime *Runtime,
	options flags.Flagger,
	opts ...CommandRunnerOption,
) *CommandRunner {

	cr := &CommandRunner{
		Key:         key,
		Descriptors: desc,
		Run:         run,
		Runtime:     runtime,
		Common:      options,
	}

	for _, opt := range opts {
		opt(cr)
	}

	return cr
}

func withOptions(f *AssetHandlerFlags) CommandRunnerOption {
	return func(c *CommandRunner) {
		switch c.Key {
		case "create":
			c.Options = f.Create
		case "delete":
			c.Options = f.Delete
		case "get":
			c.Options = f.Get
		case "describe":
			c.Options = f.Describe
		case "copy":
			c.Options = f.Copy
		case "clear":
			c.Options = f.Clear
		case "import":
			c.Options = f.Import
		case "export":
			c.Options = f.Export
		case "load":
			c.Options = f.Load
		case "dump":
			c.Options = f.Dump
		}
	}
}

func NewCommand(c *CommandRunner) *cobra.Command {
	desc, exists := c.Descriptors[c.Key]

	if !exists || desc.Disabled {
		return nil
	}

	var example string

	if desc.Example != "" {
		var lines []string
		for _, ele := range strings.Split(desc.Example, "\n") {
			lines = append(lines, fmt.Sprintf("  %s", ele))
		}
		example = strings.Join(lines, "\n")
	}

	cmd := &cobra.Command{
		Use:     desc.Use,
		GroupID: desc.Group,

		Short: desc.Short(),
		Long:  desc.Description,

		Example: example,

		Hidden: desc.Hidden,

		RunE: func(cmd *cobra.Command, args []string) error {

			req := runners.Request{
				Args:    args,
				Options: c.Options,
				Common:  c.Common,
				Runner:  c.Runner,
				Config:  c.Runtime.GetConfig(),
				Verbose: c.Runtime.IsVerbose(),
			}

			resp, err := c.Run(req)
			if err != nil {
				return err
			}

			// Create renderer based on configured output format
			termCfg := c.Runtime.GetTerminalConfig()
			renderer, err := output.NewRenderer(termCfg.DefaultOutput, termCfg.Pager)
			if err != nil {
				return err
			}

			// Render the response
			return renderer.Render(resp)
		},
	}

	if desc.ExactArgs > 0 {
		cmd.Args = cobra.ExactArgs(desc.ExactArgs)
	}

	return cmd
}
