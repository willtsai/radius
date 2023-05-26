/*
Copyright 2023 The Radius Authors.

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

package commonflags

import "github.com/spf13/cobra"

// AddContextFlagVar defines the '--context' flag for a command and binds it to a variable.
func AddContextFlagVar(cmd *cobra.Command, ref *string) {
	cmd.Flags().StringVarP(ref, "context", "c", "", "The Kubernetes context, will use the default if unset")
}

// AddNamespaceFlagVar defines the '--namespace' flag for a command and binds it to a variable.
func AddNamespaceFlagVar(cmd *cobra.Command, ref *string) {
	cmd.Flags().StringVarP(ref, "namespace", "n", "", "The Kubernetes namespace")
}
