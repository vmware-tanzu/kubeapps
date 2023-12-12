// Copyright 2021-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/adhocore/gronx"
	apprepov1alpha1 "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/pkg/helm"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	log "k8s.io/klog/v2"
)

const (
	// LabelRepoName is the label used to identify the repository name.
	LabelRepoName = "apprepositories.kubeapps.com/repo-name"
	// LabelRepoNamespace is the label used to identify the repository namespace.
	LabelRepoNamespace = "apprepositories.kubeapps.com/repo-namespace"

	// Although a k8s typical length is 63, some characters are appended from the cronjob
	// to its spawned jobs therefore restricting this limit up to 52
	// https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/
	MAX_CRONJOB_CHARS = 52

	syncJobContainerName    = "sync"
	cleanupJobContainerName = "delete"
)

// newCronJob creates a new CronJob for a AppRepository resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the AppRepository resource that 'owns' it.
func newCronJob(apprepo *apprepov1alpha1.AppRepository, config Config) (*batchv1.CronJob, error) {
	var err error
	gron := gronx.New()
	cronTime := config.Crontab

	defaultValid := gron.IsValid(cronTime)
	if !defaultValid {
		return nil, fmt.Errorf("invalid default crontab for apprepo %q: %s", apprepo.GetName(), cronTime)
	}

	// If the apprepo has its own interval,
	// use that instead of the default global crontab.
	if apprepo.Spec.Interval != "" {
		// if the passed interval is indeed a cron expression, use it straight
		if gron.IsValid(apprepo.Spec.Interval) {
			cronTime = apprepo.Spec.Interval
		} else {
			// otherwise, convert it
			cronTime, err = intervalToCron(apprepo.Spec.Interval)
		}
	}
	// If the interval is invalid, use the default global crontab
	if err != nil {
		log.Errorf("Invalid interval for apprepo %q: %v", apprepo.GetName(), err)
		cronTime = config.Crontab
	}

	var concurrencyPolicy batchv1.ConcurrencyPolicy
	switch config.ConcurrencyPolicy {
	case "Allow", "allow":
		concurrencyPolicy = batchv1.AllowConcurrent
	case "Forbid", "forbid":
		concurrencyPolicy = batchv1.ForbidConcurrent
	case "Replace", "replace":
		concurrencyPolicy = batchv1.ReplaceConcurrent
	default:
		concurrencyPolicy = batchv1.ReplaceConcurrent
	}

	cronjob := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:            cronJobName(apprepo.GetNamespace(), apprepo.GetName(), false),
			OwnerReferences: ownerReferencesForAppRepo(apprepo, config.KubeappsNamespace),
			Labels:          jobLabels(apprepo, config),
			Annotations:     config.ParsedCustomAnnotations,
		},
		Spec: batchv1.CronJobSpec{
			Schedule:                   cronTime,
			SuccessfulJobsHistoryLimit: func(val int32) *int32 { return &val }(config.SuccessfulJobsHistoryLimit),
			FailedJobsHistoryLimit:     func(val int32) *int32 { return &val }(config.FailedJobsHistoryLimit),
			ConcurrencyPolicy:          concurrencyPolicy,
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: syncJobSpec(apprepo, config),
			},
		},
	}

	return cronjob, nil
}

// newSyncJob triggers a job for the AppRepository resource. It also sets the
// appropriate OwnerReferences on the resource
func newSyncJob(apprepo *apprepov1alpha1.AppRepository, config Config) *batchv1.Job {
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName:    cronJobName(apprepo.GetNamespace(), apprepo.GetName(), true),
			OwnerReferences: ownerReferencesForAppRepo(apprepo, config.KubeappsNamespace),
			Annotations:     config.ParsedCustomLabels,
			Labels:          config.ParsedCustomAnnotations,
		},
		Spec: syncJobSpec(apprepo, config),
	}
}

// jobSpec returns a batchv1.JobSpec for running the chart-repo sync job
func syncJobSpec(apprepo *apprepov1alpha1.AppRepository, config Config) batchv1.JobSpec {
	return generateJobSpec(apprepo, config, corev1.RestartPolicyOnFailure, syncJobContainerName, apprepoSyncJobArgs(apprepo, config))
}

// apprepoSyncJobArgs returns a list of args for the sync container
func apprepoSyncJobArgs(apprepo *apprepov1alpha1.AppRepository, config Config) []string {
	args := append([]string{syncJobContainerName}, dbFlags(config)...)

	if config.UserAgentComment != "" {
		args = append(args, "--user-agent-comment="+config.UserAgentComment)
	}

	args = append(args, "--global-repos-namespace="+config.GlobalPackagingNamespace)
	args = append(args, "--namespace="+apprepo.GetNamespace(), apprepo.GetName(), apprepo.Spec.URL, apprepo.Spec.Type)

	if len(apprepo.Spec.OCIRepositories) > 0 {
		args = append(args, "--oci-repositories", strings.Join(apprepo.Spec.OCIRepositories, ","))
	}

	if apprepo.Spec.TLSInsecureSkipVerify {
		args = append(args, "--tls-insecure-skip-verify")
	}

	if apprepo.Spec.PassCredentials {
		args = append(args, "--pass-credentials")
	}

	if apprepo.Spec.FilterRule.JQ != "" {
		rulesJSON, err := json.Marshal(apprepo.Spec.FilterRule)
		if err != nil {
			log.Errorf("Unable to parse filter rules for %s: %v", apprepo.GetName(), err)
		} else {
			args = append(args, "--filter-rules", string(rulesJSON))
		}
	}

	return args
}

// apprepoJobEnvVars returns a list of env variables for the sync/cleanup container
func apprepoJobEnvVars(apprepo *apprepov1alpha1.AppRepository, config Config) []corev1.EnvVar {
	var envVars []corev1.EnvVar
	envVars = append(envVars, corev1.EnvVar{
		Name: "DB_PASSWORD",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: config.DBSecretName},
				Key:                  config.DBSecretKey,
			},
		},
	})
	if config.OciCatalogUrl != "" {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "OCI_CATALOG_URL",
			Value: config.OciCatalogUrl,
		})
	}
	if apprepo.Spec.Auth.Header != nil {
		if apprepo.Spec.Auth.Header.SecretKeyRef.Key == ".dockerconfigjson" {
			envVars = append(envVars, corev1.EnvVar{
				Name: "DOCKER_CONFIG_JSON",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: secretKeyRefForRepo(apprepo.Spec.Auth.Header.SecretKeyRef, apprepo, config),
				},
			})
		} else {
			envVars = append(envVars, corev1.EnvVar{
				Name: "AUTHORIZATION_HEADER",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: secretKeyRefForRepo(apprepo.Spec.Auth.Header.SecretKeyRef, apprepo, config),
				},
			})
		}
	}
	return envVars
}

// newCleanupJob triggers a job for the AppRepository resource. It also sets the
// appropriate OwnerReferences on the resource
func newCleanupJob(apprepo *apprepov1alpha1.AppRepository, config Config) *batchv1.Job {
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: deleteJobName(apprepo.GetNamespace(), apprepo.GetName()),
			Namespace:    config.KubeappsNamespace,
			Annotations:  config.ParsedCustomAnnotations,
			Labels:       config.ParsedCustomLabels,
		},
		Spec: generateJobSpec(apprepo, config, corev1.RestartPolicyNever, cleanupJobContainerName, apprepoCleanupJobArgs(apprepo.GetNamespace(), apprepo.GetName(), config)),
	}
}

// apprepoCleanupJobArgs returns a list of args for the repo cleanup container
func apprepoCleanupJobArgs(namespace, name string, config Config) []string {
	return append([]string{
		cleanupJobContainerName,
		name,
		"--namespace=" + namespace,
	}, dbFlags(config)...)
}

func dbFlags(config Config) []string {
	return []string{
		"--database-url=" + config.DBURL,
		"--database-user=" + config.DBUser,
		"--database-name=" + config.DBName,
	}
}

// generateJobSpec returns a batchv1.JobSpec for running the chart-repo sync and clean-up jobs
func generateJobSpec(apprepo *apprepov1alpha1.AppRepository, config Config, restartPolicy corev1.RestartPolicy, containerName string, containerArgs []string) batchv1.JobSpec {
	volumes := []corev1.Volume{}
	volumeMounts := []corev1.VolumeMount{}
	if apprepo.Spec.Auth.CustomCA != nil {
		volumes = append(volumes, corev1.Volume{
			Name: apprepo.Spec.Auth.CustomCA.SecretKeyRef.Name,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: secretKeyRefForRepo(apprepo.Spec.Auth.CustomCA.SecretKeyRef, apprepo, config).Name,
					Items: []corev1.KeyToPath{
						{Key: apprepo.Spec.Auth.CustomCA.SecretKeyRef.Key, Path: "ca.crt"},
					},
				},
			},
		})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      apprepo.Spec.Auth.CustomCA.SecretKeyRef.Name,
			ReadOnly:  true,
			MountPath: "/usr/local/share/ca-certificates",
		})
	}

	// Get the predefined pod spec for the apprepo definition if exists
	podTemplateSpec := apprepo.Spec.SyncJobPodTemplate
	// Add labels
	if len(podTemplateSpec.ObjectMeta.Labels) == 0 {
		podTemplateSpec.ObjectMeta.Labels = map[string]string{}
	}
	for k, v := range jobLabels(apprepo, config) {
		podTemplateSpec.ObjectMeta.Labels[k] = v
	}
	podTemplateSpec.ObjectMeta.Annotations = config.ParsedCustomAnnotations
	// If there's an issue, won't restart
	podTemplateSpec.Spec.RestartPolicy = restartPolicy
	// Populate container spec
	if len(podTemplateSpec.Spec.Containers) == 0 {
		podTemplateSpec.Spec.Containers = []corev1.Container{{}}
	}
	// Populate ImagePullSecrets spec
	podTemplateSpec.Spec.ImagePullSecrets = append(podTemplateSpec.Spec.ImagePullSecrets, config.ImagePullSecretsRefs...)

	podTemplateSpec.Spec.Containers[0].Name = containerName
	podTemplateSpec.Spec.Containers[0].Image = config.RepoSyncImage
	podTemplateSpec.Spec.Containers[0].ImagePullPolicy = corev1.PullIfNotPresent
	podTemplateSpec.Spec.Containers[0].Command = []string{config.RepoSyncCommand}
	podTemplateSpec.Spec.Containers[0].Args = containerArgs
	podTemplateSpec.Spec.Containers[0].Env = append(podTemplateSpec.Spec.Containers[0].Env, apprepoJobEnvVars(apprepo, config)...)
	podTemplateSpec.Spec.Containers[0].VolumeMounts = append(podTemplateSpec.Spec.Containers[0].VolumeMounts, volumeMounts...)
	// Add volumes
	podTemplateSpec.Spec.Volumes = append(podTemplateSpec.Spec.Volumes, volumes...)

	return batchv1.JobSpec{
		TTLSecondsAfterFinished: ttlLifetimeJobs(config),
		ActiveDeadlineSeconds: activeDeadlineSeconds(config),
		Template:                podTemplateSpec,
	}
}

// jobLabels returns the labels for the job and cronjob resources
func jobLabels(apprepo *apprepov1alpha1.AppRepository, config Config) map[string]string {
	// Adding the default labels
	labels := map[string]string{
		LabelRepoName:      apprepo.GetName(),
		LabelRepoNamespace: apprepo.GetNamespace(),
	}
	// Add the custom labels from the config
	for k, v := range config.ParsedCustomLabels {
		labels[k] = v
	}
	return labels
}

// secretKeyRefForRepo returns a secret key ref with a name depending on whether
// this repo is in the kubeapps namespace or not. If the repo is not in the
// kubeapps namespace, then the secret will have been copied from another namespace
// into the kubeapps namespace and have a slightly different name.
func secretKeyRefForRepo(keyRef corev1.SecretKeySelector, apprepo *apprepov1alpha1.AppRepository, config Config) *corev1.SecretKeySelector {
	if apprepo.ObjectMeta.Namespace == config.KubeappsNamespace {
		return &keyRef
	}
	keyRef.LocalObjectReference.Name = helm.SecretNameForNamespacedRepo(apprepo.ObjectMeta.Name, apprepo.ObjectMeta.Namespace)
	return &keyRef
}

// ownerReferencesForAppRepo returns populated owner references for app repos in the same namespace
// as the cronjob and nil otherwise.
func ownerReferencesForAppRepo(apprepo *apprepov1alpha1.AppRepository, childNamespace string) []metav1.OwnerReference {
	if apprepo.GetNamespace() == childNamespace {
		return []metav1.OwnerReference{
			*metav1.NewControllerRef(apprepo, schema.GroupVersionKind{
				Group:   apprepov1alpha1.SchemeGroupVersion.Group,
				Version: apprepov1alpha1.SchemeGroupVersion.Version,
				Kind:    AppRepository,
			}),
		}
	}
	return nil
}

// ttlLifetimeJobs return time to live set by user otherwise return nil
func ttlLifetimeJobs(config Config) *int32 {
	if config.TTLSecondsAfterFinished != "" {
		configTTL, err := strconv.ParseInt(config.TTLSecondsAfterFinished, 10, 32)
		if err == nil {
			result := int32(configTTL)
			return &result
		}
	}
	return nil
}

// activeDeadlineSeconds return active deadline seconds set by user otherwise return nil
func activeDeadlineSeconds(config Config) *int64 {
	if config.ActiveDeadlineSeconds != "" {
		configDeadline, err := strconv.ParseInt(config.ActiveDeadlineSeconds, 10, 64)
		if err == nil {
			result := int64(configDeadline)
			return &result
		}
	}
	return nil
}

// cronJobName returns a unique name for the CronJob managed by an AppRepository
func cronJobName(namespace, name string, addDash bool) string {
	// the "apprepo--sync-" string has 14 chars, which leaves us 52-14=38 chars for the final name
	return generateJobName(namespace, name, "apprepo-%s-sync-%s", addDash)

}

// deleteJobName returns a unique name for the Job to cleanup AppRepository
func deleteJobName(namespace, name string) string {
	// the "apprepo--cleanup--" string has 18 chars, which leaves us 52-18=34 chars for the final name
	return generateJobName(namespace, name, "apprepo-%s-cleanup-%s-", false)
}

// generateJobName returns a unique name for the Job managed by an AppRepository
func generateJobName(namespace, name, pattern string, addDash bool) string {
	// ensure there are enough placeholders to be replaces later
	if strings.Count(pattern, "%s") != 2 {
		return ""
	}

	// calculate the length used by the name pattern
	patternLen := len(strings.ReplaceAll(pattern, "%s", ""))

	// for example: the "apprepo--cleanup--" string has 18 chars, which leaves us 52-18=34 chars for the final name
	maxNamespaceLength, rem := (MAX_CRONJOB_CHARS-patternLen)/2, (MAX_CRONJOB_CHARS-patternLen)%2
	maxNameLength := maxNamespaceLength
	if rem > 0 && !addDash {
		maxNameLength++
	}

	if addDash {
		pattern = fmt.Sprintf("%s-", pattern)
	}

	truncatedName := fmt.Sprintf(pattern, truncateAndHashString(namespace, maxNamespaceLength), truncateAndHashString(name, maxNameLength))

	return truncatedName
}
