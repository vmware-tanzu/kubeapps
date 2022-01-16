/*
Copyright 2021 VMware. All Rights Reserved.

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

package server

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	apprepov1alpha1 "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const repoSyncImage = "bitnami/kubeapps-asset-syncer:2.0.0-scratch-r2"

var defaultTTL = int32(3600)

func Test_newCronJob(t *testing.T) {
	tests := []struct {
		name             string
		crontab          string
		userAgentComment string
		apprepo          *apprepov1alpha1.AppRepository
		expected         batchv1beta1.CronJob
	}{
		{
			"my-charts",
			"*/10 * * * *",
			"",
			&apprepov1alpha1.AppRepository{
				TypeMeta: metav1.TypeMeta{
					Kind:       "AppRepository",
					APIVersion: "kubeapps.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-charts",
					Namespace: "kubeapps",
					Labels: map[string]string{
						"name":       "my-charts",
						"created-by": "kubeapps",
					},
				},
				Spec: apprepov1alpha1.AppRepositorySpec{
					Type: "helm",
					URL:  "https://charts.acme.com/my-charts",
				},
			},
			batchv1beta1.CronJob{
				ObjectMeta: metav1.ObjectMeta{
					Name: "apprepo-kubeapps-sync-my-charts",
					OwnerReferences: []metav1.OwnerReference{
						*metav1.NewControllerRef(
							&apprepov1alpha1.AppRepository{ObjectMeta: metav1.ObjectMeta{Name: "my-charts"}},
							schema.GroupVersionKind{
								Group:   apprepov1alpha1.SchemeGroupVersion.Group,
								Version: apprepov1alpha1.SchemeGroupVersion.Version,
								Kind:    "AppRepository",
							}),
					},
					Labels: map[string]string{
						LabelRepoName:      "my-charts",
						LabelRepoNamespace: "kubeapps",
					},
					Annotations: map[string]string{},
				},
				Spec: batchv1beta1.CronJobSpec{
					Schedule:          "*/10 * * * *",
					ConcurrencyPolicy: "Replace",
					JobTemplate: batchv1beta1.JobTemplateSpec{
						Spec: batchv1.JobSpec{
							TTLSecondsAfterFinished: &defaultTTL,
							Template: corev1.PodTemplateSpec{
								ObjectMeta: metav1.ObjectMeta{
									Labels: map[string]string{
										LabelRepoName:      "my-charts",
										LabelRepoNamespace: "kubeapps",
									},
									Annotations: map[string]string{},
								},
								Spec: corev1.PodSpec{
									RestartPolicy: "OnFailure",
									Containers: []corev1.Container{
										{
											Name:            "sync",
											Image:           repoSyncImage,
											ImagePullPolicy: "IfNotPresent",
											Command:         []string{"/chart-repo"},
											Args: []string{
												"sync",
												"--database-url=postgresql.kubeapps",
												"--database-user=admin",
												"--database-name=assets",
												"--global-repos-namespace=kubeapps-global",
												"--namespace=kubeapps",
												"my-charts",
												"https://charts.acme.com/my-charts",
												"helm",
											},
											Env: []corev1.EnvVar{
												{
													Name: "DB_PASSWORD",
													ValueFrom: &corev1.EnvVarSource{
														SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "postgresql"}, Key: "postgresql-root-password"}},
												},
											},
											VolumeMounts: nil,
										},
									},
									Volumes: nil,
								},
							},
						},
					},
				},
			},
		},
		{
			"my-charts with auth, userAgent and crontab configuration",
			"*/20 * * * *",
			"kubeapps/v2.3",
			&apprepov1alpha1.AppRepository{
				TypeMeta: metav1.TypeMeta{
					Kind:       "AppRepository",
					APIVersion: "kubeapps.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-charts",
					Namespace: "kubeapps",
					Labels: map[string]string{
						"name":       "my-charts",
						"created-by": "kubeapps",
					},
				},
				Spec: apprepov1alpha1.AppRepositorySpec{
					Type: "helm",
					URL:  "https://charts.acme.com/my-charts",
					Auth: apprepov1alpha1.AppRepositoryAuth{
						Header: &apprepov1alpha1.AppRepositoryAuthHeader{
							SecretKeyRef: corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "apprepo-my-charts-secrets"}, Key: "AuthorizationHeader"}},
					},
				},
			},
			batchv1beta1.CronJob{
				ObjectMeta: metav1.ObjectMeta{
					Name: "apprepo-kubeapps-sync-my-charts",
					OwnerReferences: []metav1.OwnerReference{
						*metav1.NewControllerRef(
							&apprepov1alpha1.AppRepository{ObjectMeta: metav1.ObjectMeta{Name: "my-charts"}},
							schema.GroupVersionKind{
								Group:   apprepov1alpha1.SchemeGroupVersion.Group,
								Version: apprepov1alpha1.SchemeGroupVersion.Version,
								Kind:    "AppRepository",
							}),
					},
					Labels: map[string]string{
						LabelRepoName:      "my-charts",
						LabelRepoNamespace: "kubeapps",
					},
					Annotations: map[string]string{},
				},
				Spec: batchv1beta1.CronJobSpec{
					Schedule:          "*/20 * * * *",
					ConcurrencyPolicy: "Replace",
					JobTemplate: batchv1beta1.JobTemplateSpec{
						Spec: batchv1.JobSpec{
							TTLSecondsAfterFinished: &defaultTTL,
							Template: corev1.PodTemplateSpec{
								ObjectMeta: metav1.ObjectMeta{
									Labels: map[string]string{
										LabelRepoName:      "my-charts",
										LabelRepoNamespace: "kubeapps",
									},
									Annotations: map[string]string{},
								},
								Spec: corev1.PodSpec{
									RestartPolicy: "OnFailure",
									Containers: []corev1.Container{
										{
											Name:            "sync",
											Image:           repoSyncImage,
											ImagePullPolicy: "IfNotPresent",
											Command:         []string{"/chart-repo"},
											Args: []string{
												"sync",
												"--database-url=postgresql.kubeapps",
												"--database-user=admin",
												"--database-name=assets",
												"--user-agent-comment=kubeapps/v2.3",
												"--global-repos-namespace=kubeapps-global",
												"--namespace=kubeapps",
												"my-charts",
												"https://charts.acme.com/my-charts",
												"helm",
											},
											Env: []corev1.EnvVar{
												{
													Name: "DB_PASSWORD",
													ValueFrom: &corev1.EnvVarSource{
														SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "postgresql"}, Key: "postgresql-root-password"}},
												},
												{
													Name: "AUTHORIZATION_HEADER",
													ValueFrom: &corev1.EnvVarSource{
														SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "apprepo-my-charts-secrets"}, Key: "AuthorizationHeader"}},
												},
											},
											VolumeMounts: nil,
										},
									},
									Volumes: nil,
								},
							},
						},
					},
				},
			},
		},
		{
			"a cronjob for an app repo in another namespace references the repo secret in kubeapps",
			"*/20 * * * *",
			"kubeapps/v2.3",
			&apprepov1alpha1.AppRepository{
				TypeMeta: metav1.TypeMeta{
					Kind:       "AppRepository",
					APIVersion: "kubeapps.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-charts-in-otherns",
					Namespace: "otherns",
					Labels: map[string]string{
						"name":       "my-charts",
						"created-by": "kubeapps",
					},
				},
				Spec: apprepov1alpha1.AppRepositorySpec{
					Type: "helm",
					URL:  "https://charts.acme.com/my-charts",
					Auth: apprepov1alpha1.AppRepositoryAuth{
						Header: &apprepov1alpha1.AppRepositoryAuthHeader{
							SecretKeyRef: corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "apprepo-my-charts-in-otherns"}, Key: "AuthorizationHeader"}},
					},
				},
			},
			batchv1beta1.CronJob{
				ObjectMeta: metav1.ObjectMeta{
					Name: "apprepo-otherns-sync-my-charts-in-otherns",
					Labels: map[string]string{
						LabelRepoName:      "my-charts-in-otherns",
						LabelRepoNamespace: "otherns",
					},
					Annotations: map[string]string{},
				},
				Spec: batchv1beta1.CronJobSpec{
					Schedule:          "*/20 * * * *",
					ConcurrencyPolicy: "Replace",
					JobTemplate: batchv1beta1.JobTemplateSpec{
						Spec: batchv1.JobSpec{
							TTLSecondsAfterFinished: &defaultTTL,
							Template: corev1.PodTemplateSpec{
								ObjectMeta: metav1.ObjectMeta{
									Labels: map[string]string{
										LabelRepoName:      "my-charts-in-otherns",
										LabelRepoNamespace: "otherns",
									},
									Annotations: map[string]string{},
								},
								Spec: corev1.PodSpec{
									RestartPolicy: "OnFailure",
									Containers: []corev1.Container{
										{
											Name:            "sync",
											Image:           repoSyncImage,
											ImagePullPolicy: "IfNotPresent",
											Command:         []string{"/chart-repo"},
											Args: []string{
												"sync",
												"--database-url=postgresql.kubeapps",
												"--database-user=admin",
												"--database-name=assets",
												"--user-agent-comment=kubeapps/v2.3",
												"--global-repos-namespace=kubeapps-global",
												"--namespace=otherns",
												"my-charts-in-otherns",
												"https://charts.acme.com/my-charts",
												"helm",
											},
											Env: []corev1.EnvVar{
												{
													Name: "DB_PASSWORD",
													ValueFrom: &corev1.EnvVarSource{
														SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "postgresql"}, Key: "postgresql-root-password"}},
												},
												{
													Name: "AUTHORIZATION_HEADER",
													ValueFrom: &corev1.EnvVarSource{
														SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "otherns-apprepo-my-charts-in-otherns"}, Key: "AuthorizationHeader"}},
												},
											},
											VolumeMounts: nil,
										},
									},
									Volumes: nil,
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := makeDefaultConfig()
			config.Crontab = tt.crontab
			config.UserAgentComment = tt.userAgentComment

			result := newCronJob(tt.apprepo, config)
			if got, want := tt.expected, *result; !cmp.Equal(want, got) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}

func Test_newSyncJob(t *testing.T) {
	tests := []struct {
		name             string
		userAgentComment string
		apprepo          *apprepov1alpha1.AppRepository
		expected         batchv1.Job
	}{
		{
			"my-charts",
			"",
			&apprepov1alpha1.AppRepository{
				TypeMeta: metav1.TypeMeta{
					Kind:       "AppRepository",
					APIVersion: "kubeapps.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-charts",
					Namespace: "kubeapps",
					Labels: map[string]string{
						"name":       "my-charts",
						"created-by": "kubeapps",
					},
				},
				Spec: apprepov1alpha1.AppRepositorySpec{
					Type: "helm",
					URL:  "https://charts.acme.com/my-charts",
				},
			},
			batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "apprepo-kubeapps-sync-my-charts-",
					OwnerReferences: []metav1.OwnerReference{
						*metav1.NewControllerRef(
							&apprepov1alpha1.AppRepository{ObjectMeta: metav1.ObjectMeta{Name: "my-charts"}},
							schema.GroupVersionKind{
								Group:   apprepov1alpha1.SchemeGroupVersion.Group,
								Version: apprepov1alpha1.SchemeGroupVersion.Version,
								Kind:    "AppRepository",
							},
						),
					},
					Annotations: map[string]string{},
					Labels:      map[string]string{},
				},
				Spec: batchv1.JobSpec{
					TTLSecondsAfterFinished: &defaultTTL,
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								LabelRepoName:      "my-charts",
								LabelRepoNamespace: "kubeapps",
							},
							Annotations: map[string]string{},
						},
						Spec: corev1.PodSpec{
							RestartPolicy: "OnFailure",
							Containers: []corev1.Container{
								{
									Name:            "sync",
									Image:           repoSyncImage,
									ImagePullPolicy: "IfNotPresent",
									Command:         []string{"/chart-repo"},
									Args: []string{
										"sync",
										"--database-url=postgresql.kubeapps",
										"--database-user=admin",
										"--database-name=assets",
										"--global-repos-namespace=kubeapps-global",
										"--namespace=kubeapps",
										"my-charts",
										"https://charts.acme.com/my-charts",
										"helm",
									},
									Env: []corev1.EnvVar{
										{
											Name: "DB_PASSWORD",
											ValueFrom: &corev1.EnvVarSource{
												SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "postgresql"}, Key: "postgresql-root-password"}},
										},
									},
									VolumeMounts: nil,
								},
							},
							Volumes: nil,
						},
					},
				},
			},
		},
		{
			"an app repository in another namespace results in jobs without owner references",
			"",
			&apprepov1alpha1.AppRepository{
				TypeMeta: metav1.TypeMeta{
					Kind:       "AppRepository",
					APIVersion: "kubeapps.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-charts",
					Namespace: "my-other-namespace",
					Labels: map[string]string{
						"name":       "my-charts",
						"created-by": "kubeapps",
					},
				},
				Spec: apprepov1alpha1.AppRepositorySpec{
					Type: "helm",
					URL:  "https://charts.acme.com/my-charts",
				},
			},
			batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "apprepo-my-other-namespace-sync-my-charts-",
					Annotations:  map[string]string{},
					Labels:       map[string]string{},
				},
				Spec: batchv1.JobSpec{
					TTLSecondsAfterFinished: &defaultTTL,
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								LabelRepoName:      "my-charts",
								LabelRepoNamespace: "my-other-namespace",
							},
							Annotations: map[string]string{},
						},
						Spec: corev1.PodSpec{
							RestartPolicy: "OnFailure",
							Containers: []corev1.Container{
								{
									Name:            "sync",
									Image:           repoSyncImage,
									ImagePullPolicy: "IfNotPresent",
									Command:         []string{"/chart-repo"},
									Args: []string{
										"sync",
										"--database-url=postgresql.kubeapps",
										"--database-user=admin",
										"--database-name=assets",
										"--global-repos-namespace=kubeapps-global",
										"--namespace=my-other-namespace",
										"my-charts",
										"https://charts.acme.com/my-charts",
										"helm",
									},
									Env: []corev1.EnvVar{
										{
											Name: "DB_PASSWORD",
											ValueFrom: &corev1.EnvVarSource{
												SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "postgresql"}, Key: "postgresql-root-password"}},
										},
									},
									VolumeMounts: nil,
								},
							},
							Volumes: nil,
						},
					},
				},
			},
		},
		{
			"my-charts with auth and userAgent comment",
			"kubeapps/v2.3",
			&apprepov1alpha1.AppRepository{
				TypeMeta: metav1.TypeMeta{
					Kind:       "AppRepository",
					APIVersion: "kubeapps.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-charts",
					Namespace: "kubeapps",
					Labels: map[string]string{
						"name":       "my-charts",
						"created-by": "kubeapps",
					},
				},
				Spec: apprepov1alpha1.AppRepositorySpec{
					Type: "helm",
					URL:  "https://charts.acme.com/my-charts",
					Auth: apprepov1alpha1.AppRepositoryAuth{
						Header: &apprepov1alpha1.AppRepositoryAuthHeader{
							SecretKeyRef: corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "apprepo-my-charts-secrets"}, Key: "AuthorizationHeader"}},
					},
				},
			},
			batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "apprepo-kubeapps-sync-my-charts-",
					OwnerReferences: []metav1.OwnerReference{
						*metav1.NewControllerRef(
							&apprepov1alpha1.AppRepository{ObjectMeta: metav1.ObjectMeta{Name: "my-charts"}},
							schema.GroupVersionKind{
								Group:   apprepov1alpha1.SchemeGroupVersion.Group,
								Version: apprepov1alpha1.SchemeGroupVersion.Version,
								Kind:    "AppRepository",
							},
						),
					},
					Annotations: map[string]string{},
					Labels:      map[string]string{},
				},
				Spec: batchv1.JobSpec{
					TTLSecondsAfterFinished: &defaultTTL,
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								LabelRepoName:      "my-charts",
								LabelRepoNamespace: "kubeapps",
							},
							Annotations: map[string]string{},
						},
						Spec: corev1.PodSpec{
							RestartPolicy: "OnFailure",
							Containers: []corev1.Container{
								{
									Name:            "sync",
									Image:           repoSyncImage,
									ImagePullPolicy: "IfNotPresent",
									Command:         []string{"/chart-repo"},
									Args: []string{
										"sync",
										"--database-url=postgresql.kubeapps",
										"--database-user=admin",
										"--database-name=assets",
										"--user-agent-comment=kubeapps/v2.3",
										"--global-repos-namespace=kubeapps-global",
										"--namespace=kubeapps",
										"my-charts",
										"https://charts.acme.com/my-charts",
										"helm",
									},
									Env: []corev1.EnvVar{
										{
											Name: "DB_PASSWORD",
											ValueFrom: &corev1.EnvVarSource{
												SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "postgresql"}, Key: "postgresql-root-password"}},
										},
										{
											Name: "AUTHORIZATION_HEADER",
											ValueFrom: &corev1.EnvVarSource{
												SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "apprepo-my-charts-secrets"}, Key: "AuthorizationHeader"}},
										},
									},
									VolumeMounts: nil,
								},
							},
							Volumes: nil,
						},
					},
				},
			},
		},
		{
			"my-charts with a customCA",
			"",
			&apprepov1alpha1.AppRepository{
				TypeMeta: metav1.TypeMeta{
					Kind:       "AppRepository",
					APIVersion: "kubeapps.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-charts",
					Namespace: "kubeapps",
					Labels: map[string]string{
						"name":       "my-charts",
						"created-by": "kubeapps",
					},
				},
				Spec: apprepov1alpha1.AppRepositorySpec{
					Type: "helm",
					URL:  "https://charts.acme.com/my-charts",
					Auth: apprepov1alpha1.AppRepositoryAuth{
						CustomCA: &apprepov1alpha1.AppRepositoryCustomCA{
							SecretKeyRef: corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "ca-cert-test"}, Key: "foo"},
						},
					},
				},
			},
			batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "apprepo-kubeapps-sync-my-charts-",
					OwnerReferences: []metav1.OwnerReference{
						*metav1.NewControllerRef(
							&apprepov1alpha1.AppRepository{ObjectMeta: metav1.ObjectMeta{Name: "my-charts"}},
							schema.GroupVersionKind{
								Group:   apprepov1alpha1.SchemeGroupVersion.Group,
								Version: apprepov1alpha1.SchemeGroupVersion.Version,
								Kind:    "AppRepository",
							},
						),
					},
					Annotations: map[string]string{},
					Labels:      map[string]string{},
				},
				Spec: batchv1.JobSpec{
					TTLSecondsAfterFinished: &defaultTTL,
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								LabelRepoName:      "my-charts",
								LabelRepoNamespace: "kubeapps",
							},
							Annotations: map[string]string{},
						},
						Spec: corev1.PodSpec{
							RestartPolicy: "OnFailure",
							Containers: []corev1.Container{
								{
									Name:            "sync",
									Image:           repoSyncImage,
									ImagePullPolicy: "IfNotPresent",
									Command:         []string{"/chart-repo"},
									Args: []string{
										"sync",
										"--database-url=postgresql.kubeapps",
										"--database-user=admin",
										"--database-name=assets",
										"--global-repos-namespace=kubeapps-global",
										"--namespace=kubeapps",
										"my-charts",
										"https://charts.acme.com/my-charts",
										"helm",
									},
									Env: []corev1.EnvVar{
										{
											Name: "DB_PASSWORD",
											ValueFrom: &corev1.EnvVarSource{
												SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "postgresql"}, Key: "postgresql-root-password"}},
										},
									},
									VolumeMounts: []corev1.VolumeMount{{
										Name:      "ca-cert-test",
										ReadOnly:  true,
										MountPath: "/usr/local/share/ca-certificates",
									}},
								},
							},
							Volumes: []corev1.Volume{{
								Name: "ca-cert-test",
								VolumeSource: corev1.VolumeSource{
									Secret: &corev1.SecretVolumeSource{
										SecretName: "ca-cert-test",
										Items: []corev1.KeyToPath{
											{Key: "foo", Path: "ca.crt"},
										},
									},
								},
							}},
						},
					},
				},
			},
		},
		{
			"my-charts with a customCA and auth header",
			"",
			&apprepov1alpha1.AppRepository{
				TypeMeta: metav1.TypeMeta{
					Kind:       "AppRepository",
					APIVersion: "kubeapps.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-charts",
					Namespace: "kubeapps",
					Labels: map[string]string{
						"name":       "my-charts",
						"created-by": "kubeapps",
					},
				},
				Spec: apprepov1alpha1.AppRepositorySpec{
					Type: "helm",
					URL:  "https://charts.acme.com/my-charts",
					Auth: apprepov1alpha1.AppRepositoryAuth{
						CustomCA: &apprepov1alpha1.AppRepositoryCustomCA{
							SecretKeyRef: corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "ca-cert-test"}, Key: "foo"},
						},
						Header: &apprepov1alpha1.AppRepositoryAuthHeader{
							SecretKeyRef: corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "apprepo-my-charts-secrets"}, Key: "AuthorizationHeader"},
						},
					},
				},
			},
			batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "apprepo-kubeapps-sync-my-charts-",
					OwnerReferences: []metav1.OwnerReference{
						*metav1.NewControllerRef(
							&apprepov1alpha1.AppRepository{ObjectMeta: metav1.ObjectMeta{Name: "my-charts"}},
							schema.GroupVersionKind{
								Group:   apprepov1alpha1.SchemeGroupVersion.Group,
								Version: apprepov1alpha1.SchemeGroupVersion.Version,
								Kind:    "AppRepository",
							},
						),
					},
					Annotations: map[string]string{},
					Labels:      map[string]string{},
				},
				Spec: batchv1.JobSpec{
					TTLSecondsAfterFinished: &defaultTTL,
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								LabelRepoName:      "my-charts",
								LabelRepoNamespace: "kubeapps",
							},
							Annotations: map[string]string{},
						},
						Spec: corev1.PodSpec{
							RestartPolicy: "OnFailure",
							Containers: []corev1.Container{
								{
									Name:            "sync",
									Image:           repoSyncImage,
									ImagePullPolicy: "IfNotPresent",
									Command:         []string{"/chart-repo"},
									Args: []string{
										"sync",
										"--database-url=postgresql.kubeapps",
										"--database-user=admin",
										"--database-name=assets",
										"--global-repos-namespace=kubeapps-global",
										"--namespace=kubeapps",
										"my-charts",
										"https://charts.acme.com/my-charts",
										"helm",
									},
									Env: []corev1.EnvVar{
										{
											Name: "DB_PASSWORD",
											ValueFrom: &corev1.EnvVarSource{
												SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "postgresql"}, Key: "postgresql-root-password"}},
										},
										{
											Name: "AUTHORIZATION_HEADER",
											ValueFrom: &corev1.EnvVarSource{
												SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "apprepo-my-charts-secrets"}, Key: "AuthorizationHeader"}},
										},
									},
									VolumeMounts: []corev1.VolumeMount{{
										Name:      "ca-cert-test",
										ReadOnly:  true,
										MountPath: "/usr/local/share/ca-certificates",
									}},
								},
							},
							Volumes: []corev1.Volume{{
								Name: "ca-cert-test",
								VolumeSource: corev1.VolumeSource{
									Secret: &corev1.SecretVolumeSource{
										SecretName: "ca-cert-test",
										Items: []corev1.KeyToPath{
											{Key: "foo", Path: "ca.crt"},
										},
									},
								},
							}},
						},
					},
				},
			},
		},
		{
			"my-charts linked to docker registry creds",
			"",
			&apprepov1alpha1.AppRepository{
				TypeMeta: metav1.TypeMeta{
					Kind:       "AppRepository",
					APIVersion: "kubeapps.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-charts",
					Namespace: "kubeapps",
					Labels: map[string]string{
						"name":       "my-charts",
						"created-by": "kubeapps",
					},
				},
				Spec: apprepov1alpha1.AppRepositorySpec{
					Type: "helm",
					URL:  "https://charts.acme.com/my-charts",
					Auth: apprepov1alpha1.AppRepositoryAuth{
						Header: &apprepov1alpha1.AppRepositoryAuthHeader{
							SecretKeyRef: corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "apprepo-my-charts-secrets"}, Key: ".dockerconfigjson"},
						},
					},
				},
			},
			batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "apprepo-kubeapps-sync-my-charts-",
					OwnerReferences: []metav1.OwnerReference{
						*metav1.NewControllerRef(
							&apprepov1alpha1.AppRepository{ObjectMeta: metav1.ObjectMeta{Name: "my-charts"}},
							schema.GroupVersionKind{
								Group:   apprepov1alpha1.SchemeGroupVersion.Group,
								Version: apprepov1alpha1.SchemeGroupVersion.Version,
								Kind:    "AppRepository",
							},
						),
					},
					Annotations: map[string]string{},
					Labels:      map[string]string{},
				},
				Spec: batchv1.JobSpec{
					TTLSecondsAfterFinished: &defaultTTL,
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								LabelRepoName:      "my-charts",
								LabelRepoNamespace: "kubeapps",
							},
							Annotations: map[string]string{},
						},
						Spec: corev1.PodSpec{
							RestartPolicy: "OnFailure",
							Containers: []corev1.Container{
								{
									Name:            "sync",
									Image:           repoSyncImage,
									ImagePullPolicy: "IfNotPresent",
									Command:         []string{"/chart-repo"},
									Args: []string{
										"sync",
										"--database-url=postgresql.kubeapps",
										"--database-user=admin",
										"--database-name=assets",
										"--global-repos-namespace=kubeapps-global",
										"--namespace=kubeapps",
										"my-charts",
										"https://charts.acme.com/my-charts",
										"helm",
									},
									Env: []corev1.EnvVar{
										{
											Name: "DB_PASSWORD",
											ValueFrom: &corev1.EnvVarSource{
												SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "postgresql"}, Key: "postgresql-root-password"}},
										},
										{
											Name: "DOCKER_CONFIG_JSON",
											ValueFrom: &corev1.EnvVarSource{
												SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "apprepo-my-charts-secrets"}, Key: ".dockerconfigjson"}},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			"my-charts with a custom pod template",
			"",
			&apprepov1alpha1.AppRepository{
				TypeMeta: metav1.TypeMeta{
					Kind:       "AppRepository",
					APIVersion: "kubeapps.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-charts",
					Namespace: "kubeapps",
					Labels: map[string]string{
						"name":       "my-charts",
						"created-by": "kubeapps",
					},
				},
				Spec: apprepov1alpha1.AppRepositorySpec{
					Type: "helm",
					URL:  "https://charts.acme.com/my-charts",
					SyncJobPodTemplate: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"foo": "bar",
							},
							Annotations: map[string]string{},
						},
						Spec: corev1.PodSpec{
							Affinity: &corev1.Affinity{NodeAffinity: &corev1.NodeAffinity{RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{}}},
							Containers: []corev1.Container{
								{
									Env: []corev1.EnvVar{
										{Name: "FOO", Value: "BAR"},
									},
									VolumeMounts: []corev1.VolumeMount{{Name: "foo", MountPath: "/bar"}},
								},
							},
							Volumes: []corev1.Volume{{Name: "foo", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}}},
						},
					},
				},
			},
			batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "apprepo-kubeapps-sync-my-charts-",
					OwnerReferences: []metav1.OwnerReference{
						*metav1.NewControllerRef(
							&apprepov1alpha1.AppRepository{ObjectMeta: metav1.ObjectMeta{Name: "my-charts"}},
							schema.GroupVersionKind{
								Group:   apprepov1alpha1.SchemeGroupVersion.Group,
								Version: apprepov1alpha1.SchemeGroupVersion.Version,
								Kind:    "AppRepository",
							},
						),
					},
					Annotations: map[string]string{},
					Labels:      map[string]string{},
				},
				Spec: batchv1.JobSpec{
					TTLSecondsAfterFinished: &defaultTTL,
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								LabelRepoName:      "my-charts",
								LabelRepoNamespace: "kubeapps",
								"foo":              "bar",
							},
							Annotations: map[string]string{},
						},
						Spec: corev1.PodSpec{
							Affinity:      &corev1.Affinity{NodeAffinity: &corev1.NodeAffinity{RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{}}},
							RestartPolicy: "OnFailure",
							Containers: []corev1.Container{
								{
									Name:            "sync",
									Image:           repoSyncImage,
									ImagePullPolicy: "IfNotPresent",
									Command:         []string{"/chart-repo"},
									Args: []string{
										"sync",
										"--database-url=postgresql.kubeapps",
										"--database-user=admin",
										"--database-name=assets",
										"--global-repos-namespace=kubeapps-global",
										"--namespace=kubeapps",
										"my-charts",
										"https://charts.acme.com/my-charts",
										"helm",
									},
									Env: []corev1.EnvVar{
										{Name: "FOO", Value: "BAR"},
										{
											Name: "DB_PASSWORD",
											ValueFrom: &corev1.EnvVarSource{
												SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "postgresql"}, Key: "postgresql-root-password"}},
										},
									},
									VolumeMounts: []corev1.VolumeMount{{Name: "foo", MountPath: "/bar"}},
								},
							},
							Volumes: []corev1.Volume{{Name: "foo", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}}},
						},
					},
				},
			},
		},
		{
			"OCI registry with repositories",
			"",
			&apprepov1alpha1.AppRepository{
				TypeMeta: metav1.TypeMeta{
					Kind:       "AppRepository",
					APIVersion: "kubeapps.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-charts",
					Namespace: "kubeapps",
					Labels: map[string]string{
						"name":       "my-charts",
						"created-by": "kubeapps",
					},
				},
				Spec: apprepov1alpha1.AppRepositorySpec{
					Type:            "oci",
					URL:             "https://charts.acme.com/my-charts",
					OCIRepositories: []string{"apache", "jenkins"},
				},
			},
			batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "apprepo-kubeapps-sync-my-charts-",
					OwnerReferences: []metav1.OwnerReference{
						*metav1.NewControllerRef(
							&apprepov1alpha1.AppRepository{ObjectMeta: metav1.ObjectMeta{Name: "my-charts"}},
							schema.GroupVersionKind{
								Group:   apprepov1alpha1.SchemeGroupVersion.Group,
								Version: apprepov1alpha1.SchemeGroupVersion.Version,
								Kind:    "AppRepository",
							},
						),
					},
					Annotations: map[string]string{},
					Labels:      map[string]string{},
				},
				Spec: batchv1.JobSpec{
					TTLSecondsAfterFinished: &defaultTTL,
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								LabelRepoName:      "my-charts",
								LabelRepoNamespace: "kubeapps",
							},
							Annotations: map[string]string{},
						},
						Spec: corev1.PodSpec{
							RestartPolicy: "OnFailure",
							Containers: []corev1.Container{
								{
									Name:            "sync",
									Image:           repoSyncImage,
									ImagePullPolicy: "IfNotPresent",
									Command:         []string{"/chart-repo"},
									Args: []string{
										"sync",
										"--database-url=postgresql.kubeapps",
										"--database-user=admin",
										"--database-name=assets",
										"--global-repos-namespace=kubeapps-global",
										"--namespace=kubeapps",
										"my-charts",
										"https://charts.acme.com/my-charts",
										"oci",
										"--oci-repositories",
										"apache,jenkins",
									},
									Env: []corev1.EnvVar{
										{
											Name: "DB_PASSWORD",
											ValueFrom: &corev1.EnvVarSource{
												SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "postgresql"}, Key: "postgresql-root-password"}},
										},
									},
									VolumeMounts: nil,
								},
							},
							Volumes: nil,
						},
					},
				},
			},
		},
		{
			"Skip TLS verification",
			"",
			&apprepov1alpha1.AppRepository{
				TypeMeta: metav1.TypeMeta{
					Kind:       "AppRepository",
					APIVersion: "kubeapps.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-charts",
					Namespace: "kubeapps",
					Labels: map[string]string{
						"name":       "my-charts",
						"created-by": "kubeapps",
					},
				},
				Spec: apprepov1alpha1.AppRepositorySpec{
					Type:                  "oci",
					URL:                   "https://charts.acme.com/my-charts",
					OCIRepositories:       []string{"apache", "jenkins"},
					TLSInsecureSkipVerify: true,
				},
			},
			batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "apprepo-kubeapps-sync-my-charts-",
					OwnerReferences: []metav1.OwnerReference{
						*metav1.NewControllerRef(
							&apprepov1alpha1.AppRepository{ObjectMeta: metav1.ObjectMeta{Name: "my-charts"}},
							schema.GroupVersionKind{
								Group:   apprepov1alpha1.SchemeGroupVersion.Group,
								Version: apprepov1alpha1.SchemeGroupVersion.Version,
								Kind:    "AppRepository",
							},
						),
					},
					Annotations: map[string]string{},
					Labels:      map[string]string{},
				},
				Spec: batchv1.JobSpec{
					TTLSecondsAfterFinished: &defaultTTL,
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								LabelRepoName:      "my-charts",
								LabelRepoNamespace: "kubeapps",
							},
							Annotations: map[string]string{},
						},
						Spec: corev1.PodSpec{
							RestartPolicy: "OnFailure",
							Containers: []corev1.Container{
								{
									Name:            "sync",
									Image:           repoSyncImage,
									ImagePullPolicy: "IfNotPresent",
									Command:         []string{"/chart-repo"},
									Args: []string{
										"sync",
										"--database-url=postgresql.kubeapps",
										"--database-user=admin",
										"--database-name=assets",
										"--global-repos-namespace=kubeapps-global",
										"--namespace=kubeapps",
										"my-charts",
										"https://charts.acme.com/my-charts",
										"oci",
										"--oci-repositories",
										"apache,jenkins",
										"--tls-insecure-skip-verify",
									},
									Env: []corev1.EnvVar{
										{
											Name: "DB_PASSWORD",
											ValueFrom: &corev1.EnvVarSource{
												SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "postgresql"}, Key: "postgresql-root-password"}},
										},
									},
									VolumeMounts: nil,
								},
							},
							Volumes: nil,
						},
					},
				},
			},
		},
		{
			"Paas credentials",
			"",
			&apprepov1alpha1.AppRepository{
				TypeMeta: metav1.TypeMeta{
					Kind:       "AppRepository",
					APIVersion: "kubeapps.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-charts",
					Namespace: "kubeapps",
					Labels: map[string]string{
						"name":       "my-charts",
						"created-by": "kubeapps",
					},
				},
				Spec: apprepov1alpha1.AppRepositorySpec{
					Type:            "oci",
					URL:             "https://charts.acme.com/my-charts",
					OCIRepositories: []string{"apache", "jenkins"},
					PassCredentials: true,
				},
			},
			batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "apprepo-kubeapps-sync-my-charts-",
					OwnerReferences: []metav1.OwnerReference{
						*metav1.NewControllerRef(
							&apprepov1alpha1.AppRepository{ObjectMeta: metav1.ObjectMeta{Name: "my-charts"}},
							schema.GroupVersionKind{
								Group:   apprepov1alpha1.SchemeGroupVersion.Group,
								Version: apprepov1alpha1.SchemeGroupVersion.Version,
								Kind:    "AppRepository",
							},
						),
					},
					Annotations: map[string]string{},
					Labels:      map[string]string{},
				},
				Spec: batchv1.JobSpec{
					TTLSecondsAfterFinished: &defaultTTL,
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								LabelRepoName:      "my-charts",
								LabelRepoNamespace: "kubeapps",
							},
							Annotations: map[string]string{},
						},
						Spec: corev1.PodSpec{
							RestartPolicy: "OnFailure",
							Containers: []corev1.Container{
								{
									Name:            "sync",
									Image:           repoSyncImage,
									ImagePullPolicy: "IfNotPresent",
									Command:         []string{"/chart-repo"},
									Args: []string{
										"sync",
										"--database-url=postgresql.kubeapps",
										"--database-user=admin",
										"--database-name=assets",
										"--global-repos-namespace=kubeapps-global",
										"--namespace=kubeapps",
										"my-charts",
										"https://charts.acme.com/my-charts",
										"oci",
										"--oci-repositories",
										"apache,jenkins",
										"--pass-credentials",
									},
									Env: []corev1.EnvVar{
										{
											Name: "DB_PASSWORD",
											ValueFrom: &corev1.EnvVarSource{
												SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "postgresql"}, Key: "postgresql-root-password"}},
										},
									},
									VolumeMounts: nil,
								},
							},
							Volumes: nil,
						},
					},
				},
			},
		},
		{
			"Repository with filters",
			"",
			&apprepov1alpha1.AppRepository{
				TypeMeta: metav1.TypeMeta{
					Kind:       "AppRepository",
					APIVersion: "kubeapps.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-charts",
					Namespace: "kubeapps",
					Labels: map[string]string{
						"name":       "my-charts",
						"created-by": "kubeapps",
					},
				},
				Spec: apprepov1alpha1.AppRepositorySpec{
					Type: "helm",
					URL:  "https://charts.acme.com/my-charts",
					FilterRule: apprepov1alpha1.FilterRuleSpec{
						JQ: ".name == $var1", Variables: map[string]string{"$var1": "wordpress"},
					},
				},
			},
			batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "apprepo-kubeapps-sync-my-charts-",
					OwnerReferences: []metav1.OwnerReference{
						*metav1.NewControllerRef(
							&apprepov1alpha1.AppRepository{ObjectMeta: metav1.ObjectMeta{Name: "my-charts"}},
							schema.GroupVersionKind{
								Group:   apprepov1alpha1.SchemeGroupVersion.Group,
								Version: apprepov1alpha1.SchemeGroupVersion.Version,
								Kind:    "AppRepository",
							},
						),
					},
					Annotations: map[string]string{},
					Labels:      map[string]string{},
				},
				Spec: batchv1.JobSpec{
					TTLSecondsAfterFinished: &defaultTTL,
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								LabelRepoName:      "my-charts",
								LabelRepoNamespace: "kubeapps",
							},
							Annotations: map[string]string{},
						},
						Spec: corev1.PodSpec{
							RestartPolicy: "OnFailure",
							Containers: []corev1.Container{
								{
									Name:            "sync",
									Image:           repoSyncImage,
									ImagePullPolicy: "IfNotPresent",
									Command:         []string{"/chart-repo"},
									Args: []string{
										"sync",
										"--database-url=postgresql.kubeapps",
										"--database-user=admin",
										"--database-name=assets",
										"--global-repos-namespace=kubeapps-global",
										"--namespace=kubeapps",
										"my-charts",
										"https://charts.acme.com/my-charts",
										"helm",
										"--filter-rules",
										`{"jq":".name == $var1","variables":{"$var1":"wordpress"}}`,
									},
									Env: []corev1.EnvVar{
										{
											Name: "DB_PASSWORD",
											ValueFrom: &corev1.EnvVarSource{
												SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "postgresql"}, Key: "postgresql-root-password"}},
										},
									},
									VolumeMounts: nil,
								},
							},
							Volumes: nil,
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := makeDefaultConfig()
			config.UserAgentComment = tt.userAgentComment

			result := newSyncJob(tt.apprepo, config)
			if got, want := tt.expected, *result; !cmp.Equal(want, got) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}

func Test_newCleanupJob(t *testing.T) {
	tests := []struct {
		name          string
		repoName      string
		repoNamespace string
		expected      batchv1.Job
	}{
		{
			"my-charts",
			"my-charts",
			"kubeapps",
			batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "apprepo-kubeapps-cleanup-my-charts-",
					Namespace:    "kubeapps",
					Annotations:  map[string]string{},
					Labels:       map[string]string{},
				},
				Spec: batchv1.JobSpec{
					TTLSecondsAfterFinished: &defaultTTL,
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							RestartPolicy: "Never",
							Containers: []corev1.Container{
								{
									Name:            "delete",
									Image:           repoSyncImage,
									ImagePullPolicy: "IfNotPresent",
									Command:         []string{"/chart-repo"},
									Args: []string{
										"delete",
										"my-charts",
										"--namespace=kubeapps",
										"--database-url=postgresql.kubeapps",
										"--database-user=admin",
										"--database-name=assets",
									},
									Env: []corev1.EnvVar{
										{
											Name: "DB_PASSWORD",
											ValueFrom: &corev1.EnvVarSource{
												SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "postgresql"}, Key: "postgresql-root-password"}},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := newCleanupJob("kubeapps", tt.repoNamespace, tt.repoName, makeDefaultConfig())
			if got, want := tt.expected, *result; !cmp.Equal(want, got) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
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
			object: &batchv1beta1.CronJob{
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
			object: &batchv1beta1.CronJob{
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

func makeDefaultConfig() Config {
	return Config{
		Kubeconfig:               "",
		MasterURL:                "",
		RepoSyncImage:            repoSyncImage,
		RepoSyncImagePullSecrets: []string{},
		RepoSyncCommand:          "/chart-repo",
		KubeappsNamespace:        "kubeapps",
		GlobalReposNamespace:     "kubeapps-global",
		ReposPerNamespace:        true,
		DBURL:                    "postgresql.kubeapps",
		DBUser:                   "admin",
		DBName:                   "assets",
		DBSecretName:             "postgresql",
		DBSecretKey:              "postgresql-root-password",
		UserAgentComment:         "",
		TTLSecondsAfterFinished:  "3600",
		Crontab:                  "*/10 * * * *",
		CustomAnnotations:        []string{},
		CustomLabels:             []string{},
		ParsedCustomLabels:       map[string]string{},
		ParsedCustomAnnotations:  map[string]string{},
	}
}
