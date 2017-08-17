FROM golang:1.8-alpine

ADD bin/linux/kube-calico-controller /kube-calico-controller

CMD ["/kube-calico-controller"]

