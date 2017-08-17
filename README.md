# kube-calico-controller
[![Build Status](https://travis-ci.org/ottoyiu/kube-calico-controller.svg?branch=master)](https://travis-ci.org/ottoyiu/kube-calico-controller) [![Go Report Card](https://goreportcard.com/badge/github.com/ottoyiu/kube-calico-controller)](https://goreportcard.com/report/github.com/ottoyiu/kube-calico-controller)

**WARNING** - this controller is still a Work In Progress. It should only be used for testing or development purposes. Please use extreme caution when using this in production.

A Kubernetes controller that ensures Kubernetes nodes and Calico's etcd datstore remains in-sync. This prevents stale nodes from lingering in etcd leading to slow route programming when a new node joins the Kubernetes cluster. This is especially useful when running a cluster in AWS, where an autoscaling group commonly creates nodes with new host names and delete hosts and their hostnames no longer exist.

This Controller is unnecessary for Calico deployments where the datasource is set to Kubernetes instead of etcd. At the time of writing this README, the Kubernetes datastore driver in Calico 2.4 is still not feature parity with the etcd datastore driver which may prevent users from switching to it. This is where the controller can come in handy.

[https://github.com/kubernetes/kops/issues/3224](This issue) describes the problem in detail of what this controller tries to solve.

## Quick Start

## Requirements
- Kubernetes cluster with Calico as the network provider, using etcd as the datasource.
- Kubernetes >= 1.6
- Calico >= 2.0

## Kops integration
Integration progress can be tracked using [https://github.com/kubernetes/kops/issues/3224](this issue).

## Usage

### Environnmental Variables
