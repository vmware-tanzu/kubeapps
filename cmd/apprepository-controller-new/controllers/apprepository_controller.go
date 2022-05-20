// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"
	"time"

	helmgetter "helm.sh/helm/v3/pkg/getter"
	"k8s.io/apimachinery/pkg/runtime"
	kuberecorder "k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/ratelimiter"

	"github.com/fluxcd/pkg/apis/meta"
	helper "github.com/fluxcd/pkg/runtime/controller"
	"github.com/fluxcd/pkg/runtime/patch"
	"github.com/vmware-tanzu/kubeapps/apprepository-controller-new/api/v1alpha2"
	"github.com/vmware-tanzu/kubeapps/apprepository-controller-new/internal/helm/repository"
	sreconcile "github.com/vmware-tanzu/kubeapps/apprepository-controller-new/internal/reconcile"
	"github.com/vmware-tanzu/kubeapps/apprepository-controller-new/internal/reconcile/summarize"
)

// appRepositoryReadyCondition contains the information required to summarize a
// v1alpha2.AppRepository Ready Condition.
var appRepositoryReadyCondition = summarize.Conditions{
	Target: meta.ReadyCondition,
	Owned: []string{
		v1alpha2.StorageOperationFailedCondition,
		v1alpha2.FetchFailedCondition,
		v1alpha2.ArtifactOutdatedCondition,
		v1alpha2.ArtifactInStorageCondition,
		meta.ReadyCondition,
		meta.ReconcilingCondition,
		meta.StalledCondition,
	},
	Summarize: []string{
		v1alpha2.StorageOperationFailedCondition,
		v1alpha2.FetchFailedCondition,
		v1alpha2.ArtifactOutdatedCondition,
		v1alpha2.ArtifactInStorageCondition,
		meta.StalledCondition,
		meta.ReconcilingCondition,
	},
	NegativePolarity: []string{
		v1alpha2.StorageOperationFailedCondition,
		v1alpha2.FetchFailedCondition,
		v1alpha2.ArtifactOutdatedCondition,
		meta.StalledCondition,
		meta.ReconcilingCondition,
	},
}

//+kubebuilder:rbac:groups=kubeapps.com,resources=apprepositories,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubeapps.com,resources=apprepositories/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubeapps.com,resources=apprepositories/finalizers,verbs=update

// AppRepositoryReconciler reconciles a AppRepository object
type AppRepositoryReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	Getters helmgetter.Providers
	helper.Metrics
	kuberecorder.EventRecorder
	ControllerName string
}

type AppRepositoryReconcilerOptions struct {
	MaxConcurrentReconciles int
	RateLimiter             ratelimiter.RateLimiter
}

// appRepositoryReconcileFunc is the function type for all the
// v1alpha2.AppRepository (sub)reconcile functions. The type implementations
// are grouped and executed serially to perform the complete reconcile of the
// object.
type appRepositoryReconcileFunc func(ctx context.Context, obj *v1alpha2.AppRepository, artifact *v1alpha2.Artifact, repo *repository.ChartRepository) (sreconcile.Result, error)

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AppRepository object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.2/pkg/reconcile
func (r *AppRepositoryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, retErr error) {
	defer func() {
		msg := fmt.Sprintf("-Reconcile [%s]", req.NamespacedName)
		defer ctrl.LoggerFrom(ctx).Info(msg)
	}()

	msg := fmt.Sprintf("+Reconcile [%s]", req.NamespacedName)
	ctrl.LoggerFrom(ctx).Info(msg)

	start := time.Now()
	_ = ctrl.LoggerFrom(ctx)

	// Fetch the AppRepository
	obj := &v1alpha2.AppRepository{}
	if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Initialize the patch helper with the current version of the object.
	patchHelper, err := patch.NewHelper(obj, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}

	// recResult stores the abstracted reconcile result.
	var recResult sreconcile.Result

	// Always attempt to patch the object after each reconciliation.
	// NOTE: The final runtime result and error are set in this block.
	defer func() {
		summarizeHelper := summarize.NewHelper(r.EventRecorder, patchHelper)
		summarizeOpts := []summarize.Option{
			summarize.WithConditions(appRepositoryReadyCondition),
			summarize.WithReconcileResult(recResult),
			summarize.WithReconcileError(retErr),
			summarize.WithIgnoreNotFound(),
			summarize.WithProcessors(
				summarize.RecordContextualError,
				summarize.RecordReconcileReq,
			),
			summarize.WithResultBuilder(sreconcile.AlwaysRequeueResultBuilder{RequeueAfter: obj.GetRequeueAfter()}),
			summarize.WithPatchFieldOwner(r.ControllerName),
		}
		result, retErr = summarizeHelper.SummarizeAndPatch(ctx, obj, summarizeOpts...)

		// Always record readiness and duration metrics
		r.Metrics.RecordReadiness(ctx, obj)
		r.Metrics.RecordDuration(ctx, obj, start)
	}()

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AppRepositoryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha2.AppRepository{}).
		Complete(r)
}
