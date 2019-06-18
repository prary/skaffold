/*
Copyright 2019 The Skaffold Authors

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

package cmd

import (
	"context"
	"io"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/runner"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest"
	"github.com/spf13/cobra"
)

// NewCmdDelete describes the CLI command to delete deployed resources.
func NewCmdDelete(out io.Writer) *cobra.Command {
	return NewCmd(out, "delete").
		WithDescription("Delete the deployed resources").
		WithCommonFlags().
		NoArgs(cancelWithCtrlC(context.Background(), doDelete))
}

func doDelete(ctx context.Context, out io.Writer) error {
	return withRunner(ctx, func(r *runner.SkaffoldRunner, _ *latest.SkaffoldConfig) error {
		return r.Deployer.Cleanup(ctx, out)
	})
}
