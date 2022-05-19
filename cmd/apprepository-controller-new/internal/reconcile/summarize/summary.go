/*
Copyright 2022 The Flux authors

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

package summarize

import (
	"context"
	"errors"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	kuberecorder "k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/fluxcd/pkg/apis/meta"
	"github.com/fluxcd/pkg/runtime/conditions"
	"github.com/fluxcd/pkg/runtime/patch"

	"github.com/vmware-tanzu/kubeapps/apprepository-controller-new/internal/reconcile"
)

// Conditions contains all the conditions information needed to summarize the
// target condition.
type Conditions struct {
	// Target is the target condition, e.g.: Ready.
	Target string
	// Owned conditions are the conditions owned by the reconciler for this
	// target condition.
	Owned []string
	// Summarize conditions are the conditions that the target condition depends
	// on.
	Summarize []string
	// NegativePolarity conditions are the conditions in Summarize with negative
	// polarity.
	NegativePolarity []string
}

// Helper is SummarizeAndPatch helper.
type Helper struct {
	recorder    kuberecorder.EventRecorder
	patchHelper *patch.Helper
}

// NewHelper returns an initialized Helper.
func NewHelper(recorder kuberecorder.EventRecorder, patchHelper *patch.Helper) *Helper {
	return &Helper{
		recorder:    recorder,
		patchHelper: patchHelper,
	}
}

// HelperOptions contains options for SummarizeAndPatch.
// Summarizing and patching at the very end of a reconciliation involves
// computing the result of the reconciler. This requires providing the
// ReconcileResult, ReconcileError and a ResultBuilder in the context of the
// reconciliation.
// For using this to perform intermediate patching in the middle of a
// reconciliation, no ReconcileResult, ReconcileError or ResultBuilder should
// be provided. Only Conditions summary would be calculated and patched.
type HelperOptions struct {
	// Conditions are conditions that needs to be summarized and persisted on
	// the object.
	Conditions []Conditions
	// Processors are chain of ResultProcessors for processing the results. This
	// can be used to analyze and modify the results. This enables injecting
	// custom middlewares in the SummarizeAndPatch operation.
	Processors []ResultProcessor
	// IgnoreNotFound can be used to ignores any resource not found error during
	// patching.
	IgnoreNotFound bool
	// ReconcileResult is the abstracted result of reconciliation.
	ReconcileResult reconcile.Result
	// ReconcileError is the reconciliation error.
	ReconcileError error
	// ResultBuilder defines how the reconciliation result is computed.
	ResultBuilder reconcile.RuntimeResultBuilder
	// PatchFieldOwner defines the field owner configuration for the Kubernetes
	// patch operation.
	PatchFieldOwner string
}

// Option is configuration that modifies SummarizeAndPatch.
type Option func(*HelperOptions)

// WithConditions sets the Conditions for which summary is calculated in
// SummarizeAndPatch.
func WithConditions(condns ...Conditions) Option {
	return func(s *HelperOptions) {
		s.Conditions = append(s.Conditions, condns...)
	}
}

// WithProcessors can be used to inject middlewares in the SummarizeAndPatch
// process, to be executed before the result calculation and patching.
func WithProcessors(rps ...ResultProcessor) Option {
	return func(s *HelperOptions) {
		s.Processors = append(s.Processors, rps...)
	}
}

// WithIgnoreNotFound skips any resource not found error during patching.
func WithIgnoreNotFound() Option {
	return func(s *HelperOptions) {
		s.IgnoreNotFound = true
	}
}

// WithResultBuilder sets the strategy for result computation in
// SummarizeAndPatch.
func WithResultBuilder(rb reconcile.RuntimeResultBuilder) Option {
	return func(s *HelperOptions) {
		s.ResultBuilder = rb
	}
}

// WithReconcileResult sets the value of input result used to calculate the
// results of reconciliation in SummarizeAndPatch.
func WithReconcileResult(rr reconcile.Result) Option {
	return func(s *HelperOptions) {
		s.ReconcileResult = rr
	}
}

// WithReconcileError sets the value of input error used to calculate the
// results reconciliation in SummarizeAndPatch.
func WithReconcileError(re error) Option {
	return func(s *HelperOptions) {
		s.ReconcileError = re
	}
}

// WithPatchFieldOwner sets the FieldOwner in the patch helper.
func WithPatchFieldOwner(fieldOwner string) Option {
	return func(s *HelperOptions) {
		s.PatchFieldOwner = fieldOwner
	}
}

// SummarizeAndPatch summarizes and patches the result to the target object.
// When used at the very end of a reconciliation, the result builder must be
// specified using the Option WithResultBuilder(). The returned result and error
// can be returned as the return values of the reconciliation.
// When used in the middle of a reconciliation, no result builder should be set
// and the result can be ignored.
func (h *Helper) SummarizeAndPatch(ctx context.Context, obj conditions.Setter, options ...Option) (ctrl.Result, error) {
	// Calculate the options.
	opts := &HelperOptions{}
	for _, o := range options {
		o(opts)
	}
	// Combined the owned conditions of all the conditions for the patcher.
	ownedConditions := []string{}
	for _, c := range opts.Conditions {
		ownedConditions = append(ownedConditions, c.Owned...)
	}
	// Patch the object, prioritizing the conditions owned by the controller in
	// case of any conflicts.
	patchOpts := []patch.Option{
		patch.WithOwnedConditions{
			Conditions: ownedConditions,
		},
	}
	if opts.PatchFieldOwner != "" {
		patchOpts = append(patchOpts, patch.WithFieldOwner(opts.PatchFieldOwner))
	}

	// Process the results of reconciliation.
	for _, processor := range opts.Processors {
		processor(ctx, h.recorder, obj, opts.ReconcileResult, opts.ReconcileError)
	}

	var result ctrl.Result
	var recErr error
	if opts.ResultBuilder != nil {
		// Compute the reconcile results, obtain patch options and reconcile error.
		var pOpts []patch.Option
		pOpts, result, recErr = reconcile.ComputeReconcileResult(obj, opts.ReconcileResult, opts.ReconcileError, opts.ResultBuilder)
		patchOpts = append(patchOpts, pOpts...)
	}

	// Summarize conditions. This must be performed only after computing the
	// reconcile result, since the object status is adjusted based on the
	// reconcile result and error.
	for _, c := range opts.Conditions {
		conditions.SetSummary(obj,
			c.Target,
			conditions.WithConditions(
				c.Summarize...,
			),
			conditions.WithNegativePolarityConditions(
				c.NegativePolarity...,
			),
		)
	}

	// If object is not stalled, result is success and runtime error is nil,
	// ensure that Ready=True. Else, use the Ready failure message as the
	// runtime error message. This ensures that the reconciliation would be
	// retried as the object isn't ready.
	// NOTE: This is applicable to Ready condition only because it is a special
	// condition in kstatus that reflects the overall state of an object.
	if isNonStalledSuccess(obj, opts.ResultBuilder, result, recErr) {
		if !conditions.IsReady(obj) {
			recErr = errors.New(conditions.GetMessage(obj, meta.ReadyCondition))
		}
	}

	// Finally, patch the resource.
	if err := h.patchHelper.Patch(ctx, obj, patchOpts...); err != nil {
		// Ignore patch error "not found" when the object is being deleted.
		if opts.IgnoreNotFound && !obj.GetDeletionTimestamp().IsZero() {
			err = kerrors.FilterOut(err, func(e error) bool { return apierrors.IsNotFound(e) })
		}
		recErr = kerrors.NewAggregate([]error{recErr, err})
	}

	return result, recErr
}

// isNonStalledSuccess checks if the reconciliation was successful and has not
// resulted in stalled situation.
func isNonStalledSuccess(obj conditions.Setter, rb reconcile.RuntimeResultBuilder, result ctrl.Result, recErr error) bool {
	if !conditions.IsStalled(obj) && recErr == nil {
		// Without result builder, it can't be determined if the result is
		// success.
		if rb != nil {
			return rb.IsSuccess(result)
		}
	}
	return false
}
