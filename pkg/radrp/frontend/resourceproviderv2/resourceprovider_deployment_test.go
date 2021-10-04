// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package resourceproviderv2

import (
	"context"
	"testing"

	"github.com/Azure/radius/pkg/model/components"
	"github.com/Azure/radius/pkg/model/revision"
	"github.com/Azure/radius/pkg/radlogger"
	"github.com/Azure/radius/pkg/radrp/db"
	"github.com/Azure/radius/pkg/radrp/deployment"
	"github.com/Azure/radius/pkg/renderers/containerv1alpha1"
	"github.com/go-logr/logr"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func createContext(t *testing.T) context.Context {
	logger, err := radlogger.NewTestLogger(t)
	if err != nil {
		t.Log("Unable to initialize logger")
		return context.Background()
	}
	return logr.NewContext(context.Background(), logger)
}

func Test_DeploymentCreated_NoComponents(t *testing.T) {
	ctx := createContext(t)
	app := db.NewApplication()
	newer := db.NewDeployment()
	rp := rp{}

	actions, err := rp.computeDeploymentActions(ctx, app, nil, newer)
	require.NoError(t, err)

	require.True(t, deploymentIsNoOp(actions))
	require.Empty(t, actions)
}

func Test_DeploymentCreated_ValidationError(t *testing.T) {
	ctx := createContext(t)
	app := db.NewApplication()
	newer := db.NewDeployment()
	newer.Properties.Components = []*db.DeploymentComponent{
		{
			ComponentName: "", // intentionally empty
		},
	}

	rp := rp{}

	_, err := rp.computeDeploymentActions(ctx, app, nil, newer)
	require.Error(t, err)
}

func Test_DeploymentCreated_ErrMissingComponent(t *testing.T) {
	ctx := createContext(t)
	app := db.NewApplication()
	newer := db.NewDeployment()
	newer.Properties.Components = []*db.DeploymentComponent{
		{
			ComponentName: "A",
		},
	}

	rp := rp{}

	_, err := rp.computeDeploymentActions(ctx, app, nil, newer)
	require.Error(t, err)
}

func Test_DeploymentCreated_OneComponent(t *testing.T) {
	ctx := createContext(t)
	app := db.NewApplication()
	app.Components["A"] = db.Component{
		Revision:   revision.Revision("1"),
		Kind:       containerv1alpha1.Kind,
		Properties: *db.NewComponentProperties(),
	}

	newer := db.NewDeployment()
	newer.Properties.Components = []*db.DeploymentComponent{
		{
			ComponentName: "A",
		},
	}

	rp := rp{}

	actions, err := rp.computeDeploymentActions(ctx, app, nil, newer)
	require.NoError(t, err)

	require.False(t, deploymentIsNoOp(actions))

	// Updates to the components are in the actions
	require.Len(t, actions, 1)

	require.Contains(t, actions, "A")
	action := actions["A"]
	require.Equal(t, deployment.CreateWorkload, action.Operation)
	require.Equal(t, "A", action.ComponentName)
	require.Equal(t, app.Components["A"], *action.Definition)
	require.Equal(t, revision.Revision(""), action.OldRevision)
	require.Equal(t, revision.Revision("1"), action.NewRevision)
}

func Test_DeploymentCreated_MultipleComponents(t *testing.T) {
	ctx := createContext(t)
	app := db.NewApplication()
	app.Components["A"] = db.Component{
		Kind:       containerv1alpha1.Kind,
		Revision:   revision.Revision("1"),
		Properties: *db.NewComponentProperties(),
	}
	app.Components["B"] = db.Component{
		Kind:       containerv1alpha1.Kind,
		Revision:   revision.Revision("1"),
		Properties: *db.NewComponentProperties(),
	}

	app.Components["C"] = db.Component{
		Kind:       containerv1alpha1.Kind,
		Revision:   revision.Revision("1"),
		Properties: *db.NewComponentProperties(),
	}

	newer := db.NewDeployment()
	newer.Properties.Components = []*db.DeploymentComponent{
		{
			ComponentName: "A",
		},
		{
			ComponentName: "B",
		},
		{
			ComponentName: "C",
		},
	}

	rp := rp{}

	actions, err := rp.computeDeploymentActions(ctx, app, nil, newer)
	require.NoError(t, err)

	require.False(t, deploymentIsNoOp(actions))

	// Revision is updated in the deployment.
	require.Equal(t, revision.Revision("1"), newer.Properties.Components[0].Revision)
	require.Equal(t, revision.Revision("1"), newer.Properties.Components[1].Revision)
	require.Equal(t, revision.Revision("1"), newer.Properties.Components[2].Revision)

	// Updates to the components are in the actions
	require.Len(t, actions, 3)

	require.Contains(t, actions, "A")
	action := actions["A"]
	require.Equal(t, deployment.CreateWorkload, action.Operation)
	require.Equal(t, "A", action.ComponentName)
	require.Equal(t, app.Components["A"], *action.Definition)
	require.Equal(t, revision.Revision(""), action.OldRevision)
	require.Equal(t, revision.Revision("1"), action.NewRevision)

	require.Contains(t, actions, "B")
	action = actions["B"]
	require.Equal(t, deployment.CreateWorkload, action.Operation)
	require.Equal(t, "B", action.ComponentName)
	require.Equal(t, app.Components["B"], *action.Definition)
	require.Equal(t, revision.Revision(""), action.OldRevision)
	require.Equal(t, revision.Revision("1"), action.NewRevision)

	require.Contains(t, actions, "C")
	action = actions["C"]
	require.Equal(t, deployment.CreateWorkload, action.Operation)
	require.Equal(t, "C", action.ComponentName)
	require.Equal(t, app.Components["C"], *action.Definition)
	require.Equal(t, revision.Revision(""), action.OldRevision)
	require.Equal(t, revision.Revision("1"), action.NewRevision)
}

func Test_DeploymentUpdate_RegistersForHealthChecks(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockDeploymentProcessor := deployment.NewMockDeploymentProcessor(ctrl)
	mockDeploymentProcessor.EXPECT().RegisterForHealthChecks(gomock.Any(), gomock.Any(), gomock.Any()).Times(3).Return(nil)

	rp := rp{
		deploy: mockDeploymentProcessor,
	}

	actions := map[string]deployment.ComponentAction{
		"C1": {
			ApplicationName: "A",
			ComponentName:   "C1",
			Operation:       deployment.UpdateWorkload,
			Definition: &db.Component{
				Kind:       containerv1alpha1.Kind,
				Revision:   revision.Revision("1"),
				Properties: *db.NewComponentProperties(),
			},
		},
		"C2": {
			ApplicationName: "A",
			ComponentName:   "C2",
			Operation:       deployment.UpdateWorkload,
			Definition: &db.Component{
				Kind:       containerv1alpha1.Kind,
				Revision:   revision.Revision("1"),
				Properties: *db.NewComponentProperties(),
			},
		},
		"C3": {
			ApplicationName: "A",
			ComponentName:   "C3",
			Operation:       deployment.UpdateWorkload,
			Definition: &db.Component{
				Kind:       containerv1alpha1.Kind,
				Revision:   revision.Revision("1"),
				Properties: *db.NewComponentProperties(),
			},
		},
	}

	ctx := createContext(t)
	err := rp.registerForHealthChecks(ctx, actions)
	require.NoError(t, err)

}

func Test_DeploymentUpdated_OneComponent_Deleted(t *testing.T) {
	ctx := createContext(t)
	app := db.NewApplication()
	app.Components["A"] = db.Component{
		Kind:       containerv1alpha1.Kind,
		Revision:   revision.Revision("1"),
		Properties: *db.NewComponentProperties(),
	}

	older := db.NewDeployment()
	older.Properties.Components = []*db.DeploymentComponent{
		{
			ComponentName: "A",
			Revision:      revision.Revision("1"),
		},
	}

	newer := db.NewDeployment()
	newer.Properties.Components = []*db.DeploymentComponent{}

	rp := rp{}

	actions, err := rp.computeDeploymentActions(ctx, app, older, newer)
	require.NoError(t, err)

	require.False(t, deploymentIsNoOp(actions))

	// Updates to the components are in the actions
	require.Len(t, actions, 1)

	require.Contains(t, actions, "A")
	action := actions["A"]
	require.Equal(t, deployment.DeleteWorkload, action.Operation)
	require.Equal(t, "A", action.ComponentName)
	require.Nil(t, action.Definition)
	require.Equal(t, revision.Revision("1"), action.OldRevision)
	require.Equal(t, revision.Revision(""), action.NewRevision)
}

func Test_DeploymentUpdated_OneComponent_NoAction(t *testing.T) {
	ctx := createContext(t)
	app := db.NewApplication()
	app.Components["A"] = db.Component{
		Kind:       containerv1alpha1.Kind,
		Revision:   revision.Revision("2"),
		Properties: *db.NewComponentProperties(),
	}

	older := db.NewDeployment()
	older.Properties.Components = []*db.DeploymentComponent{
		{
			ComponentName: "A",
			Revision:      revision.Revision("2"),
		},
	}

	newer := db.NewDeployment()
	newer.Properties.Components = []*db.DeploymentComponent{
		{
			ComponentName: "A",
			Revision:      revision.Revision("2"),
		},
	}

	rp := rp{}

	actions, err := rp.computeDeploymentActions(ctx, app, older, newer)
	require.NoError(t, err)

	require.True(t, deploymentIsNoOp(actions))

	// Updates to the components are in the actions
	require.Len(t, actions, 1)

	require.Contains(t, actions, "A")
	action := actions["A"]
	require.Equal(t, deployment.None, action.Operation)
	require.Equal(t, "A", action.ComponentName)
	require.Equal(t, app.Components["A"], *action.Definition)
	require.Equal(t, revision.Revision("2"), action.OldRevision)
	require.Equal(t, revision.Revision("2"), action.NewRevision)
}

func Test_DeploymentUpdated_OneComponent_RevisionUpgraded(t *testing.T) {
	ctx := createContext(t)
	app := db.NewApplication()
	app.Components["A"] = db.Component{
		Kind:       containerv1alpha1.Kind,
		Revision:   revision.Revision("2"),
		Properties: *db.NewComponentProperties(),
	}

	older := db.NewDeployment()
	older.Properties.Components = []*db.DeploymentComponent{
		{
			ComponentName: "A",
			Revision:      revision.Revision("1"),
		},
	}

	newer := db.NewDeployment()
	newer.Properties.Components = []*db.DeploymentComponent{
		{
			ComponentName: "A",
			Revision:      revision.Revision("2"),
		},
	}

	rp := rp{}

	actions, err := rp.computeDeploymentActions(ctx, app, older, newer)
	require.NoError(t, err)

	require.False(t, deploymentIsNoOp(actions))

	// Updates to the components are in the actions
	require.Len(t, actions, 1)

	require.Contains(t, actions, "A")
	action := actions["A"]
	require.Equal(t, deployment.UpdateWorkload, action.Operation)
	require.Equal(t, "A", action.ComponentName)
	require.Equal(t, app.Components["A"], *action.Definition)
	require.Equal(t, revision.Revision("1"), action.OldRevision)
	require.Equal(t, revision.Revision("2"), action.NewRevision)
}

func Test_DeploymentCreated_MultipleComponents_ServiceBinding(t *testing.T) {
	ctx := createContext(t)
	app := db.NewApplication()
	app.Name = "testapp"
	app.Components["A"] = db.Component{
		Kind:     containerv1alpha1.Kind,
		Revision: revision.Revision("1"),
		Properties: db.ComponentProperties{
			Build: map[string]interface{}{},
			Run:   map[string]interface{}{},
			Bindings: map[string]db.ComponentBinding{
				"web": {
					Kind: "http",
				},
			},
			Uses: []db.ComponentDependency{
				{
					Binding: components.NewComponentBindingExpression(app.Name, "B", "web", ""),
				},
			},
		},
	}

	app.Components["B"] = db.Component{
		Kind:     containerv1alpha1.Kind,
		Revision: revision.Revision("1"),
		Properties: db.ComponentProperties{
			Build: map[string]interface{}{},
			Run:   map[string]interface{}{},
			Bindings: map[string]db.ComponentBinding{
				"web": {
					Kind: "http",
					AdditionalProperties: map[string]interface{}{
						"port": 80,
					},
				},
			},
			Uses: []db.ComponentDependency{
				{
					Binding: components.NewComponentBindingExpression(app.Name, "A", "web", ""),
				},
			},
		},
	}

	newer := db.NewDeployment()
	newer.Properties.Components = []*db.DeploymentComponent{
		{
			ComponentName: "A",
			Revision:      revision.Revision("1"),
		},
		{
			ComponentName: "B",
			Revision:      revision.Revision("1"),
		},
	}

	rp := rp{}

	actions, err := rp.computeDeploymentActions(ctx, app, nil, newer)
	require.NoError(t, err)

	require.False(t, deploymentIsNoOp(actions))

	// Updates to the components are in the actions
	require.Len(t, actions, 2)

	require.Contains(t, actions, "A")
	action := actions["A"]
	require.Equal(t, deployment.CreateWorkload, action.Operation)
	require.Equal(t, "A", action.ComponentName)

	require.Contains(t, actions, "B")
	action = actions["B"]
	require.Equal(t, deployment.CreateWorkload, action.Operation)
	require.Equal(t, "B", action.ComponentName)
}

func Test_DeploymentUpdated_RenderRealisticContainer(t *testing.T) {
	ctx := createContext(t)
	app := db.NewApplication()
	app.Name = "radius/myapp"
	app.Components["A"] = db.Component{
		Kind:     containerv1alpha1.Kind,
		Revision: revision.Revision("1"),
		Properties: db.ComponentProperties{
			Build: map[string]interface{}{},
			Run: map[string]interface{}{
				"container": map[string]interface{}{
					"image": "rynowak/frontend:0.5.0-dev",
					"env": map[string]interface{}{
						"SERVICE__BACKEND__HOST": "backend",
						"SERVICE__BACKEND__PORT": "80",
					},
				},
			},
		},
	}

	newer := db.NewDeployment()
	newer.Properties.Components = []*db.DeploymentComponent{
		{
			ComponentName: "A",
			Revision:      revision.Revision("1"),
		},
	}

	rp := rp{}

	actions, err := rp.computeDeploymentActions(ctx, app, nil, newer)
	require.NoError(t, err)

	require.False(t, deploymentIsNoOp(actions))

	// Updates to the components are in the actions
	require.Len(t, actions, 1)

	require.Contains(t, actions, "A")
	action := actions["A"]
	require.Equal(t, deployment.CreateWorkload, action.Operation)
	require.Equal(t, "myapp", action.ApplicationName)
	require.Equal(t, "A", action.ComponentName)
	require.Equal(t, app.Components["A"], *action.Definition)

	// validate the workload
	require.Equal(t, containerv1alpha1.Kind, action.Component.Kind)
	require.Equal(t, "A", action.Component.Name)

	component := containerv1alpha1.ContainerComponent{}
	err = action.Component.AsRequired(containerv1alpha1.Kind, &component)
	require.NoError(t, err)

	cont := component.Run.Container
	require.Equal(t, "rynowak/frontend:0.5.0-dev", cont.Image)

	require.Len(t, cont.Env, 2)
	require.Equal(t, "backend", cont.Env["SERVICE__BACKEND__HOST"])
	require.Equal(t, "80", cont.Env["SERVICE__BACKEND__PORT"])
}

func Test_DeploymentCreated_RenderContainerWithDapr(t *testing.T) {
	ctx := createContext(t)
	app := db.NewApplication()
	app.Name = "radius/myapp"
	app.Components["A"] = db.Component{
		Kind:     containerv1alpha1.Kind,
		Revision: revision.Revision("1"),
		Properties: db.ComponentProperties{
			Build: map[string]interface{}{},
			Run: map[string]interface{}{
				"container": map[string]interface{}{
					"image": "rynowak/frontend:0.5.0-dev",
				},
			},
			Traits: []db.ComponentTrait{
				{
					Kind: "dapr.io/Sidecar@v1alpha1",
					AdditionalProperties: map[string]interface{}{
						"appId":   "frontend",
						"appPort": 80,
					},
				},
			},
		},
	}

	newer := db.NewDeployment()
	newer.Properties.Components = []*db.DeploymentComponent{
		{
			ComponentName: "A",
			Revision:      revision.Revision("1"),
		},
	}

	rp := rp{}

	actions, err := rp.computeDeploymentActions(ctx, app, nil, newer)
	require.NoError(t, err)

	require.False(t, deploymentIsNoOp(actions))

	// Updates to the components are in the actions
	require.Len(t, actions, 1)

	require.Contains(t, actions, "A")
	action := actions["A"]
	require.Equal(t, deployment.CreateWorkload, action.Operation)
	require.Equal(t, "A", action.ComponentName)
	require.Equal(t, app.Components["A"], *action.Definition)

	require.Equal(t, "dapr.io/Sidecar@v1alpha1", action.Component.Traits[0].Kind)
	require.Equal(t, map[string]interface{}{
		"appId":   "frontend",
		"appPort": 80,
	}, action.Component.Traits[0].AdditionalProperties)
}
