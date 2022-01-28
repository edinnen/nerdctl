/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newComposeExecExecCommand() *cobra.Command {
	var execCommand = &cobra.Command{
		Use:               "exec [OPTIONS] CONTAINER COMMAND [ARG...]",
		Args:              cobra.MinimumNArgs(2),
		Short:             "Run a command in a running container",
		RunE:              composeExecAction,
		ValidArgsFunction: execShellComplete,
		SilenceUsage:      true,
		SilenceErrors:     true,
	}
	execCommand.Flags().SetInterspersed(false)

	execCommand.Flags().BoolP("tty", "t", false, "(Currently -t needs to correspond to -i)")
	execCommand.Flags().BoolP("interactive", "i", false, "Keep STDIN open even if not attached")
	execCommand.Flags().BoolP("detach", "d", false, "Detached mode: run command in the background")
	execCommand.Flags().StringP("workdir", "w", "", "Working directory inside the container")
	// env needs to be StringArray, not StringSlice, to prevent "FOO=foo1,foo2" from being split to {"FOO=foo1", "foo2"}
	execCommand.Flags().StringArrayP("env", "e", nil, "Set environment variables")
	// env-file is defined as StringSlice, not StringArray, to allow specifying "--env-file=FILE1,FILE2" (compatible with Podman)
	execCommand.Flags().StringSlice("env-file", nil, "Set environment variables from file")
	execCommand.Flags().Bool("privileged", false, "Give extended privileges to the command")
	execCommand.Flags().StringP("user", "u", "", "Username or UID (format: <name|uid>[:<group|gid>])")
	return execCommand
}

func composeExecAction(cmd *cobra.Command, args []string) error {
	// simulate the behavior of double dash
	newArg := []string{}
	if len(args) >= 2 && args[1] == "--" {
		newArg = append(newArg, args[:1]...)
		newArg = append(newArg, args[2:]...)
		args = newArg
	}

	client, ctx, cancel, err := newClient(cmd)
	if err != nil {
		return err
	}
	defer cancel()

	c, err := getComposer(cmd, client)
	if err != nil {
		return err
	}

	serviceNames, err := c.ServiceNames(args[0])

	containers, err := c.Containers(ctx, serviceNames...)
	if err != nil {
		return err
	}

	if len(containers) == 0 {
		return fmt.Errorf("no containers found for %s", serviceNames)
	}

	if len(containers) > 1 {
		return fmt.Errorf("multiple containers found for %s", serviceNames)
	}

	container := containers[0]
	return execActionWithContainer(ctx, cmd, args, container, client)
}
