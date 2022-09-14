// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/adler32"
	"strconv"
	"strings"
	"time"

	apprepov1alpha1 "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	clientset "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/client/clientset/versioned"
	appreposcheme "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/client/clientset/versioned/scheme"
	informers "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/client/informers/externalversions"
	listers "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/client/listers/apprepository/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/pkg/kube"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	batchlisters "k8s.io/client-go/listers/batch/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	log "k8s.io/klog/v2"
)

const controllerAgentName = "apprepository-controller"

// Although a k8s typical length is 63, some characters are appended from the cronjob
// to its spawned jobs therefore restricting this limit up to 52
// https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/
const MAX_CRONJOB_CHARS = 52

const (
	// SuccessSynced is used as part of the Event 'reason' when an AppRepository
	// is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when an
	// AppRepository fails to sync due to a CronJob of the same name already
	// existing.
	ErrResourceExists = "ErrResourceExists"

	// LabelRepoName is the label used to identify the repository name.
	LabelRepoName = "apprepositories.kubeapps.com/repo-name"
	// LabelRepoNamespace is the label used to identify the repository namespace.
	LabelRepoNamespace = "apprepositories.kubeapps.com/repo-namespace"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a CronJob already existing
	MessageResourceExists = "Resource %q already exists and is not managed by AppRepository"
	// MessageResourceSynced is the message used for an Event fired when an
	// AppRepsitory is synced successfully
	MessageResourceSynced = "AppRepository synced successfully"
)

// Controller is the controller implementation for AppRepository resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// apprepoclientset is the clientset for AppRepository resources
	apprepoclientset clientset.Interface

	cronjobsLister batchlisters.CronJobLister
	cronjobsSynced cache.InformerSynced
	appreposLister listers.AppRepositoryLister
	appreposSynced cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder

	conf Config
}

// NewController returns a new sample controller
func NewController(
	kubeclientset kubernetes.Interface,
	apprepoclientset clientset.Interface,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	apprepoInformerFactory informers.SharedInformerFactory,
	conf *Config) *Controller {

	// obtain references to shared index informers for the CronJob and
	// AppRepository types.
	cronjobInformer := kubeInformerFactory.Batch().V1().CronJobs()
	apprepoInformer := apprepoInformerFactory.Kubeapps().V1alpha1().AppRepositories()

	// Create event broadcaster
	// Add apprepository-controller types to the default Kubernetes Scheme so
	// Events can be logged for apprepository-controller types.
	err := appreposcheme.AddToScheme(scheme.Scheme)
	if err != nil {
		return nil
	}

	log.Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(log.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeclientset:    kubeclientset,
		apprepoclientset: apprepoclientset,
		cronjobsLister:   cronjobInformer.Lister(),
		cronjobsSynced:   cronjobInformer.Informer().HasSynced,
		appreposLister:   apprepoInformer.Lister(),
		appreposSynced:   apprepoInformer.Informer().HasSynced,
		workqueue:        workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "AppRepositories"),
		recorder:         recorder,
		conf:             *conf,
	}

	log.Info("Setting up event handlers")
	// Set up an event handler for when AppRepository resources change
	apprepoInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueAppRepo,
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldApp := oldObj.(*apprepov1alpha1.AppRepository)
			newApp := newObj.(*apprepov1alpha1.AppRepository)
			if oldApp.Spec.URL != newApp.Spec.URL || oldApp.Spec.ResyncRequests != newApp.Spec.ResyncRequests {
				controller.enqueueAppRepo(newApp)
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				controller.workqueue.AddRateLimited(key)
			}
		},
	})

	// Set up an event handler for when CronJob resources get deleted. This
	// handler will lookup the owner of the given CronJob, and if it is owned by a
	// AppRepository resource will enqueue that AppRepository resource for
	// processing so the CronJob gets correctly recreated. This way, we don't need
	// to implement custom logic for handling CronJob resources. More info on this
	// pattern:
	// https://github.com/kubernetes/community/blob/8cafef897a22026d42f5e5bb3f104febe7e29830/contributors/devel/controllers.md
	cronjobInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: controller.handleObject,
	})

	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	log.Info("Starting AppRepository controller")

	// Wait for the caches to be synced before starting workers
	log.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.cronjobsSynced, c.appreposSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	log.Info("Starting workers")
	// Launch two workers to process AppRepository resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	log.Info("Started workers")
	<-stopCh
	log.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// AppRepository resource to be synced.
		if err := c.syncHandler(key); err != nil {
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		log.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the AppRepository
// resource with the current status of the resource.
func (c *Controller) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the AppRepository resource with this namespace/name
	apprepo, err := c.appreposLister.AppRepositories(namespace).Get(name)
	if err != nil {
		// The AppRepository resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			log.Infof("AppRepository '%s' no longer exists so performing cleanup of charts from the DB", key)
			// Trigger a Job to perfrom the cleanup of the charts in the DB corresponding to deleted AppRepository
			_, err = c.kubeclientset.BatchV1().Jobs(c.conf.KubeappsNamespace).Create(context.TODO(), newCleanupJob(c.conf.KubeappsNamespace, namespace, name, c.conf), metav1.CreateOptions{})
			if err != nil {
				log.Errorf("Unable to create cleanup job: %v", err)
				return err
			}

			// TODO: Workaround until the sync jobs are moved to the repoNamespace (#1647)
			// Delete the cronjob in the Kubeapps namespace to avoid re-syncing the repository
			err = c.kubeclientset.BatchV1().CronJobs(c.conf.KubeappsNamespace).Delete(context.TODO(), cronJobName(namespace, name, false), metav1.DeleteOptions{})
			if err != nil && !errors.IsNotFound(err) {
				log.Errorf("Unable to delete sync cronjob: %v", err)
				return err
			}
			return nil
		}
		return fmt.Errorf("Error fetching object with key %s from store: %v", key, err)
	}

	// Get the cronjob with the same name as AppRepository
	cronjobName := cronJobName(namespace, name, false)
	cronjob, err := c.cronjobsLister.CronJobs(c.conf.KubeappsNamespace).Get(cronjobName)
	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(err) {
		log.Infof("Creating CronJob %q for AppRepository %q", cronjobName, apprepo.GetName())
		cronjob, err = c.kubeclientset.BatchV1().CronJobs(c.conf.KubeappsNamespace).Create(context.TODO(), newCronJob(apprepo, c.conf), metav1.CreateOptions{})
		if err != nil {
			return err
		}

		// Trigger a manual Job for the initial sync
		_, err = c.kubeclientset.BatchV1().Jobs(c.conf.KubeappsNamespace).Create(context.TODO(), newSyncJob(apprepo, c.conf), metav1.CreateOptions{})
	} else if err == nil {
		// If the resource already exists, we'll update it
		log.Infof("Updating CronJob %q in namespace %q for AppRepository %q in namespace %q", cronjobName, c.conf.KubeappsNamespace, apprepo.GetName(), apprepo.GetNamespace())
		cronjob, err = c.kubeclientset.BatchV1().CronJobs(c.conf.KubeappsNamespace).Update(context.TODO(), newCronJob(apprepo, c.conf), metav1.UpdateOptions{})
		if err != nil {
			return err
		}

		// The AppRepository has changed, launch a manual Job
		_, err = c.kubeclientset.BatchV1().Jobs(c.conf.KubeappsNamespace).Create(context.TODO(), newSyncJob(apprepo, c.conf), metav1.CreateOptions{})
	}

	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// If the CronJob is not controlled by this AppRepository resource and it is not a
	// cronjob for an app repo in another namespace, then we should
	// log a warning to the event recorder and return it.
	if !metav1.IsControlledBy(cronjob, apprepo) && !objectBelongsTo(cronjob, apprepo) {
		msg := fmt.Sprintf(MessageResourceExists, cronjob.Name)
		c.recorder.Event(apprepo, corev1.EventTypeWarning, ErrResourceExists, msg)
		return fmt.Errorf(msg)
	}

	if apprepo.GetNamespace() == c.conf.KubeappsNamespace {
		c.recorder.Event(apprepo, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	}
	return nil
}

// belongsTo is similar to IsControlledBy, but enables us to establish a relationship
// between cronjobs and app repositories in different namespaces.
func objectBelongsTo(object, parent metav1.Object) bool {
	labels := object.GetLabels()
	return labels[LabelRepoName] == parent.GetName() && labels[LabelRepoNamespace] == parent.GetNamespace()
}

// enqueueAppRepo takes a AppRepository resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than AppRepository.
func (c *Controller) enqueueAppRepo(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
}

// handleObject will take any resource implementing metav1.Object and attempt to
// find the AppRepository resource that 'owns' it. It does this by looking at
// the objects metadata.ownerReferences field for an appropriate OwnerReference.
// It then enqueues that AppRepository resource to be processed. If the object
// does not have an appropriate OwnerReference, it will simply be skipped.
func (c *Controller) handleObject(obj interface{}) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("error decoding object, invalid type"))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			runtime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
			return
		}
		log.Infof("Recovered deleted object '%s' from tombstone", object.GetName())
	}
	log.Infof("Processing object: %s", object.GetName())
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		// If this object is not owned by an AppRepository, we should not do
		// anything more with it.
		if ownerRef.Kind != "AppRepository" {
			return
		}

		apprepo, err := c.appreposLister.AppRepositories(object.GetNamespace()).Get(ownerRef.Name)
		if err != nil {
			log.Infof("ignoring orphaned object '%s' of AppRepository '%s'", object.GetSelfLink(), ownerRef.Name)
			return
		}

		if apprepo.ObjectMeta.DeletionTimestamp != nil {
			log.Infof("ignoring object %q of AppRepository %q with deletion timestamp %q", object.GetSelfLink(), ownerRef.Name, apprepo.ObjectMeta.DeletionTimestamp)
			return
		}

		c.enqueueAppRepo(apprepo)
		return
	}
}

// ownerReferencesForAppRepo returns populated owner references for app repos in the same namespace
// as the cronjob and nil otherwise.
func ownerReferencesForAppRepo(apprepo *apprepov1alpha1.AppRepository, childNamespace string) []metav1.OwnerReference {
	if apprepo.GetNamespace() == childNamespace {
		return []metav1.OwnerReference{
			*metav1.NewControllerRef(apprepo, schema.GroupVersionKind{
				Group:   apprepov1alpha1.SchemeGroupVersion.Group,
				Version: apprepov1alpha1.SchemeGroupVersion.Version,
				Kind:    "AppRepository",
			}),
		}
	}
	return nil
}

// newCronJob creates a new CronJob for a AppRepository resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the AppRepository resource that 'owns' it.
func newCronJob(apprepo *apprepov1alpha1.AppRepository, config Config) *batchv1.CronJob {
	return &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:            cronJobName(apprepo.Namespace, apprepo.Name, false),
			OwnerReferences: ownerReferencesForAppRepo(apprepo, config.KubeappsNamespace),
			Labels:          jobLabels(apprepo, config),
			Annotations:     config.ParsedCustomAnnotations,
		},
		Spec: batchv1.CronJobSpec{
			Schedule: config.Crontab,
			// Set to replace as short-circuit in k8s <1.12
			// TODO re-evaluate ConcurrentPolicy when 1.12+ is mainstream (i.e 1.14)
			// https://github.com/kubernetes/kubernetes/issues/54870
			ConcurrencyPolicy: "Replace",
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: syncJobSpec(apprepo, config),
			},
		},
	}
}

// newSyncJob triggers a job for the AppRepository resource. It also sets the
// appropriate OwnerReferences on the resource
func newSyncJob(apprepo *apprepov1alpha1.AppRepository, config Config) *batchv1.Job {
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName:    cronJobName(apprepo.Namespace, apprepo.Name, true),
			OwnerReferences: ownerReferencesForAppRepo(apprepo, config.KubeappsNamespace),
			Annotations:     config.ParsedCustomLabels,
			Labels:          config.ParsedCustomAnnotations,
		},
		Spec: syncJobSpec(apprepo, config),
	}
}

// jobSpec returns a batchv1.JobSpec for running the chart-repo sync job
func syncJobSpec(apprepo *apprepov1alpha1.AppRepository, config Config) batchv1.JobSpec {
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
	// If there's an issue, will restart pod until successful or replaced
	// by another instance of the job scheduled by the cronjob
	// see: cronJobSpec.concurrencyPolicy
	podTemplateSpec.Spec.RestartPolicy = "OnFailure"
	// Populate container spec
	if len(podTemplateSpec.Spec.Containers) == 0 {
		podTemplateSpec.Spec.Containers = []corev1.Container{{}}
	}
	// Populate ImagePullSecrets spec
	podTemplateSpec.Spec.ImagePullSecrets = append(podTemplateSpec.Spec.ImagePullSecrets, config.ImagePullSecretsRefs...)

	podTemplateSpec.Spec.Containers[0].Name = "sync"
	podTemplateSpec.Spec.Containers[0].Image = config.RepoSyncImage
	podTemplateSpec.Spec.Containers[0].ImagePullPolicy = "IfNotPresent"
	podTemplateSpec.Spec.Containers[0].Command = []string{config.RepoSyncCommand}
	podTemplateSpec.Spec.Containers[0].Args = apprepoSyncJobArgs(apprepo, config)
	podTemplateSpec.Spec.Containers[0].Env = append(podTemplateSpec.Spec.Containers[0].Env, apprepoSyncJobEnvVars(apprepo, config)...)
	podTemplateSpec.Spec.Containers[0].VolumeMounts = append(podTemplateSpec.Spec.Containers[0].VolumeMounts, volumeMounts...)
	// Add volumes
	podTemplateSpec.Spec.Volumes = append(podTemplateSpec.Spec.Volumes, volumes...)

	return batchv1.JobSpec{
		TTLSecondsAfterFinished: ttlLifetimeJobs(config),
		Template:                podTemplateSpec,
	}
}

// newCleanupJob triggers a job for the AppRepository resource. It also sets the
// appropriate OwnerReferences on the resource
func newCleanupJob(kubeappsNamespace, repoNamespace, name string, config Config) *batchv1.Job {
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: deleteJobName(repoNamespace, name),
			Namespace:    kubeappsNamespace,
			Annotations:  config.ParsedCustomAnnotations,
			Labels:       config.ParsedCustomLabels,
		},
		Spec: cleanupJobSpec(repoNamespace, name, config),
	}
}

// cleanupJobSpec returns a batchv1.JobSpec for running the chart-repo delete job
func cleanupJobSpec(namespace, name string, config Config) batchv1.JobSpec {
	return batchv1.JobSpec{
		TTLSecondsAfterFinished: ttlLifetimeJobs(config),
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				// If there's an issue, delay till the next cron
				RestartPolicy:    "Never",
				ImagePullSecrets: config.ImagePullSecretsRefs,
				Containers: []corev1.Container{
					{
						Name:            "delete",
						Image:           config.RepoSyncImage,
						ImagePullPolicy: "IfNotPresent",
						Command:         []string{config.RepoSyncCommand},
						Args:            apprepoCleanupJobArgs(namespace, name, config),
						Env: []corev1.EnvVar{
							{
								Name: "DB_PASSWORD",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{Name: config.DBSecretName},
										Key:                  config.DBSecretKey,
									},
								},
							},
						},
					},
				},
			},
		},
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
	maxNamesapceLength, rem := (MAX_CRONJOB_CHARS-patternLen)/2, (MAX_CRONJOB_CHARS-patternLen)%2
	maxNameLength := maxNamesapceLength
	if rem > 0 && !addDash {
		maxNameLength++
	}

	if addDash {
		pattern = fmt.Sprintf("%s-", pattern)
	}

	truncatedName := fmt.Sprintf(pattern, truncateAndHashString(namespace, maxNamesapceLength), truncateAndHashString(name, maxNameLength))

	return truncatedName
}

// truncateAndHashString truncates the string to a max length and hashes the rest of it
// Ex: truncateAndHashString(aaaaaaaaaaaaaaaaaaaaaaaaaa,12) becomes "a-2067663226"
func truncateAndHashString(name string, length int) string {
	if len(name) > length {
		if length < 11 {
			return name[:length]
		}
		log.Warningf("Name %q exceedes %d characters (got %d)", name, length, len(name))
		// max length chars, minus 10 chars (the adler32 hash returns up to 10 digits), minus 1 for the '-'
		splitPoint := length - 11
		part1 := name[:splitPoint]
		part2 := name[splitPoint:]
		hashedPart2 := fmt.Sprint(adler32.Checksum([]byte(part2)))
		name = fmt.Sprintf("%s-%s", part1, hashedPart2)
	}
	return name
}

// apprepoSyncJobArgs returns a list of args for the sync container
func apprepoSyncJobArgs(apprepo *apprepov1alpha1.AppRepository, config Config) []string {
	args := append([]string{"sync"}, dbFlags(config)...)

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
			log.Errorf("Unable to parse filter rules for %s: %v", apprepo.Name, err)
		} else {
			args = append(args, "--filter-rules", string(rulesJSON))
		}
	}

	return args
}

// apprepoSyncJobEnvVars returns a list of env variables for the sync container
func apprepoSyncJobEnvVars(apprepo *apprepov1alpha1.AppRepository, config Config) []corev1.EnvVar {
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

// secretKeyRefForRepo returns a secret key ref with a name depending on whether
// this repo is in the kubeapps namespace or not. If the repo is not in the
// kubeapps namespace, then the secret will have been copied from another namespace
// into the kubeapps namespace and have a slightly different name.
func secretKeyRefForRepo(keyRef corev1.SecretKeySelector, apprepo *apprepov1alpha1.AppRepository, config Config) *corev1.SecretKeySelector {
	if apprepo.ObjectMeta.Namespace == config.KubeappsNamespace {
		return &keyRef
	}
	keyRef.LocalObjectReference.Name = kube.KubeappsSecretNameForRepo(apprepo.ObjectMeta.Name, apprepo.ObjectMeta.Namespace)
	return &keyRef
}

// apprepoCleanupJobArgs returns a list of args for the repo cleanup container
func apprepoCleanupJobArgs(namespace, name string, config Config) []string {
	return append([]string{
		"delete",
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
