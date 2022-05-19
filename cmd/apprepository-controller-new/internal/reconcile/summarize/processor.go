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

	corev1 "k8s.io/api/core/v1"
	kuberecorder "k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/fluxcd/pkg/apis/meta"
	serror "github.com/vmware-tanzu/kubeapps/apprepository-controller-new/internal/error"
	"github.com/vmware-tanzu/kubeapps/apprepository-controller-new/internal/object"
	"github.com/vmware-tanzu/kubeapps/apprepository-controller-new/internal/reconcile"
)

// ResultProcessor processes the results of reconciliation (the object, result
// and error). Any errors during processing need not result in the
// reconciliation failure. The errors can be recorded as logs and events.
type ResultProcessor func(context.Context, kuberecorder.EventRecorder, client.Object, reconcile.Result, error)

// RecordContextualError is a ResultProcessor that records the contextual errors
// based on their types.
// An event is recorded for the errors that are returned to the runtime. The
// runtime handles the logging of the error.
// An event is recorded and an error is logged for errors that are known to be
// swallowed, not returned to the runtime.
func RecordContextualError(ctx context.Context, recorder kuberecorder.EventRecorder, obj client.Object, _ reconcile.Result, err error) {
	switch e := err.(type) {
	case *serror.Event:
		recorder.Eventf(obj, corev1.EventTypeWarning, e.Reason, e.Error())
	case *serror.Waiting:
		// Waiting errors are not returned to the runtime. Log it explicitly.
		ctrl.LoggerFrom(ctx).Info("reconciliation waiting", "reason", e.Err, "duration", e.RequeueAfter)
		recorder.Event(obj, corev1.EventTypeNormal, e.Reason, e.Error())
	case *serror.Stalling:
		// Stalling errors are not returned to the runtime. Log it explicitly.
		ctrl.LoggerFrom(ctx).Error(e, "reconciliation stalled")
		recorder.Eventf(obj, corev1.EventTypeWarning, e.Reason, e.Error())
	}
}

// RecordReconcileReq is a ResultProcessor that checks the reconcile
// annotation value and sets it in the object status as
// status.lastHandledReconcileAt.
func RecordReconcileReq(ctx context.Context, recorder kuberecorder.EventRecorder, obj client.Object, _ reconcile.Result, _ error) {
	if v, ok := meta.ReconcileAnnotationValue(obj.GetAnnotations()); ok {
		object.SetStatusLastHandledReconcileAt(obj, v)
	}
}
