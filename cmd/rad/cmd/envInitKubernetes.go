// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"
)

const (
	CORE_RP_API_VERSION = "2022-03-15-privatepreview"
)

var envInitKubernetesCmd = &cobra.Command{
	Use:   "kubernetes",
	Short: "Initializes a kubernetes environment",
	Long:  `Initializes a kubernetes environment.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return initSelfHosted(cmd, args, Kubernetes)
	},
}

func init() {
	envInitCmd.AddCommand(envInitKubernetesCmd)
	registerAzureProviderFlags(envInitKubernetesCmd)
	envInitKubernetesCmd.Flags().String("ucp-image", "", "Specify the UCP image to use")
	envInitKubernetesCmd.Flags().String("ucp-tag", "", "Specify the UCP tag to use")
}
