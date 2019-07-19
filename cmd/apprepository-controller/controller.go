/*
Copyright 2017 Bitnami.

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

package main

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	apprepov1alpha1 "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	clientset "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/client/clientset/versioned"
	appreposcheme "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/client/clientset/versioned/scheme"
	informers "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/client/informers/externalversions"
	listers "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/client/listers/apprepository/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
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
	batchlisters "k8s.io/client-go/listers/batch/v1beta1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

const controllerAgentName = "apprepository-controller"

const (
	// SuccessSynced is used as part of the Event 'reason' when an AppRepository
	// is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when an
	// AppRepository fails to sync due to a CronJob of the same name already
	// existing.
	ErrResourceExists = "ErrResourceExists"

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
}

// NewController returns a new sample controller
func NewController(
	kubeclientset kubernetes.Interface,
	apprepoclientset clientset.Interface,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	apprepoInformerFactory informers.SharedInformerFactory) *Controller {

	// obtain references to shared index informers for the CronJob and
	// AppRepository types.
	cronjobInformer := kubeInformerFactory.Batch().V1beta1().CronJobs()
	apprepoInformer := apprepoInformerFactory.Kubeapps().V1alpha1().AppRepositories()

	// Create event broadcaster
	// Add apprepository-controller types to the default Kubernetes Scheme so
	// Events can be logged for apprepository-controller types.
	appreposcheme.AddToScheme(scheme.Scheme)
	glog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
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
	}

	glog.Info("Setting up event handlers")
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
	glog.Info("Starting AppRepository controller")

	// Wait for the caches to be synced before starting workers
	glog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.cronjobsSynced, c.appreposSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	glog.Info("Starting workers")
	// Launch two workers to process AppRepository resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	glog.Info("Started workers")
	<-stopCh
	glog.Info("Shutting down workers")

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
		glog.Infof("Successfully synced '%s'", key)
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
			glog.Infof("AppRepository '%s' no longer exists so performing cleanup of charts from the DB", key)
			// Trigger a Job to perfrom the cleanup of the charts in the DB corresponding to deleted AppRepository
			_, err = c.kubeclientset.BatchV1().Jobs(namespace).Create(newCleanupJob(name, namespace))
			return nil
		}
		return fmt.Errorf("Error fetching object with key %s from store: %v", key, err)
	}

	// Get the cronjob with the same name as AppRepository
	cronjobName := cronJobName(apprepo)
	cronjob, err := c.cronjobsLister.CronJobs(apprepo.Namespace).Get(cronjobName)
	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(err) {
		glog.V(4).Infof("Creating CronJob %q for AppRepository %q", cronjobName, apprepo.GetName())
		cronjob, err = c.kubeclientset.BatchV1beta1().CronJobs(apprepo.Namespace).Create(newCronJob(apprepo))
		if err != nil {
			return err
		}

		// Trigger a manual Job for the initial sync
		_, err = c.kubeclientset.BatchV1().Jobs(apprepo.Namespace).Create(newSyncJob(apprepo))
	} else if err == nil {
		// If the resource already exists, we'll update it
		glog.V(4).Infof("Updating CronJob %q for AppRepository %q", cronjobName, apprepo.GetName())
		cronjob, err = c.kubeclientset.BatchV1beta1().CronJobs(apprepo.Namespace).Update(newCronJob(apprepo))
		if err != nil {
			return err
		}

		// The AppRepository has changed, launch a manual Job
		_, err = c.kubeclientset.BatchV1().Jobs(apprepo.Namespace).Create(newSyncJob(apprepo))
	}

	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// If the CronJob is not controlled by this AppRepository resource, we should
	// log a warning to the event recorder and ret
	if !metav1.IsControlledBy(cronjob, apprepo) {
		msg := fmt.Sprintf(MessageResourceExists, cronjob.Name)
		c.recorder.Event(apprepo, corev1.EventTypeWarning, ErrResourceExists, msg)
		return fmt.Errorf(msg)
	}

	// If an error occurs during Update, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	c.recorder.Event(apprepo, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
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
		glog.V(4).Infof("Recovered deleted object '%s' from tombstone", object.GetName())
	}
	glog.V(4).Infof("Processing object: %s", object.GetName())
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		// If this object is not owned by an AppRepository, we should not do
		// anything more with it.
		if ownerRef.Kind != "AppRepository" {
			return
		}

		apprepo, err := c.appreposLister.AppRepositories(object.GetNamespace()).Get(ownerRef.Name)
		if err != nil {
			glog.V(4).Infof("ignoring orphaned object '%s' of AppRepository '%s'", object.GetSelfLink(), ownerRef.Name)
			return
		}

		c.enqueueAppRepo(apprepo)
		return
	}
}

// newCronJob creates a new CronJob for a AppRepository resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the AppRepository resource that 'owns' it.
func newCronJob(apprepo *apprepov1alpha1.AppRepository) *batchv1beta1.CronJob {
	return &batchv1beta1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cronJobName(apprepo),
			Namespace: apprepo.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(apprepo, schema.GroupVersionKind{
					Group:   apprepov1alpha1.SchemeGroupVersion.Group,
					Version: apprepov1alpha1.SchemeGroupVersion.Version,
					Kind:    "AppRepository",
				}),
			},
		},
		Spec: batchv1beta1.CronJobSpec{
			// TODO: make schedule customisable
			Schedule: "*/10 * * * *",
			// Set to replace as short-circuit in k8s <1.12
			// TODO re-evaluate ConcurrentPolicy when 1.12+ is mainstream (i.e 1.14)
			// https://github.com/kubernetes/kubernetes/issues/54870
			ConcurrencyPolicy: "Replace",
			JobTemplate: batchv1beta1.JobTemplateSpec{
				Spec: syncJobSpec(apprepo),
			},
		},
	}
}

// newSyncJob triggers a job for the AppRepository resource. It also sets the
// appropriate OwnerReferences on the resource
func newSyncJob(apprepo *apprepov1alpha1.AppRepository) *batchv1.Job {
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: cronJobName(apprepo) + "-",
			Namespace:    apprepo.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(apprepo, schema.GroupVersionKind{
					Group:   apprepov1alpha1.SchemeGroupVersion.Group,
					Version: apprepov1alpha1.SchemeGroupVersion.Version,
					Kind:    "AppRepository",
				}),
			},
		},
		Spec: syncJobSpec(apprepo),
	}
}

// jobSpec returns a batchv1.JobSpec for running the chart-repo sync job
func syncJobSpec(apprepo *apprepov1alpha1.AppRepository) batchv1.JobSpec {
	volumes := []corev1.Volume{}
	volumeMounts := []corev1.VolumeMount{}
	if apprepo.Spec.Auth.CustomCA != nil {
		volumes = append(volumes, corev1.Volume{
			Name: apprepo.Spec.Auth.CustomCA.SecretKeyRef.Name,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: apprepo.Spec.Auth.CustomCA.SecretKeyRef.Name,
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
	for k, v := range jobLabels(apprepo) {
		podTemplateSpec.ObjectMeta.Labels[k] = v
	}
	// If there's an issue, will restart pod until sucessful or replaced
	// by another instance of the job scheduled by the cronjob
	// see: cronJobSpec.concurrencyPolicy
	podTemplateSpec.Spec.RestartPolicy = "OnFailure"
	// Populate container spec
	if len(podTemplateSpec.Spec.Containers) == 0 {
		podTemplateSpec.Spec.Containers = []corev1.Container{{}}
	}
	podTemplateSpec.Spec.Containers[0].Name = "sync"
	podTemplateSpec.Spec.Containers[0].Image = repoSyncImage
	podTemplateSpec.Spec.Containers[0].Command = []string{"/chart-repo"}
	podTemplateSpec.Spec.Containers[0].Args = apprepoSyncJobArgs(apprepo)
	podTemplateSpec.Spec.Containers[0].Env = append(podTemplateSpec.Spec.Containers[0].Env, apprepoSyncJobEnvVars(apprepo)...)
	podTemplateSpec.Spec.Containers[0].VolumeMounts = append(podTemplateSpec.Spec.Containers[0].VolumeMounts, volumeMounts...)
	// Add volumes
	podTemplateSpec.Spec.Volumes = append(podTemplateSpec.Spec.Volumes, volumes...)

	return batchv1.JobSpec{
		Template: podTemplateSpec,
	}
}

// newCleanupJob triggers a job for the AppRepository resource. It also sets the
// appropriate OwnerReferences on the resource
func newCleanupJob(reponame, namespace string) *batchv1.Job {
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: deleteJobName(reponame) + "-",
			Namespace:    namespace,
		},
		Spec: cleanupJobSpec(reponame),
	}
}

// cleanupJobSpec returns a batchv1.JobSpec for running the chart-repo delete job
func cleanupJobSpec(repoName string) batchv1.JobSpec {
	return batchv1.JobSpec{
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				// If there's an issue, delay till the next cron
				RestartPolicy: "Never",
				Containers: []corev1.Container{
					{
						Name:    "delete",
						Image:   repoSyncImage,
						Command: []string{"/chart-repo"},
						Args:    apprepoCleanupJobArgs(repoName),
						Env: []corev1.EnvVar{
							{
								Name: "MONGO_PASSWORD",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{Name: mongoSecretName},
										Key:                  "mongodb-root-password",
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
func jobLabels(apprepo *apprepov1alpha1.AppRepository) map[string]string {
	return map[string]string{
		"apprepositories.kubeapps.com/repo-name": apprepo.Name,
	}
}

// cronJobName returns a unique name for the CronJob managed by an AppRepository
func cronJobName(apprepo *apprepov1alpha1.AppRepository) string {
	return fmt.Sprintf("apprepo-sync-%s", apprepo.GetName())
}

// deleteJobName returns a unique name for the Job to cleanup AppRepository
func deleteJobName(reponame string) string {
	return fmt.Sprintf("apprepo-cleanup-%s", reponame)
}

// apprepoSyncJobArgs returns a list of args for the sync container
func apprepoSyncJobArgs(apprepo *apprepov1alpha1.AppRepository) []string {
	args := []string{
		"sync",
		"--mongo-url=" + mongoURL,
		"--mongo-user=root",
	}

	if userAgentComment != "" {
		args = append(args, "--user-agent-comment="+userAgentComment)
	}

	return append(args, apprepo.GetName(), apprepo.Spec.URL)
}

// apprepoSyncJobEnvVars returns a list of env variables for the sync container
func apprepoSyncJobEnvVars(apprepo *apprepov1alpha1.AppRepository) []corev1.EnvVar {
	var envVars []corev1.EnvVar
	envVars = append(envVars, corev1.EnvVar{
		Name: "MONGO_PASSWORD",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: mongoSecretName},
				Key:                  "mongodb-root-password",
			},
		},
	})
	if apprepo.Spec.Auth.Header != nil {
		envVars = append(envVars, corev1.EnvVar{
			Name: "AUTHORIZATION_HEADER",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &apprepo.Spec.Auth.Header.SecretKeyRef,
			},
		})
	}
	return envVars
}

// apprepoCleanupJobArgs returns a list of args for the repo cleanup container
func apprepoCleanupJobArgs(repoName string) []string {
	return []string{
		"delete",
		repoName,
		"--mongo-url=" + mongoURL,
		"--mongo-user=root",
	}
}
