package main

import (
	"testing"

	"github.com/go-test/deep"
	apprepov1alpha1 "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func Test_newCronJob(t *testing.T) {
	mongoURL = "mongodb.kubeapps"
	mongoSecretName = "mongodb"
	tests := []struct {
		name             string
		apprepo          *apprepov1alpha1.AppRepository
		expected         batchv1beta1.CronJob
		userAgentComment string
	}{
		{
			"my-charts",
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
					Name:      "apprepo-sync-my-charts",
					Namespace: "kubeapps",
					OwnerReferences: []metav1.OwnerReference{
						*metav1.NewControllerRef(
							&apprepov1alpha1.AppRepository{ObjectMeta: metav1.ObjectMeta{Name: "my-charts"}},
							schema.GroupVersionKind{
								Group:   apprepov1alpha1.SchemeGroupVersion.Group,
								Version: apprepov1alpha1.SchemeGroupVersion.Version,
								Kind:    "AppRepository",
							}),
					},
				},
				Spec: batchv1beta1.CronJobSpec{
					Schedule:          "*/10 * * * *",
					ConcurrencyPolicy: "Replace",
					JobTemplate: batchv1beta1.JobTemplateSpec{
						Spec: batchv1.JobSpec{
							Template: corev1.PodTemplateSpec{
								ObjectMeta: metav1.ObjectMeta{
									Labels: map[string]string{"apprepositories.kubeapps.com/repo-name": "my-charts"},
								},
								Spec: corev1.PodSpec{
									RestartPolicy: "OnFailure",
									Containers: []corev1.Container{
										{
											Name:    "sync",
											Image:   repoSyncImage,
											Command: []string{"/chart-repo"},
											Args: []string{
												"sync",
												"--mongo-url=mongodb.kubeapps",
												"--mongo-user=root",
												"my-charts",
												"https://charts.acme.com/my-charts",
											},
											Env: []corev1.EnvVar{
												{
													Name: "MONGO_PASSWORD",
													ValueFrom: &corev1.EnvVarSource{
														SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "mongodb"}, Key: "mongodb-root-password"}},
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
			"",
		},
		{
			"my-charts with auth and userAgent comment",
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
					Name:      "apprepo-sync-my-charts",
					Namespace: "kubeapps",
					OwnerReferences: []metav1.OwnerReference{
						*metav1.NewControllerRef(
							&apprepov1alpha1.AppRepository{ObjectMeta: metav1.ObjectMeta{Name: "my-charts"}},
							schema.GroupVersionKind{
								Group:   apprepov1alpha1.SchemeGroupVersion.Group,
								Version: apprepov1alpha1.SchemeGroupVersion.Version,
								Kind:    "AppRepository",
							}),
					},
				},
				Spec: batchv1beta1.CronJobSpec{
					Schedule:          "*/10 * * * *",
					ConcurrencyPolicy: "Replace",
					JobTemplate: batchv1beta1.JobTemplateSpec{
						Spec: batchv1.JobSpec{
							Template: corev1.PodTemplateSpec{
								ObjectMeta: metav1.ObjectMeta{
									Labels: map[string]string{"apprepositories.kubeapps.com/repo-name": "my-charts"},
								},
								Spec: corev1.PodSpec{
									RestartPolicy: "OnFailure",
									Containers: []corev1.Container{
										{
											Name:    "sync",
											Image:   repoSyncImage,
											Command: []string{"/chart-repo"},
											Args: []string{
												"sync",
												"--mongo-url=mongodb.kubeapps",
												"--mongo-user=root",
												"--user-agent-comment=kubeapps/v2.3",
												"my-charts",
												"https://charts.acme.com/my-charts",
											},
											Env: []corev1.EnvVar{
												{
													Name: "MONGO_PASSWORD",
													ValueFrom: &corev1.EnvVarSource{
														SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "mongodb"}, Key: "mongodb-root-password"}},
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
			"kubeapps/v2.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.userAgentComment != "" {
				userAgentComment = tt.userAgentComment
				defer func() { userAgentComment = "" }()
			}
			result := newCronJob(tt.apprepo)
			if diff := deep.Equal(tt.expected, *result); diff != nil {
				t.Error(diff)
			}
		})
	}
}

func Test_newSyncJob(t *testing.T) {
	mongoURL = "mongodb.kubeapps"
	mongoSecretName = "mongodb"
	tests := []struct {
		name             string
		apprepo          *apprepov1alpha1.AppRepository
		expected         batchv1.Job
		userAgentComment string
	}{
		{
			"my-charts",
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
					GenerateName: "apprepo-sync-my-charts-",
					Namespace:    "kubeapps",
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
				},
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{"apprepositories.kubeapps.com/repo-name": "my-charts"},
						},
						Spec: corev1.PodSpec{
							RestartPolicy: "OnFailure",
							Containers: []corev1.Container{
								{
									Name:    "sync",
									Image:   repoSyncImage,
									Command: []string{"/chart-repo"},
									Args: []string{
										"sync",
										"--mongo-url=mongodb.kubeapps",
										"--mongo-user=root",
										"my-charts",
										"https://charts.acme.com/my-charts",
									},
									Env: []corev1.EnvVar{
										{
											Name: "MONGO_PASSWORD",
											ValueFrom: &corev1.EnvVarSource{
												SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "mongodb"}, Key: "mongodb-root-password"}},
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
			"",
		},
		{
			"my-charts with auth and userAgent comment",
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
					GenerateName: "apprepo-sync-my-charts-",
					Namespace:    "kubeapps",
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
				},
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{"apprepositories.kubeapps.com/repo-name": "my-charts"},
						},
						Spec: corev1.PodSpec{
							RestartPolicy: "OnFailure",
							Containers: []corev1.Container{
								{
									Name:    "sync",
									Image:   repoSyncImage,
									Command: []string{"/chart-repo"},
									Args: []string{
										"sync",
										"--mongo-url=mongodb.kubeapps",
										"--mongo-user=root",
										"--user-agent-comment=kubeapps/v2.3",
										"my-charts",
										"https://charts.acme.com/my-charts",
									},
									Env: []corev1.EnvVar{
										{
											Name: "MONGO_PASSWORD",
											ValueFrom: &corev1.EnvVarSource{
												SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "mongodb"}, Key: "mongodb-root-password"}},
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
			"kubeapps/v2.3",
		},
		{
			"my-charts with a customCA",
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
					GenerateName: "apprepo-sync-my-charts-",
					Namespace:    "kubeapps",
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
				},
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{"apprepositories.kubeapps.com/repo-name": "my-charts"},
						},
						Spec: corev1.PodSpec{
							RestartPolicy: "OnFailure",
							Containers: []corev1.Container{
								{
									Name:    "sync",
									Image:   repoSyncImage,
									Command: []string{"/chart-repo"},
									Args: []string{
										"sync",
										"--mongo-url=mongodb.kubeapps",
										"--mongo-user=root",
										"my-charts",
										"https://charts.acme.com/my-charts",
									},
									Env: []corev1.EnvVar{
										{
											Name: "MONGO_PASSWORD",
											ValueFrom: &corev1.EnvVarSource{
												SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "mongodb"}, Key: "mongodb-root-password"}},
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
			"",
		},
		{
			"my-charts with a customCA and auth header",
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
					GenerateName: "apprepo-sync-my-charts-",
					Namespace:    "kubeapps",
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
				},
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{"apprepositories.kubeapps.com/repo-name": "my-charts"},
						},
						Spec: corev1.PodSpec{
							RestartPolicy: "OnFailure",
							Containers: []corev1.Container{
								{
									Name:    "sync",
									Image:   repoSyncImage,
									Command: []string{"/chart-repo"},
									Args: []string{
										"sync",
										"--mongo-url=mongodb.kubeapps",
										"--mongo-user=root",
										"my-charts",
										"https://charts.acme.com/my-charts",
									},
									Env: []corev1.EnvVar{
										{
											Name: "MONGO_PASSWORD",
											ValueFrom: &corev1.EnvVarSource{
												SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "mongodb"}, Key: "mongodb-root-password"}},
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
			"",
		},
		{
			"my-charts with a custom pod template",
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
					GenerateName: "apprepo-sync-my-charts-",
					Namespace:    "kubeapps",
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
				},
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"apprepositories.kubeapps.com/repo-name": "my-charts",
								"foo":                                    "bar",
							},
						},
						Spec: corev1.PodSpec{
							Affinity:      &corev1.Affinity{NodeAffinity: &corev1.NodeAffinity{RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{}}},
							RestartPolicy: "OnFailure",
							Containers: []corev1.Container{
								{
									Name:    "sync",
									Image:   repoSyncImage,
									Command: []string{"/chart-repo"},
									Args: []string{
										"sync",
										"--mongo-url=mongodb.kubeapps",
										"--mongo-user=root",
										"my-charts",
										"https://charts.acme.com/my-charts",
									},
									Env: []corev1.EnvVar{
										{Name: "FOO", Value: "BAR"},
										{
											Name: "MONGO_PASSWORD",
											ValueFrom: &corev1.EnvVarSource{
												SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "mongodb"}, Key: "mongodb-root-password"}},
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
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.userAgentComment != "" {
				userAgentComment = tt.userAgentComment
				defer func() { userAgentComment = "" }()
			}

			result := newSyncJob(tt.apprepo)
			if diff := deep.Equal(tt.expected, *result); diff != nil {
				t.Error(diff)
			}
		})
	}
}

func Test_newCleanupJob(t *testing.T) {
	mongoURL = "mongodb.kubeapps"
	mongoSecretName = "mongodb"
	tests := []struct {
		name      string
		repoName  string
		namespace string
		expected  batchv1.Job
	}{
		{
			"my-charts",
			"my-charts",
			"kubeapps",
			batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "apprepo-cleanup-my-charts-",
					Namespace:    "kubeapps",
				},
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							RestartPolicy: "Never",
							Containers: []corev1.Container{
								{
									Name:    "delete",
									Image:   repoSyncImage,
									Command: []string{"/chart-repo"},
									Args: []string{
										"delete",
										"my-charts",
										"--mongo-url=mongodb.kubeapps",
										"--mongo-user=root",
									},
									Env: []corev1.EnvVar{
										{
											Name: "MONGO_PASSWORD",
											ValueFrom: &corev1.EnvVarSource{
												SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "mongodb"}, Key: "mongodb-root-password"}},
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
			result := newCleanupJob(tt.repoName, tt.namespace)
			if diff := deep.Equal(tt.expected, *result); diff != nil {
				t.Error(diff)
			}
		})
	}
}
