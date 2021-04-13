package action_test

import (
	"encoding/json"
	"testing"
	"time"

	v1 "k8s.io/api/authentication/v1"

	"capact.io/capact/internal/k8s-engine/graphql/model"
	"capact.io/capact/internal/ptr"
	"capact.io/capact/pkg/engine/api/graphql"
	"capact.io/capact/pkg/engine/k8s/api/v1alpha1"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func fixGQLAction(t *testing.T, name string) graphql.Action {
	timestamp, err := time.Parse(time.UnixDate, "Wed Feb 25 11:06:39 PST 2015")
	require.NoError(t, err)

	userInfo := graphql.UserInfo{
		Username: "user",
		Groups:   []string{"group1", "group2"},
		Extra: map[string][]string{
			"extra": {"data"},
		},
	}

	return graphql.Action{
		Name:      name,
		CreatedAt: graphql.Timestamp{Time: timestamp},
		Input: &graphql.ActionInput{
			Parameters: ptrToJSONRawMessage(`{"param":"one"}`),
			TypeInstances: []*graphql.InputTypeInstanceDetails{
				{
					Name: "in1",
					ID:   "in-id1",
					TypeRef: &graphql.ManifestReference{
						Path:     "path1",
						Revision: "0.1.0",
					},
					Optional: false,
				},
				{
					Name: "in2",
					ID:   "in-id2",
					TypeRef: &graphql.ManifestReference{
						Path:     "path2",
						Revision: "0.1.0",
					},
					Optional: true,
				},
			},
		},
		Output: &graphql.ActionOutput{
			TypeInstances: []*graphql.OutputTypeInstanceDetails{
				{
					Name: "out1",
					ID:   "id1",
					TypeRef: &graphql.ManifestReference{
						Path:     "path1",
						Revision: "0.1.0",
					},
				},
				{
					Name: "out2",
					ID:   "id2",
					TypeRef: &graphql.ManifestReference{
						Path:     "path2",
						Revision: "0.1.0",
					},
				},
			},
		},
		ActionRef: &graphql.ManifestReference{
			Path:     "foo.bar",
			Revision: "0.1.0",
		},
		Run:            true,
		DryRun:         true,
		Cancel:         true,
		RenderedAction: ptrToJSONRawMessage(`{"foo":"bar","baz":3}`),
		RenderingAdvancedMode: &graphql.ActionRenderingAdvancedMode{
			Enabled: true,
			TypeInstancesForRenderingIteration: []*graphql.InputTypeInstanceToProvide{
				{
					Name: "typeinstance1",
					TypeRef: &graphql.ManifestReference{
						Path:     "cap.type.one",
						Revision: "latest",
					},
				},
				{
					Name: "typeinstance2",
					TypeRef: &graphql.ManifestReference{
						Path:     "cap.type.two",
						Revision: "0.1.0",
					},
				},
			},
		},
		RenderedActionOverride: ptrToJSONRawMessage(`{"override":true}`),
		Status: &graphql.ActionStatus{
			Phase:     graphql.ActionStatusPhaseSucceeded,
			Timestamp: graphql.Timestamp{Time: timestamp},
			Message:   ptr.String("message"),
			Runner: &graphql.RunnerStatus{
				Status: ptrToJSONRawMessage(`{"runner":true}`),
			},
			CreatedBy:  &userInfo,
			RunBy:      &userInfo,
			CanceledBy: &userInfo,
		},
	}
}

func fixK8sAction(t *testing.T, name, namespace string) v1alpha1.Action {
	timestamp, err := time.Parse(time.UnixDate, "Wed Feb 25 11:06:39 PST 2015")
	require.NoError(t, err)

	userInfo := v1.UserInfo{
		Username: "user",
		Groups:   []string{"group1", "group2"},
		Extra: map[string]v1.ExtraValue{
			"extra": []string{"data"},
		},
	}

	return v1alpha1.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         namespace,
			CreationTimestamp: metav1.NewTime(timestamp),
		},
		Spec: v1alpha1.ActionSpec{
			ActionRef: v1alpha1.ManifestReference{
				Path:     "foo.bar",
				Revision: ptr.String("0.1.0"),
			},
			DryRun: ptr.Bool(true),
			Input: &v1alpha1.ActionInput{
				TypeInstances: &[]v1alpha1.InputTypeInstance{
					{
						Name: "in1",
						ID:   "in-id1",
					},
					{
						ID:   "in-id2",
						Name: "in2",
					},
				},
				Parameters: &v1alpha1.InputParameters{
					SecretRef: corev1.LocalObjectReference{
						Name: "secret",
					},
				},
			},
			AdvancedRendering: &v1alpha1.AdvancedRendering{
				Enabled: true,
			},
			RenderedActionOverride: &runtime.RawExtension{Raw: []byte(`{"override":true}`)},
			Run:                    ptr.Bool(true),
			Cancel:                 ptr.Bool(true),
		},
		Status: v1alpha1.ActionStatus{
			Phase:   v1alpha1.SucceededActionPhase,
			Message: ptr.String("message"),
			Runner: &v1alpha1.RunnerStatus{
				Status: &runtime.RawExtension{Raw: []byte(`{"runner":true}`)},
			},
			Output: &v1alpha1.ActionOutput{
				TypeInstances: &[]v1alpha1.OutputTypeInstanceDetails{
					{
						CommonTypeInstanceDetails: v1alpha1.CommonTypeInstanceDetails{
							Name: "out1",
							ID:   "id1",
							TypeRef: &v1alpha1.ManifestReference{
								Path:     "path1",
								Revision: ptr.String("0.1.0"),
							},
						},
					},
					{
						CommonTypeInstanceDetails: v1alpha1.CommonTypeInstanceDetails{
							Name: "out2",
							ID:   "id2",
							TypeRef: &v1alpha1.ManifestReference{
								Path:     "path2",
								Revision: ptr.String("0.1.0"),
							},
						},
					},
				},
			},
			Rendering: &v1alpha1.RenderingStatus{
				Action: &runtime.RawExtension{Raw: []byte(`{"foo":"bar","baz":3}`)},
				Input: &v1alpha1.ResolvedActionInput{
					Parameters: &runtime.RawExtension{Raw: []byte(`{"param":"one"}`)},
					TypeInstances: &[]v1alpha1.InputTypeInstanceDetails{
						{
							CommonTypeInstanceDetails: v1alpha1.CommonTypeInstanceDetails{
								Name: "in1",
								ID:   "in-id1",
								TypeRef: &v1alpha1.ManifestReference{
									Path:     "path1",
									Revision: ptr.String("0.1.0"),
								},
							},
							Optional: false,
						},
						{
							CommonTypeInstanceDetails: v1alpha1.CommonTypeInstanceDetails{
								Name: "in2",
								ID:   "in-id2",
								TypeRef: &v1alpha1.ManifestReference{
									Path:     "path2",
									Revision: ptr.String("0.1.0"),
								},
							},
							Optional: true,
						},
					},
				},
				AdvancedRendering: &v1alpha1.AdvancedRenderingStatus{
					RenderingIteration: &v1alpha1.RenderingIterationStatus{
						CurrentIterationName: "rendering-iteration",
						InputTypeInstancesToProvide: &[]v1alpha1.InputTypeInstanceToProvide{
							{
								Name: "typeinstance1",
								TypeRef: &v1alpha1.ManifestReference{
									Path:     "cap.type.one",
									Revision: nil,
								},
							},
							{
								Name: "typeinstance2",
								TypeRef: &v1alpha1.ManifestReference{
									Path:     "cap.type.two",
									Revision: ptr.String("0.1.0"),
								},
							},
						},
					},
				},
			},
			CreatedBy:          &userInfo,
			RunBy:              &userInfo,
			CanceledBy:         &userInfo,
			LastTransitionTime: metav1.NewTime(timestamp),
		},
	}
}

func fixK8sActionMinimal(name, namespace string, phase v1alpha1.ActionPhase) v1alpha1.Action {
	return v1alpha1.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.ActionSpec{
			ActionRef: v1alpha1.ManifestReference{
				Path:     "foo.bar",
				Revision: ptr.String("0.1.0"),
			},
		},
		Status: v1alpha1.ActionStatus{
			Phase: phase,
		},
	}
}

func fixK8sActionForRenderingIteration(name, namespace string) v1alpha1.Action {
	return v1alpha1.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.ActionSpec{
			ActionRef: v1alpha1.ManifestReference{
				Path:     "foo.bar",
				Revision: ptr.String("0.1.0"),
			},
			AdvancedRendering: &v1alpha1.AdvancedRendering{
				Enabled: true,
			},
			Input: &v1alpha1.ActionInput{
				TypeInstances: &[]v1alpha1.InputTypeInstance{
					{
						Name: "foo",
						ID:   "f6f73e3e-20c8-4712-b415-0cf82f9f5010",
					},
				},
			},
		},
		Status: v1alpha1.ActionStatus{
			Phase: v1alpha1.AdvancedModeRenderingIterationActionPhase,
			Rendering: &v1alpha1.RenderingStatus{
				AdvancedRendering: &v1alpha1.AdvancedRenderingStatus{
					RenderingIteration: &v1alpha1.RenderingIterationStatus{
						CurrentIterationName: "iteration-name",
						InputTypeInstancesToProvide: &[]v1alpha1.InputTypeInstanceToProvide{
							{
								Name: "typeinstance1",
								TypeRef: &v1alpha1.ManifestReference{
									Path:     "cap.type.one",
									Revision: nil,
								},
							},
							{
								Name: "typeinstance2",
								TypeRef: &v1alpha1.ManifestReference{
									Path:     "cap.type.two",
									Revision: ptr.String("0.1.0"),
								},
							},
						},
					},
				},
			},
		},
	}
}

func fixGQLInput(name string) graphql.ActionDetailsInput {
	params := graphql.JSON(`{"param":"one"}`)
	override := graphql.JSON(`{"foo":"bar"}`)

	return graphql.ActionDetailsInput{
		Name: name,
		Input: &graphql.ActionInputData{
			Parameters: &params,
			TypeInstances: []*graphql.InputTypeInstanceData{
				{
					Name: "in1",
					ID:   "in-id1",
				},
				{
					Name: "in2",
					ID:   "in-id2",
				},
			},
		},
		DryRun: ptr.Bool(true),
		ActionRef: &graphql.ManifestReferenceInput{
			Path:     "sample.action",
			Revision: ptr.String("0.1.0"),
		},
		AdvancedRendering:      ptr.Bool(true),
		RenderedActionOverride: &override,
	}
}

func fixModel(name string) model.ActionToCreateOrUpdate {
	return model.ActionToCreateOrUpdate{
		Action: v1alpha1.Action{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha1.ActionKind,
				APIVersion: v1alpha1.GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: v1alpha1.ActionSpec{
				ActionRef: v1alpha1.ManifestReference{
					Path:     "sample.action",
					Revision: ptr.String("0.1.0"),
				},
				DryRun: ptr.Bool(true),
				Input: &v1alpha1.ActionInput{
					Parameters: &v1alpha1.InputParameters{
						SecretRef: corev1.LocalObjectReference{
							Name: name,
						},
					},
					TypeInstances: &[]v1alpha1.InputTypeInstance{
						{
							Name: "in1",
							ID:   "in-id1",
						},
						{
							Name: "in2",
							ID:   "in-id2",
						},
					},
				},
				AdvancedRendering: &v1alpha1.AdvancedRendering{
					Enabled: true,
				},
				RenderedActionOverride: &runtime.RawExtension{Raw: []byte(`{"foo":"bar"}`)},
			},
		},
		InputParamsSecret: &corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Secret",
				APIVersion: corev1.SchemeGroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			StringData: map[string]string{
				"parameters.json": `{"param":"one"}`,
			},
		},
	}
}

func fixModelActionFilter(phase *v1alpha1.ActionPhase) model.ActionFilter {
	return model.ActionFilter{
		Phase: phase,
	}
}

func fixGQLAdvancedRenderingIterationInput() graphql.AdvancedModeContinueRenderingInput {
	return graphql.AdvancedModeContinueRenderingInput{
		TypeInstances: []*graphql.InputTypeInstanceData{
			{
				Name: "in1",
				ID:   "in-id1",
			},
			{
				Name: "in2",
				ID:   "in-id2",
			},
		},
	}
}

func fixModelAdvancedRenderingIterationInput() model.AdvancedModeContinueRenderingInput {
	return model.AdvancedModeContinueRenderingInput{
		TypeInstances: &[]v1alpha1.InputTypeInstance{
			{
				Name: "in1",
				ID:   "in-id1",
			},
			{
				Name: "in2",
				ID:   "in-id2",
			},
		},
	}
}

func ptrToJSONRawMessage(jsonString string) *json.RawMessage {
	var jsonRaw json.RawMessage = []byte(jsonString)
	return &jsonRaw
}
