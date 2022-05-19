/*
Copyright 2021 The Flux authors

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

package reconcile

import (
	"time"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/fluxcd/pkg/apis/meta"
	"github.com/fluxcd/pkg/runtime/conditions"
	"github.com/fluxcd/pkg/runtime/patch"

	serror "github.com/vmware-tanzu/kubeapps/apprepository-controller-new/internal/error"
)

// Result is a type for creating an abstraction for the controller-runtime
// reconcile Result to simplify the Result values.
type Result int

const (
	// ResultEmpty indicates a reconcile result which does not requeue. It is
	// also used when returning an error, since the error overshadows result.
	ResultEmpty Result = iota
	// ResultRequeue indicates a reconcile result which should immediately
	// requeue.
	ResultRequeue
	// ResultSuccess indicates a reconcile success result.
	// For a reconciler that requeues regularly at a fixed interval, runtime
	// result with a fixed RequeueAfter is success result.
	// For a reconciler that doesn't requeue on successful reconciliation,
	// an empty runtime result is success result.
	// It is usually returned at the end of a reconciler/sub-reconciler.
	ResultSuccess
)

// RuntimeResultBuilder defines an interface for runtime result builders. This
// can be implemented to build custom results based on the context of the
// reconciler.
type RuntimeResultBuilder interface {
	// BuildRuntimeResult analyzes the result and error to return a runtime
	// result.
	BuildRuntimeResult(rr Result, err error) ctrl.Result
	// IsSuccess returns if a given runtime result is success for a
	// RuntimeResultBuilder.
	IsSuccess(ctrl.Result) bool
}

// AlwaysRequeueResultBuilder implements a RuntimeResultBuilder for always
// requeuing reconcilers. A successful reconciliation result for such
// reconcilers contains a fixed RequeueAfter value.
type AlwaysRequeueResultBuilder struct {
	// RequeueAfter is the fixed period at which the reconciler requeues on
	// successful execution.
	RequeueAfter time.Duration
}

// BuildRuntimeResult converts a given Result and error into the
// return values of a controller's Reconcile function.
func (r AlwaysRequeueResultBuilder) BuildRuntimeResult(rr Result, err error) ctrl.Result {
	// Handle special errors that contribute to expressing the result.
	if e, ok := err.(*serror.Waiting); ok {
		return ctrl.Result{RequeueAfter: e.RequeueAfter}
	}

	switch rr {
	case ResultRequeue:
		return ctrl.Result{Requeue: true}
	case ResultSuccess:
		return ctrl.Result{RequeueAfter: r.RequeueAfter}
	default:
		return ctrl.Result{}
	}
}

// IsSuccess returns true if the given Result has the same RequeueAfter value
// as of the AlwaysRequeueResultBuilder.
func (r AlwaysRequeueResultBuilder) IsSuccess(result ctrl.Result) bool {
	return result.RequeueAfter == r.RequeueAfter
}

// ComputeReconcileResult analyzes the reconcile results (result + error),
// updates the status conditions of the object with any corrections and returns
// object patch configuration, runtime result and runtime error. The caller is
// responsible for using the patch configuration while patching the object in
// the API server.
// The RuntimeResultBuilder is used to define how the ctrl.Result is computed.
func ComputeReconcileResult(obj conditions.Setter, res Result, recErr error, rb RuntimeResultBuilder) ([]patch.Option, ctrl.Result, error) {
	var pOpts []patch.Option

	// Compute the runtime result.
	var result ctrl.Result
	if rb != nil {
		result = rb.BuildRuntimeResult(res, recErr)
	}

	// Remove reconciling condition on successful reconciliation.
	if recErr == nil && res == ResultSuccess {
		conditions.Delete(obj, meta.ReconcilingCondition)
	}

	// Analyze the reconcile error.
	switch t := recErr.(type) {
	case *serror.Stalling:
		if res == ResultEmpty {
			// The current generation has been reconciled successfully and it
			// has resulted in a stalled state. Return no error to stop further
			// requeuing.
			pOpts = append(pOpts, patch.WithStatusObservedGeneration{})
			conditions.MarkStalled(obj, t.Reason, t.Error())
			return pOpts, result, nil
		}
		// NOTE: Non-empty result with stalling error indicates that the
		// returned result is incorrect.
	case *serror.Waiting:
		// The reconcile resulted in waiting error, remove stalled condition if
		// present.
		conditions.Delete(obj, meta.StalledCondition)
		// The reconciler needs to wait and retry. Return no error.
		return pOpts, result, nil
	case nil:
		// The reconcile didn't result in any error, we are not in stalled
		// state. If a requeue is requested, the current generation has not been
		// reconciled successfully.
		if res != ResultRequeue {
			pOpts = append(pOpts, patch.WithStatusObservedGeneration{})
		}
		conditions.Delete(obj, meta.StalledCondition)
	default:
		// The reconcile resulted in some error, but we are not in stalled
		// state.
		conditions.Delete(obj, meta.StalledCondition)
	}

	return pOpts, result, recErr
}

// LowestRequeuingResult returns the ReconcileResult with the lowest requeue
// period.
// Weightage:
//  ResultRequeue - immediate requeue (lowest)
//  ResultSuccess - requeue at an interval
//  ResultEmpty - no requeue
func LowestRequeuingResult(i, j Result) Result {
	switch {
	case i == ResultEmpty:
		return j
	case j == ResultEmpty:
		return i
	case i == ResultRequeue:
		return i
	case j == ResultRequeue:
		return j
	default:
		return j
	}
}

// FailureRecovery finds out if a failure recovery occurred by checking the fail
// conditions in the old object and the new object.
func FailureRecovery(oldObj, newObj conditions.Getter, failConditions []string) bool {
	failuresBefore := 0
	for _, failCondition := range failConditions {
		if conditions.Get(oldObj, failCondition) != nil {
			failuresBefore++
		}
		if conditions.Get(newObj, failCondition) != nil {
			// Short-circuit, there is failure now, can't be a recovery.
			return false
		}
	}
	return failuresBefore > 0
}
