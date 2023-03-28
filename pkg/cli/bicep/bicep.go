// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package bicep

import (
	"fmt"
	"net/http"
	"os"

	"github.com/project-radius/radius/pkg/cli/tools"
)

const radBicepEnvVar = "RAD_BICEP"
const binaryName = "rad-bicep"

// IsBicepInstalled returns true if our local copy of bicep is installed
func IsBicepInstalled() (bool, error) {
	filepath, err := tools.GetLocalFilepath(radBicepEnvVar, binaryName)
	if err != nil {
		return false, err
	}

	_, err = os.Stat(filepath)
	if err != nil && os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("error checking for %s: %v", filepath, err)
	}

	return true, nil
}

// DeleteBicep cleans our local copy of bicep
func DeleteBicep() error {
	filepath, err := tools.GetLocalFilepath(radBicepEnvVar, binaryName)
	if err != nil {
		return err
	}

	err = os.Remove(filepath)
	if err != nil {
		return fmt.Errorf("failed to delete %s: %v", filepath, err)
	}

	return nil
}

// DownloadBicep updates our local copy of bicep
func DownloadBicep() error {
	dirPrefix := "bicep-extensibility"
	// Placeholders are for: channel, platform, filename
	downloadURIFmt := fmt.Sprint("https://get.radapp.dev/tools/", dirPrefix, "/%s/%s/%s")

	uri, err := tools.GetDownloadURI(downloadURIFmt, binaryName)
	if err != nil {
		return err
	}

	resp, err := http.Get(uri)
	if err != nil {
		return fmt.Errorf("failed to download bicep: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("failed to download bicep from '%s'with status code: %d", uri, resp.StatusCode)
	}

	filepath, err := tools.GetLocalFilepath(radBicepEnvVar, binaryName)
	if err != nil {
		return err
	}

	return tools.DownloadToFolder(filepath, resp)
}