# kube-calico-controller
[![Build Status](https://travis-ci.org/ottoyiu/kube-calico-controller.svg?branch=master)](https://travis-ci.org/ottoyiu/kube-calico-controller) [![Go Report Card](https://goreportcard.com/badge/github.com/ottoyiu/kube-calico-controller)](https://goreportcard.com/report/github.com/ottoyiu/kube-calico-controller)

**WARNING** - this controller is still a Work In Progress. It should only be used for testing or development purposes. Please use extreme caution when using this in production.

A Kubernetes controller that ensures Kubernetes nodes and Calico's etcd datstore remains in-sync. This prevents stale nodes from lingering in etcd leading to slow route programming when a new node joins the Kubernetes cluster. This is especially useful when running a cluster in AWS, where an autoscaling group commonly creates nodes with new host names and delete hosts and their hostnames no longer exist.

This Controller is unnecessary for Calico deployments where the datasource is set to Kubernetes instead of etcd. At the time of writing this README, the Kubernetes datastore driver in Calico 2.4 is still not feature parity with the etcd datastore driver which may prevent users from switching to it. This is where the controller can come in handy.

[This issue](https://github.com/kubernetes/kops/issues/3224) describes the problem in detail of what this controller tries to solve.

Heavily inspired/based on this prototype by Casey Davenport: [caseydavenport/calico-node-controller](https://github.com/caseydavenport/calico-node-controller)

## Quick Start

## Requirements
- Kubernetes cluster with Calico as the network provider, using etcd as the datasource.
- Kubernetes >= 1.6
- Calico >= 2.0

## Kops integration
Integration progress can be tracked using [this Issue](https://github.com/kubernetes/kops/issues/3224).

## Usage
```
Usage of bin/linux/kube-calico-controller:
  -alsologtostderr
        log to standard error as well as files
  -domain string
        Domain name for nodes if Calico datastore stores Nodes as shortnames.
  -dryrun
        dry-run mode. No changes will be made to the Calico datastore.
  -kubeconfig string
        absolute path to the kubeconfig file
  -log_backtrace_at value
        when logging hits line file:N, emit a stack trace
  -log_dir string
        If non-empty, write log files in this directory
  -logtostderr
        log to standard error instead of files
  -master string
        master url
  -stderrthreshold value
        logs at or above this threshold go to stderr
  -syncperiod int
        Sync period between Kubernetes API and Calico datastore, in seconds. (default 300)
  -v value
        log level for V logs
  -vmodule value
        comma-separated list of pattern=N settings for file-filtered logging

```
Specifying the verbosity level of logging to 4 using the -v flag will get debug level output.

You only need to specify the location to kubeconfig using the `-kubeconfig` flag if you are running the controller out of the cluster for development and testing purpose.

The etcd cluster used by Calico as the datastore must be set as an environmental variable.

### Environnmental Variables
Variable                       | Description
------------------------------ | ----------
`CALICO_ETCD_ENDPOINTS`        | Etcd Endpoints string (eg. "http://etcd-1.internal.cluster:4001,http://etcd-2.internal.cluster:4001,http://etcd-3.internal.cluster:4001") - *required*
