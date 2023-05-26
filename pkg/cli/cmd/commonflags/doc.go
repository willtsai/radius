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

// commongflags contains the common flags used by the Radius CLI. Commands should use these definitions
// to register flags with the Cobra CLI to ensure consistency across commands.
//
// When defining new commands or new flags for 'rad' follow the following guidance:
//
//   - Use an existing flag if possible, for consistency
//   - If a new flag is needed, define in it in the command where it is used
//   - If multiple commands will in the *same* group (eg: all 'rad recipes' commands) then define it a
//     'common' package inside the containing package (eg: pkg/cli/cmd/recipes/common/flags.go)
//   - If multiple commands from *different* groups will use the same flag, define it here
package commonflags
