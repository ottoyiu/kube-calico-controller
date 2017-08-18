package controller

import (
	"client-go/pkg/api"
	"client-go/tools/cache"
	"client-go/util/workqueue"
	"fmt"
	"time"

	calicache "github.com/caseydavenport/calico-node-controller/pkg/cache"
	"github.com/golang/glog"
	caliclient "github.com/projectcalico/libcalico-go/lib/client"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
)

type Controller struct {
	indexer      cache.Indexer
	queue        workqueue.RateLimitingInterface
	informer     cache.Controller
	calicoCache  calicache.ResourceCache
	calicoClient *caliclient.Client
}

func NewController(queue workqueue.RateLimitingInterface, indexer cache.Indexer, informer cache.Controller, calicoClient *caliclient.Client) *Controller {
	return &Controller{
		informer:     informer,
		indexer:      indexer,
		queue:        queue,
		calicoCache:  calicache.NewCache(),
		calicoClient: caliclient,
	}
}

func (c *Controller) processNextItem() bool {
	// Wait until there is a new item in the working queue
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	// Tell the queue that we are done with processing this key. This unblocks the key for other workers
	// This allows safe parallel processing because two pods with the same key are never processed in
	// parallel.
	defer c.queue.Done(key)

	// Invoke the method containing the business logic
	err := c.syncToCalico(key.(string))
	// Handle the error if something went wrong during the execution of the business logic
	c.handleErr(err, key)
	return true
}

// syncToCalico syncs the given update to Calico's etcd, as well as the in-memory cache
// of Calico objects for periodic syncs.
func (c *Controller) syncToCalico(key string) error {
	obj, exists, err := c.indexer.GetByKey(key)
	if err != nil {
		glog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		glog.Infof("Node %s does not exist anymore\n... attempting to find in Calico datastore", key)

		// Check if it exists in our cache.
		node, ok := c.calicoCache.Get(key)
		if ok {
			// If it does, then remove it.
			glog.Infof("Deleting stale node from calico datastore: %s\n", key)
			c.calicoCache.Delete(key)
			return c.calicoClient.Nodes().Delete(node.(api.Node).Metadata)
		}
		// Otherwise, this is a no-op.
		glog.Debugf("Node %s does not exist in Calico datastore... ignoring\n", key)
		return nil
	}
	return nil
}

// handleErr checks if an error happened and makes sure we will retry later.
func (c *Controller) handleErr(err error, key interface{}) {
	if err == nil {
		// Forget about the #AddRateLimited history of the key on every successful synchronization.
		// This ensures that future processing of updates for this key is not delayed because of
		// an outdated error history.
		c.queue.Forget(key)
		return
	}

	// This controller retries 5 times if something goes wrong. After that, it stops trying.
	if c.queue.NumRequeues(key) < 5 {
		glog.Infof("Error syncing node %v: %v", key, err)

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)
	// Report to an external entity that, even after several retries, we could not successfully process this key
	runtime.HandleError(err)
	glog.Infof("Dropping node %q out of the queue: %v", key, err)
}

func (c *Controller) Run(threadiness int, stopCh chan struct{}) {
	defer runtime.HandleCrash()

	// Let the workers stop when we are done
	defer c.queue.ShutDown()
	glog.Info("Starting Node controller")

	go c.informer.Run(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	c.populateCalicoCache()

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	<-stopCh
	glog.Info("Stopping Node controller")
}

func (c *Controller) runWorker() {
	for c.processNextItem() {
	}
}

func (c *Controller) populateCalicoCache() {
	// Populate the Calico cache.
	calicoNodes, err := c.calicoClient.Nodes().List(api.NodeMetadata{})
	if err != nil {
		panic(err)
	}
	for _, node := range calicoNodes.Items {
		name := node.Metadata.Name
		c.calicoCache.Set(name, node)
	}
}
