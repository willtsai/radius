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

package recipes

import (
	"fmt"

	v1 "github.com/project-radius/radius/pkg/armrpc/api/v1"
)

type RecipeError struct {
	v1.ErrorDetails
}

func (r *RecipeError) Error() string {
	return fmt.Sprintf("code %v: err %v", r.Code, r.Message)
}

func (e *RecipeError) Is(target error) bool {
	_, ok := target.(*RecipeError)
	return ok
}

func NewRecipeError(code string, message string, details *v1.ErrorDetails) *RecipeError {
	err := new(RecipeError)
	err.Message = message
	err.Code = code
	if details != nil {
		err.Details = []v1.ErrorDetails{*details}
	}
	return err
}
