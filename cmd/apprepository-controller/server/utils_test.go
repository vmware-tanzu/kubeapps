// Copyright 2021-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"testing"

	"github.com/adhocore/gronx"
	apprepov1alpha1 "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIntervalToCron(t *testing.T) {
	testCases := []struct {
		name         string
		interval     string
		expectedCron string
	}{
		{
			name:         "good interval, every 2 nanoseconds",
			interval:     "2ns",
			expectedCron: "*/1 * * * *",
		},
		{
			name:         "good interval, every 2 microseconds",
			interval:     "2us",
			expectedCron: "*/1 * * * *",
		},
		{
			name:         "good interval, every 2 milliseconds",
			interval:     "2ms",
			expectedCron: "*/1 * * * *",
		},
		{
			name:         "good interval, every 2 seconds",
			interval:     "2s",
			expectedCron: "*/1 * * * *",
		},
		{
			name:         "good interval, every two minutes",
			interval:     "2m",
			expectedCron: "*/2 * * * *",
		},
		{
			name:         "good interval, every two hours",
			interval:     "2h",
			expectedCron: "0 */2 * * *",
		},
		{
			name:         "good interval, every three days",
			interval:     "72h",
			expectedCron: "0 0 */3 * *",
		},
		{
			name:         "good interval, every 2 months",
			interval:     "1460h",
			expectedCron: "0 0 1 */2 *",
		},
		{
			name:         "bad interval, unsupported duration (> 1y)",
			interval:     "17532h",
			expectedCron: "",
		},
		{
			name:         "bad interval, unsupported unit (days)",
			interval:     "1d",
			expectedCron: "",
		},
	}

	gron := gronx.New()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cron, _ := intervalToCron(tc.interval)

			valid := gron.IsValid(cron)
			if cron != "" && !valid {
				t.Errorf("the generated cron is invalid: %s", cron)
			}

			if got, want := cron, tc.expectedCron; got != want {
				t.Errorf("got: %s, want: %s", got, want)
			}
		})
	}
}

func TestTruncateAndHashString(t *testing.T) {
	testCases := []struct {
		name   string
		input  string
		length int
		expect string
	}{
		{
			name:   "empty string",
			input:  "",
			length: 5,
			expect: "",
		},
		{
			name:   "0 length",
			input:  "",
			length: 0,
			expect: "",
		},
		{
			name:   "string under the max length",
			input:  "1234",
			length: 5,
			expect: "1234",
		},
		{
			name:   "string that fits in the max length",
			input:  "12345",
			length: 5,
			expect: "12345",
		},
		{
			name:   "long string whose exceeding part gets truncated but not hashed if length < 11",
			input:  "123456789",
			length: 5,
			expect: "12345",
		},
		{
			name:   "long string whose exceeding part gets truncated and hashed",
			input:  "1234567891234",
			length: 12,
			expect: "1-269222519",
		},
		{
			name:   "string under the 52-chars length",
			input:  "aaa",
			length: 52,
			expect: "aaa",
		},
		{
			name:   "string that fits in the 52-chars length",
			input:  "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			length: 52,
			expect: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		},
		{
			name:   "long string whose exceeding part gets truncated and hashed",
			input:  "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaExceedingLongName",
			length: 52,
			expect: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-2604272329",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := truncateAndHashString(tc.input, tc.length)
			if got, want := res, tc.expect; got != want {
				t.Errorf("got: %s, want: %s", got, want)
			}
		})
	}
}

func TestObjectBelongsTo(t *testing.T) {
	testCases := []struct {
		name   string
		object metav1.Object
		parent metav1.Object
		expect bool
	}{
		{
			name: "it recognises a cronjob belonging to an app repository in another namespace",
			object: &batchv1.CronJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "apprepo-kubeapps-sync-my-charts",
					Namespace: "kubeapps",
					Labels: map[string]string{
						LabelRepoName:      "my-charts",
						LabelRepoNamespace: "my-namespace",
					},
				},
			},
			parent: &apprepov1alpha1.AppRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-charts",
					Namespace: "my-namespace",
				},
			},
			expect: true,
		},
		{
			name: "it returns false if the namespace does not match",
			object: &batchv1.CronJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "apprepo-kubeapps-sync-my-charts",
					Namespace: "kubeapps",
					Labels: map[string]string{
						LabelRepoName:      "my-charts",
						LabelRepoNamespace: "my-namespace",
					},
				},
			},
			parent: &apprepov1alpha1.AppRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-charts",
					Namespace: "my-namespace2",
				},
			},
			expect: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got, want := objectBelongsTo(tc.object, tc.parent), tc.expect; got != want {
				t.Errorf("got: %t, want: %t", got, want)
			}
		})
	}
}
