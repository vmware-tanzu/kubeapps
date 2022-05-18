// Copyright XXXX-YYYY the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kubeappscomv1alpha2 "github.com/vmware-tanzu/kubeapps/api/v1alpha2"
)

// AppRepositoryReconciler reconciles a AppRepository object
type AppRepositoryReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=kubeapps.com,resources=apprepositories,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubeapps.com,resources=apprepositories/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubeapps.com,resources=apprepositories/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AppRepository object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.2/pkg/reconcile
func (r *AppRepositoryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AppRepositoryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubeappscomv1alpha2.AppRepository{}).
		Complete(r)
}
