/*
Copyright 2023.

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

package reconciler

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
)

type DeploymentState string

const (
	DeploymentStateWaiting  DeploymentState = "Waiting"
	DeploymentStateUpdating DeploymentState = "Updating"
	DeploymentStateReady    DeploymentState = "Ready"
	DeploymentStateDeleting DeploymentState = "Deleting"
	DeploymentStateDeleted  DeploymentState = "Deleted"
)

type deploymentAnnotations struct {
	Status        *deploymentStatusAnnotation
	Hash          string
	Configuration deploymentConfiguration
}

type deploymentConfiguration struct {
	Connections map[string]string
}

type deploymentStatusAnnotation struct {
	Scope       string          `json:"scope,omitempty"`
	Application string          `json:"application,omitempty"`
	Environment string          `json:"environment,omitempty"`
	Container   string          `json:"container,omitempty"`
	Operation   string          `json:"operation,omitempty"`
	State       DeploymentState `json:"state,omitempty"`
}

func readAnnotations(deployment *appsv1.Deployment) (*deploymentAnnotations, error) {
	if deployment.Annotations == nil {
		return nil, nil
	}

	enabled := deployment.Annotations[AnnotationRadiusEnabled]
	if !strings.EqualFold(enabled, "true") {
		return nil, nil
	}

	result := deploymentAnnotations{
		Configuration: deploymentConfiguration{Connections: map[string]string{}},
		Hash:          deployment.Annotations[AnnotationRadiusConfigurationHash],
	}

	for k, v := range deployment.Annotations {
		if strings.HasPrefix(k, AnnotationRadiusConnectionPrefix) {
			result.Configuration.Connections[strings.TrimPrefix(k, AnnotationRadiusConnectionPrefix)] = v
		}
	}

	status := deployment.Annotations[AnnotationRadiusStatus]
	if status == "" {
		return &result, nil
	}
	s := deploymentStatusAnnotation{}
	err := json.Unmarshal([]byte(status), &s)
	if err != nil {
		return &result, fmt.Errorf("failed to unmarshal status annotation: %w", err)
	}

	result.Status = &s

	return &result, nil
}

func (annotations *deploymentAnnotations) SetDefaults(namespace string, name string) {
	if annotations.Status == nil {
		scope := "/planes/radius/local/resourceGroups/default"
		annotations.Status = &deploymentStatusAnnotation{
			Scope:       scope,
			Environment: scope + "/providers/Applications.Core/environments/" + "default",
			Application: scope + "/providers/Applications.Core/applications/" + namespace,
			Container:   scope + "/providers/Applications.Core/containers/" + name,
		}
	}
}

func (annotations *deploymentAnnotations) ApplyToDeployment(deployment *appsv1.Deployment) error {
	if deployment.Annotations == nil {
		deployment.Annotations = map[string]string{}
	}

	status := ""
	if annotations.Status != nil {
		b, err := json.Marshal(annotations.Status)
		if err != nil {
			return err
		}

		status = string(b)
	}

	b, err := json.Marshal(&annotations.Configuration)
	if err != nil {
		return err
	}

	sum := sha1.Sum(b)
	hash := hex.EncodeToString(sum[:])

	deployment.Annotations[AnnotationRadiusEnabled] = "true"
	deployment.Annotations[AnnotationRadiusStatus] = status
	deployment.Annotations[AnnotationRadiusConfigurationHash] = hash

	for k, v := range annotations.Configuration.Connections {
		deployment.Annotations[AnnotationRadiusConnectionPrefix+k] = v
	}

	return nil
}

func (annotations *deploymentAnnotations) IsUpToDate() bool {
	if annotations.Hash == "" {
		return false
	}

	if annotations.Status == nil {
		return false
	}

	if annotations.Status.State != DeploymentStateReady {
		return false
	}

	b, err := json.Marshal(&annotations.Configuration)
	if err != nil {
		return false
	}

	sum := sha1.Sum(b)
	hash := hex.EncodeToString(sum[:])
	return hash == annotations.Hash
}
