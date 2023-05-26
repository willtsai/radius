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

// AddBicepParametersFlagVar defines the '--parameters'flag (Bicep deployment parameters) for a command and binds it to a variable.
func AddBicepParametersFlagVar(cmd *cobra.Command, ref *[]string) {
	cmd.Flags().StringArrayVarP(ref, "parameters", "p", []string{}, "Specify parameters for the deployment")
}
