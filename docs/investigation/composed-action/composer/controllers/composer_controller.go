/*
Copyright 2022.

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

package controllers

import (
	"context"
	"time"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1 "capact.io/composer/api/v1"
)

const (
	// noWait is used when requeue is needed
	// has to be higher than 0
	noWait = 1 * time.Microsecond
)

// ComposerReconciler reconciles a Composer object
type ComposerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core.capact.io,resources=composers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.capact.io,resources=composers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.capact.io,resources=composers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Composer object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *ComposerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	reportOnError := func(err error, context string) (ctrl.Result, error) {
		return ctrl.Result{}, errors.Wrap(err, context)
	}

	action := &corev1.Composer{}
	if err := r.Get(ctx, req.NamespacedName, action); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "while fetching Composer CR")
		return ctrl.Result{}, err
	}

	if action.IsUninitialized() {
		log.Info("Initializing composer status entry")
		result, err := r.initStatus(ctx, action)
		if err != nil {
			return reportOnError(err, "Init runner action")
		}
		return result, nil
	}

	if action.IsRunning() {
		log.Info("Checking running Action status")
		if err != nil {
			return reportOnError(err, "Delete runner action")
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ComposerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Composer{}).
		Complete(r)
}

func (r *ComposerReconciler) initStatus(ctx context.Context, action *corev1.Composer) (ctrl.Result, error) {
	action.Status.Phase = corev1.InitialComposerPhase

	for name := range action.Spec.Steps {
		action.Status.Results = append(action.Status.Results, corev1.ComposerResult{
			Name:  name,
			Phase: corev1.InitialComposerPhase,
		})
	}

	if err := r.Status().Update(ctx, action); err != nil {
		return ctrl.Result{}, errors.Wrap(err, "while updating action object status")
	}
	return ctrl.Result{RequeueAfter: noWait}, nil
}
