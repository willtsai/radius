// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package azure

import (
	"context"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	v1 "github.com/project-radius/radius/pkg/armrpc/api/v1"
	"github.com/project-radius/radius/pkg/corerp/datamodel"
	"github.com/project-radius/radius/pkg/corerp/handlers"
	"github.com/project-radius/radius/pkg/corerp/renderers"
	"github.com/project-radius/radius/pkg/kubernetes"
	"github.com/project-radius/radius/pkg/rp"
	"github.com/project-radius/radius/pkg/rp/outputresource"
	"github.com/stretchr/testify/require"

	corev1 "k8s.io/api/core/v1"
	csiv1 "sigs.k8s.io/secrets-store-csi-driver/apis/v1"
)

func TestMakeKeyVaultVolumeSpec(t *testing.T) {
	v, vm, err := MakeKeyVaultVolumeSpec("kvvol", "/tmp", "azkv")

	require.NoError(t, err)
	require.Equal(t, corev1.Volume{
		Name: "kvvol",
		VolumeSource: corev1.VolumeSource{
			CSI: &corev1.CSIVolumeSource{
				Driver:   "secrets-store.csi.k8s.io",
				ReadOnly: to.Ptr(true),
				VolumeAttributes: map[string]string{
					"secretProviderClass": "azkv",
				},
			},
		},
	}, v)

	require.Equal(t, corev1.VolumeMount{
		Name:      "kvvol",
		MountPath: "/tmp",
		ReadOnly:  true,
	}, vm)
}

func TestMakeKeyVaultSecretProviderClass(t *testing.T) {
	envOpt := &renderers.EnvironmentOptions{
		Namespace: "default",
		Identity: &rp.IdentitySettings{
			Kind:       rp.AzureIdentityWorkload,
			OIDCIssuer: "https://radiusoidc/00000000-0000-0000-0000-000000000000",
		},
	}

	spcTests := []struct {
		desc         string
		identityKind rp.IdentitySettingKind

		err          error
		beforeParams map[string]string
		afterParams  map[string]string
	}{
		{
			desc:         "azure.com.workload",
			identityKind: rp.AzureIdentityWorkload,
			err:          nil,
			beforeParams: map[string]string{
				"usePodIdentity": "false",
				"keyvaultName":   "vault0",
				"objects":        "params",
			},
			afterParams: map[string]string{
				"usePodIdentity": "false",
				"keyvaultName":   "vault0",
				"objects":        "params",
				"clientID":       "newClientID",
				"tenantID":       "newTenantID",
			},
		},
		{
			desc:         "azure.com.unknown",
			identityKind: "",
			err:          errUnsupportedIdentityKind,
		},
	}

	vol := &datamodel.VolumeResource{
		BaseResource: v1.BaseResource{
			TrackedResource: v1.TrackedResource{
				Name: "test-cntr",
				Type: "applications.core/volumes",
			},
		},
		Properties: datamodel.VolumeResourceProperties{
			Kind: datamodel.AzureKeyVaultVolume,
			AzureKeyVault: &datamodel.AzureKeyVaultVolumeProperties{
				Resource: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testGroup/providers/Microsoft.KeyVault/vaults/vault0",
			},
		},
	}

	for _, tc := range spcTests {
		t.Run(tc.desc, func(t *testing.T) {
			envOpt.Identity.Kind = tc.identityKind
			or, err := MakeKeyVaultSecretProviderClass("app", "spcName", vol, "params", envOpt)
			if tc.err != nil {
				require.ErrorIs(t, tc.err, err)
			} else {
				r := or.Resource.(*csiv1.SecretProviderClass)
				require.Equal(t, string(tc.identityKind), r.Annotations[kubernetes.AnnotationIdentityType])
				require.Equal(t, tc.beforeParams, r.Spec.Parameters)

				// Transform
				putOptions := &handlers.PutOptions{
					Resource: or,
					DependencyProperties: map[string]map[string]string{
						// output properties of managed identity
						outputresource.LocalIDUserAssignedManagedIdentity: {
							handlers.UserAssignedIdentityClientIDKey: "newClientID",
							handlers.UserAssignedIdentityTenantIDKey: "newTenantID",
						},
					},
				}
				err := TransformSecretProviderClass(context.Background(), putOptions)
				require.NoError(t, err)
				require.Equal(t, tc.afterParams, r.Spec.Parameters)
			}
		})
	}
}
