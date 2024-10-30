// Copyright 2021-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"
	"reflect"
	"time"

	apprepov1alpha1 "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	clientset "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/client/clientset/versioned"
	appreposcheme "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/client/clientset/versioned/scheme"
	informers "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/client/informers/externalversions"
	listers "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/client/listers/apprepository/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	batchlisters "k8s.io/client-go/listers/batch/v1"
	batchlistersv1beta1 "k8s.io/client-go/listers/batch/v1beta1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	log "k8s.io/klog/v2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// controllerAgentName is the name of the AppRepository controller
	controllerAgentName = "apprepository-controller"

	// finalizerName is the name of the finalizer added to each AppRepository to prevent deletions without clean-up
	finalizerName = "apprepositories.kubeapps.com/apprepo-cleanup-finalizer"

	// errResourceExists is used as part of the Event 'reason' when an AppRepository fails
	//  to sync due to a CronJob of the same name already existing.
	errResourceExists = "ErrResourceExists"

	// messageResourceExists is the message used for Events when a resourcefails to sync due to a CronJob already existing
	messageResourceExists = "Resource %q already exists and is not managed by AppRepository"

	// messageResourceSynced is the message used for an Event fired when an AppRepository is synced successfully
	messageResourceSynced = "AppRepository synced successfully"

	// messageSuccessSynced is used as part of the Event 'reason' when an AppRepository is synced
	messageSuccessSynced = "Synced"

	// Name of the the custom resource
	AppRepository   = "AppRepository"
	AppRepositories = "AppRepositories"
)

// Controller is the controller implementation for AppRepository resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// apprepoclientset is the clientset for AppRepository resources
	apprepoclientset clientset.Interface

	cronjobsLister        batchlisters.CronJobLister
	cronjobsListerv1beta1 batchlistersv1beta1.CronJobLister
	cronjobsSynced        cache.InformerSynced
	appreposLister        listers.AppRepositoryLister
	appreposSynced        cache.InformerSynced

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

	// obtain references to shared index informers for the AppRepository type.
	apprepoInformer := apprepoInformerFactory.Kubeapps().V1alpha1().AppRepositories()

	// Create event broadcaster
	// Add apprepository-controller types to the default Kubernetes Scheme so
	// Events can be logged for apprepository-controller types.
	err := appreposcheme.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Errorf("Unable to create new controller: %v", err)
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
		appreposLister:   apprepoInformer.Lister(),
		appreposSynced:   apprepoInformer.Informer().HasSynced,
		workqueue:        workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), AppRepositories),
		recorder:         recorder,
		conf:             *conf,
	}

	log.Info("Setting up event handlers")
	// Set up an event handler for when AppRepository resources change
	_, err = apprepoInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueAppRepo,
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldApp := oldObj.(*apprepov1alpha1.AppRepository)
			newApp := newObj.(*apprepov1alpha1.AppRepository)
			if !reflect.DeepEqual(oldApp.Spec, newApp.Spec) {
				controller.enqueueAppRepo(newApp)
			} else if !reflect.DeepEqual(oldApp.ObjectMeta, newApp.ObjectMeta) {
				// handle updates in ObjectMeta (like finalizers)
				controller.handleAppRepoMetaChangeOrDelete(newApp, false)
			}
		},
		DeleteFunc: func(obj interface{}) {
			controller.handleAppRepoMetaChangeOrDelete(obj, true)
		},
	})
	if err != nil {
		log.Warningf("Error adding AppRepository event handler: %v", err)
	}

	controller.setBatchLister(conf.V1Beta1CronJobs, kubeInformerFactory)

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

// setBatchListener sets the specific batch listener based on the config, ie. either
// a v1 batch listener or a v1beta1 batch listener.
//
// This allows users on 1.20 to continue to use the latest release.
func (c *Controller) setBatchLister(useV1Beta1 bool, kubeInformerFactory kubeinformers.SharedInformerFactory) {
	// Set up an event handler for when CronJob resources get deleted. This
	// handler will lookup the owner of the given CronJob, and if it is owned by a
	// AppRepository resource will enqueue that AppRepository resource for
	// processing so the CronJob gets correctly recreated. This way, we don't need
	// to implement custom logic for handling CronJob resources. More info on this
	// pattern:
	// https://github.com/kubernetes/community/blob/8cafef897a22026d42f5e5bb3f104febe7e29830/contributors/devel/controllers.md
	var err error
	if useV1Beta1 {
		cronjobInformer := kubeInformerFactory.Batch().V1beta1().CronJobs()
		c.cronjobsListerv1beta1 = cronjobInformer.Lister()
		c.cronjobsSynced = cronjobInformer.Informer().HasSynced
		_, err = cronjobInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
			DeleteFunc: c.handleAppRepoOwnedObject,
		})
	} else {
		cronjobInformer := kubeInformerFactory.Batch().V1().CronJobs()
		c.cronjobsLister = cronjobInformer.Lister()
		c.cronjobsSynced = cronjobInformer.Informer().HasSynced
		_, err = cronjobInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
			DeleteFunc: c.handleAppRepoOwnedObject,
		})
	}

	if err != nil {
		log.Fatalf("Error adding CronJob event handler: %v", err)
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
			return fmt.Errorf("error syncing %q: %s", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		log.Infof("Successfully synced %q", key)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

// handleAppRepoOwnedObject will take any resource implementing metav1.Object and attempt to
// find the AppRepository resource that 'owns' it. It does this by looking at
// the objects metadata.ownerReferences field for an appropriate OwnerReference.
// It then enqueues that AppRepository resource to be processed. If the object
// does not have an appropriate OwnerReference, it will simply be skipped.
func (c *Controller) handleAppRepoOwnedObject(obj interface{}) {
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
		log.Infof("Recovered deleted object %q from tombstone", object.GetName())
	}
	log.Infof("Processing object: %s", object.GetName())
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		// If this object is not owned by an AppRepository, we should not do
		// anything more with it.
		if ownerRef.Kind != AppRepository {
			return
		}

		apprepo, err := c.appreposLister.AppRepositories(object.GetNamespace()).Get(ownerRef.Name)
		if err != nil {
			log.Infof("Ignoring orphaned object %q of AppRepository %q", object.GetSelfLink(), ownerRef.Name)
			return
		}

		if apprepo.ObjectMeta.DeletionTimestamp != nil {
			apprepoKey := fmt.Sprintf("%s/%s", apprepo.GetNamespace(), apprepo.GetName())
			log.Infof("Ignoring AppRepository %q, it was already marked for deletion at timestamp %q", apprepoKey, apprepo.ObjectMeta.DeletionTimestamp)
			return
		}

		c.enqueueAppRepo(apprepo)
		return
	}
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the AppRepository
// resource with the current status of the resource.
func (c *Controller) syncHandler(key string) error {
	ctx := context.Background()

	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the AppRepository resource with this namespace/name
	apprepo, err := c.appreposLister.AppRepositories(namespace).Get(name)
	if err != nil {
		// return nil if the error is a not-found as we want to let the flow continue
		return ctrlclient.IgnoreNotFound(err)
	}

	cronjob, err := c.ensureCronJob(ctx, apprepo)
	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// if the object is not being deleted, check the finalizers
	if apprepo.GetDeletionTimestamp().IsZero() {
		// check if it contains the finalizer, if not, add it and update the object
		if !containsFinalizer(apprepo, finalizerName) {
			log.Infof("The AppRepository %q doesn't have a finalizer yet, adding one...", key)

			ok := addFinalizer(apprepo, finalizerName)
			if !ok {
				return fmt.Errorf("error adding finalizer %q to the AppRepository %q", finalizerName, key)
			}
			_, err = c.apprepoclientset.KubeappsV1alpha1().AppRepositories(namespace).Update(ctx, apprepo, metav1.UpdateOptions{})
			if err != nil {
				return fmt.Errorf("error updating the AppRepository %q: %v", key, err)
			}
		}
		// if the object is being deleted and already contains a finalizer
		// it gets handled by this func instead of handleAppRepoMetaChangeOrDelete
	} else if containsFinalizer(apprepo, finalizerName) {
		err = c.handleCleanupFinalizer(ctx, apprepo)
		if err != nil {
			log.Errorf("Error handling the finalizer %q of the AppRepository %q: %v", finalizerName, key, err)
			return err
		}
	}

	// If the CronJob is not controlled by this AppRepository resource and it is not a
	// cronjob for an app repo in another namespace, then we should
	// log a warning to the event recorder and return it.
	if !metav1.IsControlledBy(cronjob, apprepo) && !objectBelongsTo(cronjob, apprepo) {
		msg := fmt.Sprintf(messageResourceExists, cronjob.GetName())
		c.recorder.Event(apprepo, corev1.EventTypeWarning, errResourceExists, msg)
		return fmt.Errorf("%s", msg)
	}

	if apprepo.GetNamespace() == c.conf.KubeappsNamespace {
		c.recorder.Event(apprepo, corev1.EventTypeNormal, messageSuccessSynced, messageResourceSynced)
	}
	return nil
}

// ensureCronJob ensures that the cronjob exists and is up-to-date.
//
// It looks after creating either a v1 cronjob or v1beta1 cronjob depending on
// configuration.
func (c *Controller) ensureCronJob(ctx context.Context, apprepo *apprepov1alpha1.AppRepository) (metav1.Object, error) {
	apprepoKey := fmt.Sprintf("%s/%s", apprepo.GetNamespace(), apprepo.GetName())

	// Get the cronjob with the same name as AppRepository
	cronjobName := cronJobName(apprepo.GetNamespace(), apprepo.GetName(), false)

	var cronjob metav1.Object
	var getCronJobErr error

	if c.conf.V1Beta1CronJobs {
		cronjob, getCronJobErr = c.cronjobsListerv1beta1.CronJobs(c.conf.KubeappsNamespace).Get(cronjobName)
	} else {
		cronjob, getCronJobErr = c.cronjobsLister.CronJobs(c.conf.KubeappsNamespace).Get(cronjobName)
	}

	cronJobSpec, err := newCronJob(apprepo, c.conf)
	if err != nil {
		log.Errorf("Unable to create CronJob spec for AppRepository %q: %v", apprepoKey, err)
		return nil, err
	}

	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(getCronJobErr) {
		log.Infof("Creating CronJob %q in namespace %q for AppRepository %q", cronJobSpec.GetName(), c.conf.KubeappsNamespace, apprepoKey)
		if c.conf.V1Beta1CronJobs {
			cronjob, err = c.kubeclientset.BatchV1beta1().CronJobs(c.conf.KubeappsNamespace).Create(ctx, v1CronJobToV1Beta1CronJob(cronJobSpec), metav1.CreateOptions{})
		} else {
			cronjob, err = c.kubeclientset.BatchV1().CronJobs(c.conf.KubeappsNamespace).Create(ctx, cronJobSpec, metav1.CreateOptions{})
		}
		if err != nil {
			log.Errorf("Unable to create CronJob %q for AppRepository %q: %v", cronJobSpec.GetName(), apprepoKey, err)
			return nil, err
		}

		// Trigger a manual Job for the initial sync
		job := newSyncJob(apprepo, c.conf)
		log.Infof("Creating Job %q in namespace %q for starting syncing AppRepository %q", job.GetGenerateName(), c.conf.KubeappsNamespace, apprepoKey)
		_, err = c.kubeclientset.BatchV1().Jobs(c.conf.KubeappsNamespace).Create(ctx, job, metav1.CreateOptions{})
		if err != nil {
			log.Errorf("Unable to create Job %q for AppRepository %q: %v", job.GetGenerateName(), apprepoKey, err)
			return nil, err
		}
	} else if err == nil {
		// If the resource already exists, we'll update it
		log.Infof("Updating CronJob %q in namespace %q for AppRepository %q", cronJobSpec.GetName(), c.conf.KubeappsNamespace, apprepoKey)
		if c.conf.V1Beta1CronJobs {
			cronjob, err = c.kubeclientset.BatchV1beta1().CronJobs(c.conf.KubeappsNamespace).Update(ctx, v1CronJobToV1Beta1CronJob(cronJobSpec), metav1.UpdateOptions{})
		} else {
			cronjob, err = c.kubeclientset.BatchV1().CronJobs(c.conf.KubeappsNamespace).Update(ctx, cronJobSpec, metav1.UpdateOptions{})
		}
		if err != nil {
			log.Errorf("Unable to update CronJob %q for AppRepository %q: %v", cronJobSpec.GetName(), apprepoKey, err)
			return nil, err
		}

		// The AppRepository has changed, launch a manual Job
		job := newSyncJob(apprepo, c.conf)
		log.Infof("Creating Job %q in namespace %q for starting syncing AppRepository %q", job.GetGenerateName(), c.conf.KubeappsNamespace, apprepoKey)
		_, err = c.kubeclientset.BatchV1().Jobs(c.conf.KubeappsNamespace).Create(ctx, job, metav1.CreateOptions{})
		if err != nil {
			log.Errorf("Unable to create Job %q for AppRepository %q: %v", job.GetGenerateName(), apprepoKey, err)
			return nil, err
		}
	}
	return cronjob, nil
}

// handleCleanupFinalizer starts the clean-up tasks derived from the finalizer of the AppRepository
// and removes the finalizer from the object, updating the object afterwards
func (c *Controller) handleCleanupFinalizer(ctx context.Context, apprepo *apprepov1alpha1.AppRepository) error {
	apprepoKey := fmt.Sprintf("%s/%s", apprepo.GetNamespace(), apprepo.GetName())

	log.Infof("Starting the clean-up tasks derived from the finalizer of the AppRepository %q", apprepoKey)

	// start the clean-up
	err := c.cleanUpAppRepo(ctx, apprepo)
	if err != nil {
		log.Errorf("Error performing clean-up tasks derived from the finalizer of the AppRepository %q. The finalizer will be removed anyways. You might want to perform a manual clean-up. Error: %v", apprepoKey, err)
	}

	// once everything is done, remove the finalizer from the list
	ok := removeFinalizer(apprepo, finalizerName)
	if !ok {
		return fmt.Errorf("error removing finalizer from the AppRepository %q: %v", apprepoKey, err)
	}

	// update the CR removing the finalizer
	log.Infof("Removing finalizer from the AppRepository %q", apprepoKey)
	_, err = c.apprepoclientset.KubeappsV1alpha1().AppRepositories(apprepo.GetNamespace()).Update(ctx, apprepo, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("error updating the AppRepository %q: %v", apprepoKey, err)
	}

	return nil
}

func (c *Controller) cleanUpAppRepo(ctx context.Context, apprepo *apprepov1alpha1.AppRepository) error {
	apprepoKey := fmt.Sprintf("%s/%s", apprepo.GetNamespace(), apprepo.GetName())

	// Trigger a Job to perform the cleanup of the charts in the DB corresponding to deleted AppRepository
	job := newCleanupJob(apprepo, c.conf)
	log.Infof("Creating clean-up Job %q in namespace %q for deleting AppRepository %q", job.GetGenerateName(), c.conf.KubeappsNamespace, apprepoKey)
	_, err := c.kubeclientset.BatchV1().Jobs(c.conf.KubeappsNamespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		log.Errorf("Unable to create clean-up Job %q for AppRepository %q: %v", job.GetGenerateName(), apprepoKey, err)
		return err
	}

	// TODO: Workaround until the sync jobs are moved to the repoNamespace (#1647)
	// Delete the cronjob in the Kubeapps namespace to avoid re-syncing the repository
	cronJobName := cronJobName(apprepo.GetNamespace(), apprepo.GetName(), false)
	log.Infof("Deleting the CronJob %q in namespace %q for deleting AppRepository %q", cronJobName, c.conf.KubeappsNamespace, apprepoKey)
	err = c.kubeclientset.BatchV1().CronJobs(c.conf.KubeappsNamespace).Delete(ctx, cronJobName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		log.Errorf("Unable to delete CronJob %q for AppRepository %q: %v", cronJobName, apprepoKey, err)
		return err
	}
	log.Infof("The clean-up tasks on AppRepository %q succeeded", apprepoKey)

	return nil
}

func (c *Controller) handleAppRepoMetaChangeOrDelete(obj interface{}, shouldDelete bool) {
	var (
		apprepo *apprepov1alpha1.AppRepository
		ok      bool
		err     error
	)

	ctx := context.Background()

	if apprepo, ok = obj.(*apprepov1alpha1.AppRepository); !ok {
		runtime.HandleError(fmt.Errorf("error decoding object, invalid type"))
		return
	}

	apprepoKey := fmt.Sprintf("%s/%s", apprepo.GetNamespace(), apprepo.GetName())

	// if the object is not being deleted (ie, deletionTimestamp==0)
	if apprepo.GetDeletionTimestamp().IsZero() && !containsFinalizer(apprepo, finalizerName) {
		// check if it contains the finalizer, if not, add it and update the object
		log.Errorf("The AppRepository %q should be deleted, but doesn't have any finalizers. You might want to perform a manual clean-up", apprepoKey)
	}

	// if the object is being deleted and contains a finalizer
	if !apprepo.GetDeletionTimestamp().IsZero() && containsFinalizer(apprepo, finalizerName) {
		// if the object is being deleted and contains a finalizer and the event is not a deletion event,
		// then handle the finalizer-derived clean-up tasks and remove the finalizer
		if !shouldDelete {
			err = c.handleCleanupFinalizer(ctx, apprepo)
			if err != nil {
				log.Errorf("Error handling the finalizer of the AppRepository %q: %v", apprepoKey, err)
				return
			}
		} else {
			// if the object is being deleted and it is a deletion event, then do nothing else
			log.Infof("The AppRepository %q is now being deleted", apprepoKey)
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				c.workqueue.AddRateLimited(key)
			}
		}
	}
}
