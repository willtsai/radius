// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package bicep

import (
	"testing"
	"testing/fstest"

	"github.com/project-radius/radius/pkg/cli/clients"
	"github.com/stretchr/testify/require"
)

func Test_Parameters_Invalid(t *testing.T) {
	inputs := []string{
		"foo",
		"foo.json",
		"foo bar.json",
		"foo bar",
	}

	parser := ParameterParser{
		FileSystem: fstest.MapFS{},
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			parameters, err := parser.Parse(input)
			require.Error(t, err)
			require.Nil(t, parameters)
		})
	}
}

func Test_ParseParameters_Overwrite(t *testing.T) {
	inputs := []string{
		"@many.json",
		"key2=@single.json",
		"key3=value2",
		"key3=value3",
	}

	parser := ParameterParser{
		FileSystem: fstest.MapFS{
			"many.json": {
				Data: []byte(`{ "parameters": { "key1": { "value": { "someValue": true } }, "key2": { "value": "overridden-value" } } }`),
			},
			"single.json": {
				Data: []byte(`{ "someValue": "another-value" }`),
			},
		},
	}

	parameters, err := parser.Parse(inputs...)
	require.NoError(t, err)

	expected := clients.DeploymentParameters{
		"key1": map[string]interface{}{
			"value": map[string]interface{}{
				"someValue": true,
			},
		},
		"key2": map[string]interface{}{
			"value": map[string]interface{}{
				"someValue": "another-value",
			},
		},
		"key3": map[string]interface{}{
			"value": "value3",
		},
	}

	require.Equal(t, expected, parameters)
}

func Test_ParseParameters_File(t *testing.T) {
	input := []byte(`
	{
		"$schema": "https://schema.management.azure.com/schemas/2019-04-01/deploymentParameters.json#",
		"contentVersion": "1.0.0.0",
		"parameters": {
		  "param1": {
			"value": "foo"
		  },
		  "param2": {
			"value": "bar"
		  }
		}
	  }
	`)

	parser := ParameterParser{
		FileSystem: fstest.MapFS{},
	}

	parameters, err := parser.ParseFileContents(input)
	require.NoError(t, err)

	expected := clients.DeploymentParameters{
		"param1": map[string]interface{}{
			"value": "foo",
		},
		"param2": map[string]interface{}{
			"value": "bar",
		},
	}

	require.Equal(t, expected, parameters)
}
