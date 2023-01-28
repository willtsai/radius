// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package handlers

import (
	"testing"

	"github.com/project-radius/radius/pkg/linkrp"
	"github.com/stretchr/testify/require"
)

func Test_ParameterConflict(t *testing.T) {
	devParams := map[string]any{
		"throughput": 400,
		"port":       2030,
		"name":       "test-parameters",
	}
	operatorParams := map[string]any{
		"port":     2040,
		"name":     "test-parameters-conflict",
		"location": "us-east1",
	}
	expectedParams := map[string]any{
		"throughput": map[string]any{
			"value": 400,
		},
		"port": map[string]any{
			"value": 2030,
		},
		"name": map[string]any{
			"value": "test-parameters",
		},
		"location": map[string]any{
			"value": "us-east1",
		},
	}

	actualParams := createRecipeParameters(devParams, operatorParams, false, nil)
	require.Equal(t, expectedParams, actualParams)
}

func Test_ContextParameter(t *testing.T) {
	linkID := "/subscriptions/testSub/resourceGroups/testGroup/providers/applications.link/mongodatabases/mongo0"
	expectedLinkContext := linkrp.RecipeContext{
		Resource: linkrp.Resource{
			ResourceInfo: linkrp.ResourceInfo{
				ID:   "/subscriptions/testSub/resourceGroups/testGroup/providers/applications.link/mongodatabases/mongo0",
				Name: "mongo0",
			},
			Type: "applications.link/mongodatabases",
		},
		Application: linkrp.ResourceInfo{
			Name: "testApplication",
			ID:   "/subscriptions/test-sub/resourceGroups/test-group/providers/Applications.Core/applications/testApplication",
		},
		Environment: linkrp.ResourceInfo{
			Name: "env0",
			ID:   "/subscriptions/test-sub/resourceGroups/test-group/providers/Applications.Core/environments/env0",
		},
		Runtime: linkrp.Runtime{
			Kubernetes: linkrp.Kubernetes{
				Namespace:            "radius-test-app",
				EnvironmentNamespace: "radius-test-env",
			},
		},
	}

	linkContext, err := CreateRecipeContextParameter(linkID, "/subscriptions/test-sub/resourceGroups/test-group/providers/Applications.Core/environments/env0", "radius-test-env", "/subscriptions/test-sub/resourceGroups/test-group/providers/Applications.Core/applications/testApplication", "radius-test-app")
	require.NoError(t, err)
	require.Equal(t, expectedLinkContext, *linkContext)
}

func Test_DevParameterWithContextParameter(t *testing.T) {
	devParams := map[string]any{
		"throughput": 400,
		"port":       2030,
		"name":       "test-parameters",
	}
	recipeContext := linkrp.RecipeContext{
		Resource: linkrp.Resource{
			ResourceInfo: linkrp.ResourceInfo{
				ID:   "/subscriptions/testSub/resourceGroups/testGroup/providers/applications.link/mongodatabases/mongo0",
				Name: "mongo0",
			},
			Type: "Applications.Link/mongoDatabases",
		},
		Application: linkrp.ResourceInfo{
			ID:   "/subscriptions/test-sub/resourceGroups/test-group/providers/Applications.Core/applications/testApplication",
			Name: "testApplication",
		},
		Environment: linkrp.ResourceInfo{
			ID:   "/subscriptions/test-sub/resourceGroups/test-group/providers/Applications.Core/environments/env0",
			Name: "env0",
		},
		Runtime: linkrp.Runtime{
			Kubernetes: linkrp.Kubernetes{
				EnvironmentNamespace: "radius-test-env",
				Namespace:            "radius-test-app",
			},
		},
	}

	expectedParams := map[string]any{
		"throughput": map[string]any{
			"value": 400,
		},
		"port": map[string]any{
			"value": 2030,
		},
		"name": map[string]any{
			"value": "test-parameters",
		},
		"context": map[string]any{
			"value": recipeContext,
		},
	}
	actualParams := createRecipeParameters(devParams, nil, true, &recipeContext)
	require.Equal(t, expectedParams, actualParams)
}

func Test_ContextParameterError(t *testing.T) {
	envID := "error-env"
	linkContext, err := CreateRecipeContextParameter("/subscriptions/testSub/resourceGroups/testGroup/providers/applications.link/mongodatabases/mongo0", envID, "radius-test-env", "/subscriptions/test-sub/resourceGroups/test-group/providers/Applications.Core/applications/testApplication", "radius-test-app")
	require.Error(t, err)
	require.Nil(t, linkContext)
}

func Test_ACRPathParser(t *testing.T) {
	repository, tag, err := parseTemplatePath("radiusdev.azurecr.io/recipes/functionaltest/parameters/mongodatabases/azure:1.0")
	require.NoError(t, err)
	require.Equal(t, "radiusdev.azurecr.io/recipes/functionaltest/parameters/mongodatabases/azure", repository)
	require.Equal(t, "1.0", tag)
}

func Test_ACRPathParserErr(t *testing.T) {
	repository, tag, err := parseTemplatePath("http://user:passwd@example.com/test/bar:v1")
	require.Error(t, err)
	require.Equal(t, "", repository)
	require.Equal(t, "", tag)
}
