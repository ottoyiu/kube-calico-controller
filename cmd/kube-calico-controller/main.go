package main

import (
	"flag"
	"time"

	"github.com/golang/glog"

	caliclient "github.com/projectcalico/libcalico-go/lib/client"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"

	"github.com/ottoyiu/kube-calico-controller/pkg/controller"
)

// Based off of: https://github.com/kubernetes/client-go/blob/release-4.0/examples/workqueue/main.go
func main() {
	var kubeconfig string
	var master string
	var domainName string
	var syncSeconds int
	var dryRun bool

	flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	flag.StringVar(&master, "master", "", "master url")
	flag.StringVar(&domainName, "domain", "", "Domain name for nodes if Calico datastore stores Nodes as shortnames.")
	flag.IntVar(&syncSeconds, "syncperiod", 300, "Sync period between Kubernetes API and Calico datastore, in seconds.")
	flag.BoolVar(&dryRun, "dryrun", false, "dry-run mode. No changes will be made to the Calico datastore.")
	flag.Parse()

	if dryRun {
		glog.Info("Running in Dry-Run mode. No changes will be made to the Calico datastore.")
	}
	if domainName != "" {
		glog.Infof("Appending domain name %s to short hostnames to form FQDN for nodes in Calico datastore.", domainName)
	}
	glog.Infof("Sync period between Kubernetes API and Calico datastore is set to: %d seconds.", syncSeconds)

	// creates the Kubernetes connection config
	config, err := clientcmd.BuildConfigFromFlags(master, kubeconfig)
	if err != nil {
		glog.Fatal(err)
	}

	// creates the Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Fatal(err)
	}

	calicoConfig, err := caliclient.LoadClientConfig("")
	if err != nil {
		glog.Fatal(err)
	}
	calicoClient, err := caliclient.New(*calicoConfig)
	if err != nil {
		glog.Fatal(err)
	}

	// create the Node watcher
	nodeListWatcher := cache.NewListWatchFromClient(
		clientset.Core().RESTClient(),
		"nodes", v1.NamespaceAll, fields.Everything())

	// create the workqueue
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	// Bind the workqueue to a cache with the help of an informer. This way we make sure that
	// whenever the cache is updated, the Node key is added to the workqueue.
	// Note that when we finally process the item from the workqueue, we might see a newer version
	// of the Node than the version which was responsible for triggering the update.
	indexer, informer := cache.NewIndexerInformer(nodeListWatcher, &v1.Node{}, 60*time.Second, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// IndexerInformer uses a delta queue, therefore for deletes we have to use this
			// key function.
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
	}, cache.Indexers{})

	controller := controller.NewController(queue, indexer, informer, calicoClient, domainName, dryRun, syncSeconds)

	// Now let's start the controller
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	// Wait forever
	select {}
}
