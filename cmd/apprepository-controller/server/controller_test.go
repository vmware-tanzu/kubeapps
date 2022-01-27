// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"testing"

	cmp "github.com/google/go-cmp/cmp"
	apprepov1alpha1 "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	k8sbatchv1 "k8s.io/api/batch/v1"
	k8sbatchv1beta1 "k8s.io/api/batch/v1beta1"
	k8scorev1 "k8s.io/api/core/v1"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sschema "k8s.io/apimachinery/pkg/runtime/schema"
)

const repoSyncImage = "bitnami/kubeapps-asset-syncer:2.0.0-scratch-r2"

var defaultTTL = int32(3600)

func Test_newCronJob(t *testing.T) {
	tests := []struct {
		name             string
		crontab          string
		userAgentComment string
		apprepo          *apprepov1alpha1.AppRepository
		expected         k8sbatchv1beta1.CronJob
	}{
		{
			"my-charts",
			"*/10 * * * *",
			"",
			&apprepov1alpha1.AppRepository{
				TypeMeta: k8smetav1.TypeMeta{
					Kind:       "AppRepository",
					APIVersion: "kubeapps.com/v1alpha1",
				},
				ObjectMeta: k8smetav1.ObjectMeta{
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
			k8sbatchv1beta1.CronJob{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name: "apprepo-kubeapps-sync-my-charts",
					OwnerReferences: []k8smetav1.OwnerReference{
						*k8smetav1.NewControllerRef(
							&apprepov1alpha1.AppRepository{ObjectMeta: k8smetav1.ObjectMeta{Name: "my-charts"}},
							k8sschema.GroupVersionKind{
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
				Spec: k8sbatchv1beta1.CronJobSpec{
					Schedule:          "*/10 * * * *",
					ConcurrencyPolicy: "Replace",
					JobTemplate: k8sbatchv1beta1.JobTemplateSpec{
						Spec: k8sbatchv1.JobSpec{
							TTLSecondsAfterFinished: &defaultTTL,
							Template: k8scorev1.PodTemplateSpec{
								ObjectMeta: k8smetav1.ObjectMeta{
									Labels: map[string]string{
										LabelRepoName:      "my-charts",
										LabelRepoNamespace: "kubeapps",
									},
									Annotations: map[string]string{},
								},
								Spec: k8scorev1.PodSpec{
									RestartPolicy: "OnFailure",
									Containers: []k8scorev1.Container{
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
											Env: []k8scorev1.EnvVar{
												{
													Name: "DB_PASSWORD",
													ValueFrom: &k8scorev1.EnvVarSource{
														SecretKeyRef: &k8scorev1.SecretKeySelector{LocalObjectReference: k8scorev1.LocalObjectReference{Name: "postgresql"}, Key: "postgresql-root-password"}},
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
				TypeMeta: k8smetav1.TypeMeta{
					Kind:       "AppRepository",
					APIVersion: "kubeapps.com/v1alpha1",
				},
				ObjectMeta: k8smetav1.ObjectMeta{
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
							SecretKeyRef: k8scorev1.SecretKeySelector{LocalObjectReference: k8scorev1.LocalObjectReference{Name: "apprepo-my-charts-secrets"}, Key: "AuthorizationHeader"}},
					},
				},
			},
			k8sbatchv1beta1.CronJob{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name: "apprepo-kubeapps-sync-my-charts",
					OwnerReferences: []k8smetav1.OwnerReference{
						*k8smetav1.NewControllerRef(
							&apprepov1alpha1.AppRepository{ObjectMeta: k8smetav1.ObjectMeta{Name: "my-charts"}},
							k8sschema.GroupVersionKind{
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
				Spec: k8sbatchv1beta1.CronJobSpec{
					Schedule:          "*/20 * * * *",
					ConcurrencyPolicy: "Replace",
					JobTemplate: k8sbatchv1beta1.JobTemplateSpec{
						Spec: k8sbatchv1.JobSpec{
							TTLSecondsAfterFinished: &defaultTTL,
							Template: k8scorev1.PodTemplateSpec{
								ObjectMeta: k8smetav1.ObjectMeta{
									Labels: map[string]string{
										LabelRepoName:      "my-charts",
										LabelRepoNamespace: "kubeapps",
									},
									Annotations: map[string]string{},
								},
								Spec: k8scorev1.PodSpec{
									RestartPolicy: "OnFailure",
									Containers: []k8scorev1.Container{
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
											Env: []k8scorev1.EnvVar{
												{
													Name: "DB_PASSWORD",
													ValueFrom: &k8scorev1.EnvVarSource{
														SecretKeyRef: &k8scorev1.SecretKeySelector{LocalObjectReference: k8scorev1.LocalObjectReference{Name: "postgresql"}, Key: "postgresql-root-password"}},
												},
												{
													Name: "AUTHORIZATION_HEADER",
													ValueFrom: &k8scorev1.EnvVarSource{
														SecretKeyRef: &k8scorev1.SecretKeySelector{LocalObjectReference: k8scorev1.LocalObjectReference{Name: "apprepo-my-charts-secrets"}, Key: "AuthorizationHeader"}},
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
				TypeMeta: k8smetav1.TypeMeta{
					Kind:       "AppRepository",
					APIVersion: "kubeapps.com/v1alpha1",
				},
				ObjectMeta: k8smetav1.ObjectMeta{
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
							SecretKeyRef: k8scorev1.SecretKeySelector{LocalObjectReference: k8scorev1.LocalObjectReference{Name: "apprepo-my-charts-in-otherns"}, Key: "AuthorizationHeader"}},
					},
				},
			},
			k8sbatchv1beta1.CronJob{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name: "apprepo-otherns-sync-my-charts-in-otherns",
					Labels: map[string]string{
						LabelRepoName:      "my-charts-in-otherns",
						LabelRepoNamespace: "otherns",
					},
					Annotations: map[string]string{},
				},
				Spec: k8sbatchv1beta1.CronJobSpec{
					Schedule:          "*/20 * * * *",
					ConcurrencyPolicy: "Replace",
					JobTemplate: k8sbatchv1beta1.JobTemplateSpec{
						Spec: k8sbatchv1.JobSpec{
							TTLSecondsAfterFinished: &defaultTTL,
							Template: k8scorev1.PodTemplateSpec{
								ObjectMeta: k8smetav1.ObjectMeta{
									Labels: map[string]string{
										LabelRepoName:      "my-charts-in-otherns",
										LabelRepoNamespace: "otherns",
									},
									Annotations: map[string]string{},
								},
								Spec: k8scorev1.PodSpec{
									RestartPolicy: "OnFailure",
									Containers: []k8scorev1.Container{
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
											Env: []k8scorev1.EnvVar{
												{
													Name: "DB_PASSWORD",
													ValueFrom: &k8scorev1.EnvVarSource{
														SecretKeyRef: &k8scorev1.SecretKeySelector{LocalObjectReference: k8scorev1.LocalObjectReference{Name: "postgresql"}, Key: "postgresql-root-password"}},
												},
												{
													Name: "AUTHORIZATION_HEADER",
													ValueFrom: &k8scorev1.EnvVarSource{
														SecretKeyRef: &k8scorev1.SecretKeySelector{LocalObjectReference: k8scorev1.LocalObjectReference{Name: "otherns-apprepo-my-charts-in-otherns"}, Key: "AuthorizationHeader"}},
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
		expected         k8sbatchv1.Job
	}{
		{
			"my-charts",
			"",
			&apprepov1alpha1.AppRepository{
				TypeMeta: k8smetav1.TypeMeta{
					Kind:       "AppRepository",
					APIVersion: "kubeapps.com/v1alpha1",
				},
				ObjectMeta: k8smetav1.ObjectMeta{
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
			k8sbatchv1.Job{
				ObjectMeta: k8smetav1.ObjectMeta{
					GenerateName: "apprepo-kubeapps-sync-my-charts-",
					OwnerReferences: []k8smetav1.OwnerReference{
						*k8smetav1.NewControllerRef(
							&apprepov1alpha1.AppRepository{ObjectMeta: k8smetav1.ObjectMeta{Name: "my-charts"}},
							k8sschema.GroupVersionKind{
								Group:   apprepov1alpha1.SchemeGroupVersion.Group,
								Version: apprepov1alpha1.SchemeGroupVersion.Version,
								Kind:    "AppRepository",
							},
						),
					},
					Annotations: map[string]string{},
					Labels:      map[string]string{},
				},
				Spec: k8sbatchv1.JobSpec{
					TTLSecondsAfterFinished: &defaultTTL,
					Template: k8scorev1.PodTemplateSpec{
						ObjectMeta: k8smetav1.ObjectMeta{
							Labels: map[string]string{
								LabelRepoName:      "my-charts",
								LabelRepoNamespace: "kubeapps",
							},
							Annotations: map[string]string{},
						},
						Spec: k8scorev1.PodSpec{
							RestartPolicy: "OnFailure",
							Containers: []k8scorev1.Container{
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
									Env: []k8scorev1.EnvVar{
										{
											Name: "DB_PASSWORD",
											ValueFrom: &k8scorev1.EnvVarSource{
												SecretKeyRef: &k8scorev1.SecretKeySelector{LocalObjectReference: k8scorev1.LocalObjectReference{Name: "postgresql"}, Key: "postgresql-root-password"}},
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
				TypeMeta: k8smetav1.TypeMeta{
					Kind:       "AppRepository",
					APIVersion: "kubeapps.com/v1alpha1",
				},
				ObjectMeta: k8smetav1.ObjectMeta{
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
			k8sbatchv1.Job{
				ObjectMeta: k8smetav1.ObjectMeta{
					GenerateName: "apprepo-my-other-namespace-sync-my-charts-",
					Annotations:  map[string]string{},
					Labels:       map[string]string{},
				},
				Spec: k8sbatchv1.JobSpec{
					TTLSecondsAfterFinished: &defaultTTL,
					Template: k8scorev1.PodTemplateSpec{
						ObjectMeta: k8smetav1.ObjectMeta{
							Labels: map[string]string{
								LabelRepoName:      "my-charts",
								LabelRepoNamespace: "my-other-namespace",
							},
							Annotations: map[string]string{},
						},
						Spec: k8scorev1.PodSpec{
							RestartPolicy: "OnFailure",
							Containers: []k8scorev1.Container{
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
									Env: []k8scorev1.EnvVar{
										{
											Name: "DB_PASSWORD",
											ValueFrom: &k8scorev1.EnvVarSource{
												SecretKeyRef: &k8scorev1.SecretKeySelector{LocalObjectReference: k8scorev1.LocalObjectReference{Name: "postgresql"}, Key: "postgresql-root-password"}},
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
				TypeMeta: k8smetav1.TypeMeta{
					Kind:       "AppRepository",
					APIVersion: "kubeapps.com/v1alpha1",
				},
				ObjectMeta: k8smetav1.ObjectMeta{
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
							SecretKeyRef: k8scorev1.SecretKeySelector{LocalObjectReference: k8scorev1.LocalObjectReference{Name: "apprepo-my-charts-secrets"}, Key: "AuthorizationHeader"}},
					},
				},
			},
			k8sbatchv1.Job{
				ObjectMeta: k8smetav1.ObjectMeta{
					GenerateName: "apprepo-kubeapps-sync-my-charts-",
					OwnerReferences: []k8smetav1.OwnerReference{
						*k8smetav1.NewControllerRef(
							&apprepov1alpha1.AppRepository{ObjectMeta: k8smetav1.ObjectMeta{Name: "my-charts"}},
							k8sschema.GroupVersionKind{
								Group:   apprepov1alpha1.SchemeGroupVersion.Group,
								Version: apprepov1alpha1.SchemeGroupVersion.Version,
								Kind:    "AppRepository",
							},
						),
					},
					Annotations: map[string]string{},
					Labels:      map[string]string{},
				},
				Spec: k8sbatchv1.JobSpec{
					TTLSecondsAfterFinished: &defaultTTL,
					Template: k8scorev1.PodTemplateSpec{
						ObjectMeta: k8smetav1.ObjectMeta{
							Labels: map[string]string{
								LabelRepoName:      "my-charts",
								LabelRepoNamespace: "kubeapps",
							},
							Annotations: map[string]string{},
						},
						Spec: k8scorev1.PodSpec{
							RestartPolicy: "OnFailure",
							Containers: []k8scorev1.Container{
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
									Env: []k8scorev1.EnvVar{
										{
											Name: "DB_PASSWORD",
											ValueFrom: &k8scorev1.EnvVarSource{
												SecretKeyRef: &k8scorev1.SecretKeySelector{LocalObjectReference: k8scorev1.LocalObjectReference{Name: "postgresql"}, Key: "postgresql-root-password"}},
										},
										{
											Name: "AUTHORIZATION_HEADER",
											ValueFrom: &k8scorev1.EnvVarSource{
												SecretKeyRef: &k8scorev1.SecretKeySelector{LocalObjectReference: k8scorev1.LocalObjectReference{Name: "apprepo-my-charts-secrets"}, Key: "AuthorizationHeader"}},
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
				TypeMeta: k8smetav1.TypeMeta{
					Kind:       "AppRepository",
					APIVersion: "kubeapps.com/v1alpha1",
				},
				ObjectMeta: k8smetav1.ObjectMeta{
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
							SecretKeyRef: k8scorev1.SecretKeySelector{LocalObjectReference: k8scorev1.LocalObjectReference{Name: "ca-cert-test"}, Key: "foo"},
						},
					},
				},
			},
			k8sbatchv1.Job{
				ObjectMeta: k8smetav1.ObjectMeta{
					GenerateName: "apprepo-kubeapps-sync-my-charts-",
					OwnerReferences: []k8smetav1.OwnerReference{
						*k8smetav1.NewControllerRef(
							&apprepov1alpha1.AppRepository{ObjectMeta: k8smetav1.ObjectMeta{Name: "my-charts"}},
							k8sschema.GroupVersionKind{
								Group:   apprepov1alpha1.SchemeGroupVersion.Group,
								Version: apprepov1alpha1.SchemeGroupVersion.Version,
								Kind:    "AppRepository",
							},
						),
					},
					Annotations: map[string]string{},
					Labels:      map[string]string{},
				},
				Spec: k8sbatchv1.JobSpec{
					TTLSecondsAfterFinished: &defaultTTL,
					Template: k8scorev1.PodTemplateSpec{
						ObjectMeta: k8smetav1.ObjectMeta{
							Labels: map[string]string{
								LabelRepoName:      "my-charts",
								LabelRepoNamespace: "kubeapps",
							},
							Annotations: map[string]string{},
						},
						Spec: k8scorev1.PodSpec{
							RestartPolicy: "OnFailure",
							Containers: []k8scorev1.Container{
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
									Env: []k8scorev1.EnvVar{
										{
											Name: "DB_PASSWORD",
											ValueFrom: &k8scorev1.EnvVarSource{
												SecretKeyRef: &k8scorev1.SecretKeySelector{LocalObjectReference: k8scorev1.LocalObjectReference{Name: "postgresql"}, Key: "postgresql-root-password"}},
										},
									},
									VolumeMounts: []k8scorev1.VolumeMount{{
										Name:      "ca-cert-test",
										ReadOnly:  true,
										MountPath: "/usr/local/share/ca-certificates",
									}},
								},
							},
							Volumes: []k8scorev1.Volume{{
								Name: "ca-cert-test",
								VolumeSource: k8scorev1.VolumeSource{
									Secret: &k8scorev1.SecretVolumeSource{
										SecretName: "ca-cert-test",
										Items: []k8scorev1.KeyToPath{
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
				TypeMeta: k8smetav1.TypeMeta{
					Kind:       "AppRepository",
					APIVersion: "kubeapps.com/v1alpha1",
				},
				ObjectMeta: k8smetav1.ObjectMeta{
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
							SecretKeyRef: k8scorev1.SecretKeySelector{LocalObjectReference: k8scorev1.LocalObjectReference{Name: "ca-cert-test"}, Key: "foo"},
						},
						Header: &apprepov1alpha1.AppRepositoryAuthHeader{
							SecretKeyRef: k8scorev1.SecretKeySelector{LocalObjectReference: k8scorev1.LocalObjectReference{Name: "apprepo-my-charts-secrets"}, Key: "AuthorizationHeader"},
						},
					},
				},
			},
			k8sbatchv1.Job{
				ObjectMeta: k8smetav1.ObjectMeta{
					GenerateName: "apprepo-kubeapps-sync-my-charts-",
					OwnerReferences: []k8smetav1.OwnerReference{
						*k8smetav1.NewControllerRef(
							&apprepov1alpha1.AppRepository{ObjectMeta: k8smetav1.ObjectMeta{Name: "my-charts"}},
							k8sschema.GroupVersionKind{
								Group:   apprepov1alpha1.SchemeGroupVersion.Group,
								Version: apprepov1alpha1.SchemeGroupVersion.Version,
								Kind:    "AppRepository",
							},
						),
					},
					Annotations: map[string]string{},
					Labels:      map[string]string{},
				},
				Spec: k8sbatchv1.JobSpec{
					TTLSecondsAfterFinished: &defaultTTL,
					Template: k8scorev1.PodTemplateSpec{
						ObjectMeta: k8smetav1.ObjectMeta{
							Labels: map[string]string{
								LabelRepoName:      "my-charts",
								LabelRepoNamespace: "kubeapps",
							},
							Annotations: map[string]string{},
						},
						Spec: k8scorev1.PodSpec{
							RestartPolicy: "OnFailure",
							Containers: []k8scorev1.Container{
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
									Env: []k8scorev1.EnvVar{
										{
											Name: "DB_PASSWORD",
											ValueFrom: &k8scorev1.EnvVarSource{
												SecretKeyRef: &k8scorev1.SecretKeySelector{LocalObjectReference: k8scorev1.LocalObjectReference{Name: "postgresql"}, Key: "postgresql-root-password"}},
										},
										{
											Name: "AUTHORIZATION_HEADER",
											ValueFrom: &k8scorev1.EnvVarSource{
												SecretKeyRef: &k8scorev1.SecretKeySelector{LocalObjectReference: k8scorev1.LocalObjectReference{Name: "apprepo-my-charts-secrets"}, Key: "AuthorizationHeader"}},
										},
									},
									VolumeMounts: []k8scorev1.VolumeMount{{
										Name:      "ca-cert-test",
										ReadOnly:  true,
										MountPath: "/usr/local/share/ca-certificates",
									}},
								},
							},
							Volumes: []k8scorev1.Volume{{
								Name: "ca-cert-test",
								VolumeSource: k8scorev1.VolumeSource{
									Secret: &k8scorev1.SecretVolumeSource{
										SecretName: "ca-cert-test",
										Items: []k8scorev1.KeyToPath{
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
				TypeMeta: k8smetav1.TypeMeta{
					Kind:       "AppRepository",
					APIVersion: "kubeapps.com/v1alpha1",
				},
				ObjectMeta: k8smetav1.ObjectMeta{
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
							SecretKeyRef: k8scorev1.SecretKeySelector{LocalObjectReference: k8scorev1.LocalObjectReference{Name: "apprepo-my-charts-secrets"}, Key: ".dockerconfigjson"},
						},
					},
				},
			},
			k8sbatchv1.Job{
				ObjectMeta: k8smetav1.ObjectMeta{
					GenerateName: "apprepo-kubeapps-sync-my-charts-",
					OwnerReferences: []k8smetav1.OwnerReference{
						*k8smetav1.NewControllerRef(
							&apprepov1alpha1.AppRepository{ObjectMeta: k8smetav1.ObjectMeta{Name: "my-charts"}},
							k8sschema.GroupVersionKind{
								Group:   apprepov1alpha1.SchemeGroupVersion.Group,
								Version: apprepov1alpha1.SchemeGroupVersion.Version,
								Kind:    "AppRepository",
							},
						),
					},
					Annotations: map[string]string{},
					Labels:      map[string]string{},
				},
				Spec: k8sbatchv1.JobSpec{
					TTLSecondsAfterFinished: &defaultTTL,
					Template: k8scorev1.PodTemplateSpec{
						ObjectMeta: k8smetav1.ObjectMeta{
							Labels: map[string]string{
								LabelRepoName:      "my-charts",
								LabelRepoNamespace: "kubeapps",
							},
							Annotations: map[string]string{},
						},
						Spec: k8scorev1.PodSpec{
							RestartPolicy: "OnFailure",
							Containers: []k8scorev1.Container{
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
									Env: []k8scorev1.EnvVar{
										{
											Name: "DB_PASSWORD",
											ValueFrom: &k8scorev1.EnvVarSource{
												SecretKeyRef: &k8scorev1.SecretKeySelector{LocalObjectReference: k8scorev1.LocalObjectReference{Name: "postgresql"}, Key: "postgresql-root-password"}},
										},
										{
											Name: "DOCKER_CONFIG_JSON",
											ValueFrom: &k8scorev1.EnvVarSource{
												SecretKeyRef: &k8scorev1.SecretKeySelector{LocalObjectReference: k8scorev1.LocalObjectReference{Name: "apprepo-my-charts-secrets"}, Key: ".dockerconfigjson"}},
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
				TypeMeta: k8smetav1.TypeMeta{
					Kind:       "AppRepository",
					APIVersion: "kubeapps.com/v1alpha1",
				},
				ObjectMeta: k8smetav1.ObjectMeta{
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
					SyncJobPodTemplate: k8scorev1.PodTemplateSpec{
						ObjectMeta: k8smetav1.ObjectMeta{
							Labels: map[string]string{
								"foo": "bar",
							},
							Annotations: map[string]string{},
						},
						Spec: k8scorev1.PodSpec{
							Affinity: &k8scorev1.Affinity{NodeAffinity: &k8scorev1.NodeAffinity{RequiredDuringSchedulingIgnoredDuringExecution: &k8scorev1.NodeSelector{}}},
							Containers: []k8scorev1.Container{
								{
									Env: []k8scorev1.EnvVar{
										{Name: "FOO", Value: "BAR"},
									},
									VolumeMounts: []k8scorev1.VolumeMount{{Name: "foo", MountPath: "/bar"}},
								},
							},
							Volumes: []k8scorev1.Volume{{Name: "foo", VolumeSource: k8scorev1.VolumeSource{EmptyDir: &k8scorev1.EmptyDirVolumeSource{}}}},
						},
					},
				},
			},
			k8sbatchv1.Job{
				ObjectMeta: k8smetav1.ObjectMeta{
					GenerateName: "apprepo-kubeapps-sync-my-charts-",
					OwnerReferences: []k8smetav1.OwnerReference{
						*k8smetav1.NewControllerRef(
							&apprepov1alpha1.AppRepository{ObjectMeta: k8smetav1.ObjectMeta{Name: "my-charts"}},
							k8sschema.GroupVersionKind{
								Group:   apprepov1alpha1.SchemeGroupVersion.Group,
								Version: apprepov1alpha1.SchemeGroupVersion.Version,
								Kind:    "AppRepository",
							},
						),
					},
					Annotations: map[string]string{},
					Labels:      map[string]string{},
				},
				Spec: k8sbatchv1.JobSpec{
					TTLSecondsAfterFinished: &defaultTTL,
					Template: k8scorev1.PodTemplateSpec{
						ObjectMeta: k8smetav1.ObjectMeta{
							Labels: map[string]string{
								LabelRepoName:      "my-charts",
								LabelRepoNamespace: "kubeapps",
								"foo":              "bar",
							},
							Annotations: map[string]string{},
						},
						Spec: k8scorev1.PodSpec{
							Affinity:      &k8scorev1.Affinity{NodeAffinity: &k8scorev1.NodeAffinity{RequiredDuringSchedulingIgnoredDuringExecution: &k8scorev1.NodeSelector{}}},
							RestartPolicy: "OnFailure",
							Containers: []k8scorev1.Container{
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
									Env: []k8scorev1.EnvVar{
										{Name: "FOO", Value: "BAR"},
										{
											Name: "DB_PASSWORD",
											ValueFrom: &k8scorev1.EnvVarSource{
												SecretKeyRef: &k8scorev1.SecretKeySelector{LocalObjectReference: k8scorev1.LocalObjectReference{Name: "postgresql"}, Key: "postgresql-root-password"}},
										},
									},
									VolumeMounts: []k8scorev1.VolumeMount{{Name: "foo", MountPath: "/bar"}},
								},
							},
							Volumes: []k8scorev1.Volume{{Name: "foo", VolumeSource: k8scorev1.VolumeSource{EmptyDir: &k8scorev1.EmptyDirVolumeSource{}}}},
						},
					},
				},
			},
		},
		{
			"OCI registry with repositories",
			"",
			&apprepov1alpha1.AppRepository{
				TypeMeta: k8smetav1.TypeMeta{
					Kind:       "AppRepository",
					APIVersion: "kubeapps.com/v1alpha1",
				},
				ObjectMeta: k8smetav1.ObjectMeta{
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
			k8sbatchv1.Job{
				ObjectMeta: k8smetav1.ObjectMeta{
					GenerateName: "apprepo-kubeapps-sync-my-charts-",
					OwnerReferences: []k8smetav1.OwnerReference{
						*k8smetav1.NewControllerRef(
							&apprepov1alpha1.AppRepository{ObjectMeta: k8smetav1.ObjectMeta{Name: "my-charts"}},
							k8sschema.GroupVersionKind{
								Group:   apprepov1alpha1.SchemeGroupVersion.Group,
								Version: apprepov1alpha1.SchemeGroupVersion.Version,
								Kind:    "AppRepository",
							},
						),
					},
					Annotations: map[string]string{},
					Labels:      map[string]string{},
				},
				Spec: k8sbatchv1.JobSpec{
					TTLSecondsAfterFinished: &defaultTTL,
					Template: k8scorev1.PodTemplateSpec{
						ObjectMeta: k8smetav1.ObjectMeta{
							Labels: map[string]string{
								LabelRepoName:      "my-charts",
								LabelRepoNamespace: "kubeapps",
							},
							Annotations: map[string]string{},
						},
						Spec: k8scorev1.PodSpec{
							RestartPolicy: "OnFailure",
							Containers: []k8scorev1.Container{
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
									Env: []k8scorev1.EnvVar{
										{
											Name: "DB_PASSWORD",
											ValueFrom: &k8scorev1.EnvVarSource{
												SecretKeyRef: &k8scorev1.SecretKeySelector{LocalObjectReference: k8scorev1.LocalObjectReference{Name: "postgresql"}, Key: "postgresql-root-password"}},
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
				TypeMeta: k8smetav1.TypeMeta{
					Kind:       "AppRepository",
					APIVersion: "kubeapps.com/v1alpha1",
				},
				ObjectMeta: k8smetav1.ObjectMeta{
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
			k8sbatchv1.Job{
				ObjectMeta: k8smetav1.ObjectMeta{
					GenerateName: "apprepo-kubeapps-sync-my-charts-",
					OwnerReferences: []k8smetav1.OwnerReference{
						*k8smetav1.NewControllerRef(
							&apprepov1alpha1.AppRepository{ObjectMeta: k8smetav1.ObjectMeta{Name: "my-charts"}},
							k8sschema.GroupVersionKind{
								Group:   apprepov1alpha1.SchemeGroupVersion.Group,
								Version: apprepov1alpha1.SchemeGroupVersion.Version,
								Kind:    "AppRepository",
							},
						),
					},
					Annotations: map[string]string{},
					Labels:      map[string]string{},
				},
				Spec: k8sbatchv1.JobSpec{
					TTLSecondsAfterFinished: &defaultTTL,
					Template: k8scorev1.PodTemplateSpec{
						ObjectMeta: k8smetav1.ObjectMeta{
							Labels: map[string]string{
								LabelRepoName:      "my-charts",
								LabelRepoNamespace: "kubeapps",
							},
							Annotations: map[string]string{},
						},
						Spec: k8scorev1.PodSpec{
							RestartPolicy: "OnFailure",
							Containers: []k8scorev1.Container{
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
									Env: []k8scorev1.EnvVar{
										{
											Name: "DB_PASSWORD",
											ValueFrom: &k8scorev1.EnvVarSource{
												SecretKeyRef: &k8scorev1.SecretKeySelector{LocalObjectReference: k8scorev1.LocalObjectReference{Name: "postgresql"}, Key: "postgresql-root-password"}},
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
				TypeMeta: k8smetav1.TypeMeta{
					Kind:       "AppRepository",
					APIVersion: "kubeapps.com/v1alpha1",
				},
				ObjectMeta: k8smetav1.ObjectMeta{
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
			k8sbatchv1.Job{
				ObjectMeta: k8smetav1.ObjectMeta{
					GenerateName: "apprepo-kubeapps-sync-my-charts-",
					OwnerReferences: []k8smetav1.OwnerReference{
						*k8smetav1.NewControllerRef(
							&apprepov1alpha1.AppRepository{ObjectMeta: k8smetav1.ObjectMeta{Name: "my-charts"}},
							k8sschema.GroupVersionKind{
								Group:   apprepov1alpha1.SchemeGroupVersion.Group,
								Version: apprepov1alpha1.SchemeGroupVersion.Version,
								Kind:    "AppRepository",
							},
						),
					},
					Annotations: map[string]string{},
					Labels:      map[string]string{},
				},
				Spec: k8sbatchv1.JobSpec{
					TTLSecondsAfterFinished: &defaultTTL,
					Template: k8scorev1.PodTemplateSpec{
						ObjectMeta: k8smetav1.ObjectMeta{
							Labels: map[string]string{
								LabelRepoName:      "my-charts",
								LabelRepoNamespace: "kubeapps",
							},
							Annotations: map[string]string{},
						},
						Spec: k8scorev1.PodSpec{
							RestartPolicy: "OnFailure",
							Containers: []k8scorev1.Container{
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
									Env: []k8scorev1.EnvVar{
										{
											Name: "DB_PASSWORD",
											ValueFrom: &k8scorev1.EnvVarSource{
												SecretKeyRef: &k8scorev1.SecretKeySelector{LocalObjectReference: k8scorev1.LocalObjectReference{Name: "postgresql"}, Key: "postgresql-root-password"}},
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
				TypeMeta: k8smetav1.TypeMeta{
					Kind:       "AppRepository",
					APIVersion: "kubeapps.com/v1alpha1",
				},
				ObjectMeta: k8smetav1.ObjectMeta{
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
			k8sbatchv1.Job{
				ObjectMeta: k8smetav1.ObjectMeta{
					GenerateName: "apprepo-kubeapps-sync-my-charts-",
					OwnerReferences: []k8smetav1.OwnerReference{
						*k8smetav1.NewControllerRef(
							&apprepov1alpha1.AppRepository{ObjectMeta: k8smetav1.ObjectMeta{Name: "my-charts"}},
							k8sschema.GroupVersionKind{
								Group:   apprepov1alpha1.SchemeGroupVersion.Group,
								Version: apprepov1alpha1.SchemeGroupVersion.Version,
								Kind:    "AppRepository",
							},
						),
					},
					Annotations: map[string]string{},
					Labels:      map[string]string{},
				},
				Spec: k8sbatchv1.JobSpec{
					TTLSecondsAfterFinished: &defaultTTL,
					Template: k8scorev1.PodTemplateSpec{
						ObjectMeta: k8smetav1.ObjectMeta{
							Labels: map[string]string{
								LabelRepoName:      "my-charts",
								LabelRepoNamespace: "kubeapps",
							},
							Annotations: map[string]string{},
						},
						Spec: k8scorev1.PodSpec{
							RestartPolicy: "OnFailure",
							Containers: []k8scorev1.Container{
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
									Env: []k8scorev1.EnvVar{
										{
											Name: "DB_PASSWORD",
											ValueFrom: &k8scorev1.EnvVarSource{
												SecretKeyRef: &k8scorev1.SecretKeySelector{LocalObjectReference: k8scorev1.LocalObjectReference{Name: "postgresql"}, Key: "postgresql-root-password"}},
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
		expected      k8sbatchv1.Job
	}{
		{
			"my-charts",
			"my-charts",
			"kubeapps",
			k8sbatchv1.Job{
				ObjectMeta: k8smetav1.ObjectMeta{
					GenerateName: "apprepo-kubeapps-cleanup-my-charts-",
					Namespace:    "kubeapps",
					Annotations:  map[string]string{},
					Labels:       map[string]string{},
				},
				Spec: k8sbatchv1.JobSpec{
					TTLSecondsAfterFinished: &defaultTTL,
					Template: k8scorev1.PodTemplateSpec{
						Spec: k8scorev1.PodSpec{
							RestartPolicy: "Never",
							Containers: []k8scorev1.Container{
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
									Env: []k8scorev1.EnvVar{
										{
											Name: "DB_PASSWORD",
											ValueFrom: &k8scorev1.EnvVarSource{
												SecretKeyRef: &k8scorev1.SecretKeySelector{LocalObjectReference: k8scorev1.LocalObjectReference{Name: "postgresql"}, Key: "postgresql-root-password"}},
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
		object k8smetav1.Object
		parent k8smetav1.Object
		expect bool
	}{
		{
			name: "it recognises a cronjob belonging to an app repository in another namespace",
			object: &k8sbatchv1beta1.CronJob{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name:      "apprepo-kubeapps-sync-my-charts",
					Namespace: "kubeapps",
					Labels: map[string]string{
						LabelRepoName:      "my-charts",
						LabelRepoNamespace: "my-namespace",
					},
				},
			},
			parent: &apprepov1alpha1.AppRepository{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name:      "my-charts",
					Namespace: "my-namespace",
				},
			},
			expect: true,
		},
		{
			name: "it returns false if the namespace does not match",
			object: &k8sbatchv1beta1.CronJob{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name:      "apprepo-kubeapps-sync-my-charts",
					Namespace: "kubeapps",
					Labels: map[string]string{
						LabelRepoName:      "my-charts",
						LabelRepoNamespace: "my-namespace",
					},
				},
			},
			parent: &apprepov1alpha1.AppRepository{
				ObjectMeta: k8smetav1.ObjectMeta{
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
		APIServerURL:             "",
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
