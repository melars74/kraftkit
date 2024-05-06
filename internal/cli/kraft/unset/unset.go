// SPDX-License-Identifier: BSD-3-Clause
//
// Authors: Cezar Craciunoiu <cezar.craciunoiu@gmail.com>
//
// Copyright (c) 2022, Unikraft GmbH.  All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions
// are met:
//
// 1. Redistributions of source code must retain the above copyright
//    notice, this list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright
//    notice, this list of conditions and the following disclaimer in the
//    documentation and/or other materials provided with the distribution.
// 3. Neither the name of the copyright holder nor the names of its
//    contributors may be used to endorse or promote products derived from
//    this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

package unset

import (
	"context"
	"fmt"
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"kraftkit.sh/cmdfactory"
	"kraftkit.sh/log"
	"kraftkit.sh/packmanager"
	"kraftkit.sh/tui/confirm"
	"kraftkit.sh/tui/textinput"
	"kraftkit.sh/unikraft/app"
)

type UnsetOptions struct {
	Workdir string `long:"workdir" short:"w" usage:"Work on a unikernel at a path"`
}

// Unset a KConfig option in a Unikraft project.
func Unset(ctx context.Context, opts *UnsetOptions, args ...string) error {
	if opts == nil {
		opts = &UnsetOptions{}
	}

	return opts.Run(ctx, args)
}

func NewCmd() *cobra.Command {
	cmd, err := cmdfactory.New(&UnsetOptions{}, cobra.Command{
		Short:   "Unset a variable for a Unikraft project",
		Hidden:  true,
		Use:     "unset [OPTIONS] [param ...]",
		Aliases: []string{"u"},
		Long: heredoc.Doc(`
			unset a variable for a Unikraft project
		`),
		Example: heredoc.Doc(`
			# Unset variables in the cwd project
			$ kraft unset LIBDEVFS_DEV_STDOUT LWIP_TCP_SND_BUF

			# Unset variables in a project at a path
			$ kraft unset -w path/to/app LIBDEVFS_DEV_STDOUT LWIP_TCP_SND_BUF
		`),
		Annotations: map[string]string{
			cmdfactory.AnnotationHelpGroup:  "build",
			cmdfactory.AnnotationHelpHidden: "true",
		},
	})
	if err != nil {
		panic(err)
	}

	return cmd
}

func (*UnsetOptions) Pre(cmd *cobra.Command, _ []string) error {
	ctx, err := packmanager.WithDefaultUmbrellaManagerInContext(cmd.Context())
	if err != nil {
		return err
	}

	cmd.SetContext(ctx)

	return nil
}

func (opts *UnsetOptions) Run(ctx context.Context, args []string) error {
	var err error

	log.G(ctx).Warnf("This command is DEPRECATED and should not be used")

	workdir := ""
	confOpts := []string{}

	// Skip if nothing can be unset
	if len(args) == 0 {
		return fmt.Errorf("no options to unset")
	}

	// Set the working directory
	if opts.Workdir != "" {
		workdir = opts.Workdir
	} else {
		workdir, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	for _, arg := range args {
		confOpts = append(confOpts, arg+"=n")
	}

	// Check if dotconfig exists in workdir
	dotconfig := fmt.Sprintf("%s/.config", workdir)

	// Check if the file exists
	// TODO: offer option to start in interactive mode
	if _, err := os.Stat(dotconfig); os.IsNotExist(err) {
		imode, ierr := confirm.NewConfirm("Do you want to start in interactive mode:")
		if ierr != nil {
			return ierr
		}
		if imode {
			dotconfig, err = textinput.NewTextInput(
				"Path to dotconfig file:",
				"Enter path",
				"",
			)
			if err != nil {
				return err
			}
			if dotconfig == "" {
				return fmt.Errorf("dotconfig file does not exist: %s", dotconfig)
			}
		} else {
			return fmt.Errorf("dotconfig file does not exist: %s", dotconfig)
		}
	}

	// Initialize at least the configuration options for a project
	project, err := app.NewProjectFromOptions(
		ctx,
		app.WithProjectWorkdir(workdir),
		// app.WithProjectDefaultConfigPath(),
		app.WithProjectConfig(confOpts),
	)
	if err != nil {
		return err
	}

	return project.Unset(ctx, nil)
}
