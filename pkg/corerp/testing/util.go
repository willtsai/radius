// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package testing

import (
	"os"
	"path"
)

func ReadPackageFixture(directory string, filename string) []byte {
	raw, err := os.ReadFile(path.Join(directory, "testdata", filename))
	if err != nil {
		return nil
	}

	return raw
}

func ReadFixture(filename string) []byte {
	raw, err := os.ReadFile("./testdata/" + filename)
	if err != nil {
		return nil
	}

	return raw
}
