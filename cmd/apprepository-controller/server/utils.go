// Copyright 2021-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"fmt"
	"hash/adler32"
	"math"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	log "k8s.io/klog/v2"
)

// v1CronJobToV1Beta1CronJob does exactly what it says: converts a v1 cronjob to a v1beta1 cronjob.
func v1CronJobToV1Beta1CronJob(cj *batchv1.CronJob) *batchv1beta1.CronJob {
	return &batchv1beta1.CronJob{
		ObjectMeta: cj.ObjectMeta,
		Spec: batchv1beta1.CronJobSpec{
			Schedule:          cj.Spec.Schedule,
			ConcurrencyPolicy: batchv1beta1.ConcurrencyPolicy(cj.Spec.ConcurrencyPolicy),
			JobTemplate: batchv1beta1.JobTemplateSpec{
				Spec: cj.Spec.JobTemplate.Spec,
			},
		},
	}
}

// truncateAndHashString truncates the string to a max length and hashes the rest of it
// Ex: truncateAndHashString(aaaaaaaaaaaaaaaaaaaaaaaaaa,12) becomes "a-2067663226"
func truncateAndHashString(name string, length int) string {
	if len(name) > length {
		if length < 11 {
			return name[:length]
		}
		log.Warningf("Name %q exceeds %d characters (got %d)", name, length, len(name))
		// max length chars, minus 10 chars (the adler32 hash returns up to 10 digits), minus 1 for the '-'
		splitPoint := length - 11
		part1 := name[:splitPoint]
		part2 := name[splitPoint:]
		hashedPart2 := fmt.Sprint(adler32.Checksum([]byte(part2)))
		name = fmt.Sprintf("%s-%s", part1, hashedPart2)
	}
	return name
}

// belongsTo is similar to IsControlledBy, but enables us to establish a relationship
// between cronjobs and app repositories in different namespaces.
func objectBelongsTo(object, parent metav1.Object) bool {
	labels := object.GetLabels()
	return labels[LabelRepoName] == parent.GetName() && labels[LabelRepoNamespace] == parent.GetNamespace()
}

// intervalToCron transforms string durations like "1m" or "1h" to cron expressions
// Even if valid time units are "ns", "us", "ms", "s", "m", "h",
// the result will get rounded up to minutes.
// for durations over 24h, only durations below 1 year are supported
func intervalToCron(duration string) (string, error) {
	if duration == "" {
		return "", fmt.Errorf("duration cannot be empty")
	}

	d, err := time.ParseDuration(duration)
	if err != nil {
		return "", fmt.Errorf("error while parsing the duration: %s", err)
	}
	cronMins := math.Ceil(d.Minutes()) // round up to nearest minute

	if cronMins < 60 {
		// https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/#cron-schedule-syntax
		// minute(0-59) hour(0-23) dayOfMonth(1-31) month(1-12) dayOfWeek(0-6)
		return fmt.Sprintf("*/%v * * * *", cronMins), nil // every cronMins minutes
	}

	cronHours := math.Ceil(d.Hours()) // round up to nearest hour
	if cronHours < 24 {
		return fmt.Sprintf("0 */%v * * *", cronHours), nil // every cronHours hours
	}

	cronDays := math.Ceil(cronHours / 24) // get the days in cronHours, round up to nearest day
	if cronDays < 32 {
		return fmt.Sprintf("0 0 */%v * *", cronDays), nil // every cronDays days
	}

	cronMonths := math.Ceil(cronDays / 31) // get the months in cronDays, round up to nearest month
	if cronMonths < 13 {
		return fmt.Sprintf("0 0 1 */%v *", cronMonths), nil // every cronMonths months
	}

	return "", fmt.Errorf("not supported duration: %s", duration)

}

// the following pieces of code have been extracted from
// https://github.com/kubernetes-sigs/controller-runtime/blob/main/pkg/controller/controllerutil/controllerutil.go
// and modified to use the apimachinery object instead of the the controller-runtime object
// they are subject to the undermentioned license terms.

/*
Copyright 2018 The Kubernetes Authors.

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

// ContainsFinalizer checks an Object that the provided finalizer is present.
func containsFinalizer(o metav1.Object, finalizer string) bool {
	f := o.GetFinalizers()
	for _, e := range f {
		if e == finalizer {
			return true
		}
	}
	return false
}

// AddFinalizer accepts an Object and adds the provided finalizer if not present.
// It returns an indication of whether it updated the object's list of finalizers.
func addFinalizer(o metav1.Object, finalizer string) (finalizersUpdated bool) {
	f := o.GetFinalizers()
	for _, e := range f {
		if e == finalizer {
			return false
		}
	}
	o.SetFinalizers(append(f, finalizer))
	return true
}

// RemoveFinalizer accepts an Object and removes the provided finalizer if present.
// It returns an indication of whether it updated the object's list of finalizers.
func removeFinalizer(o metav1.Object, finalizer string) (finalizersUpdated bool) {
	f := o.GetFinalizers()
	for i := 0; i < len(f); i++ {
		if f[i] == finalizer {
			f = append(f[:i], f[i+1:]...)
			i--
			finalizersUpdated = true
		}
	}
	o.SetFinalizers(f)
	return
}
