package kube

import (
	"fmt"
	"net/http"
	"os"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type PodPortForwardA struct {
	KubeClient *ClientS
	Namespace  string

	PodName string

	LocalPort int
	// PodPort is the target port for the pod
	PodPort int
	// Steams configures where to write or read input from
	streams genericclioptions.IOStreams
	// stopCh is the channel used to manage the port forward lifecycle
	stopCh chan struct{}
	// readyCh communicates when the tunnel is ready to receive traffic
	readyCh chan struct{}
}

func (req *PodPortForwardA) Forward() error {
	req.stopCh = make(chan struct{}, 1)
	req.readyCh = make(chan struct{})
	req.streams = genericclioptions.IOStreams{
		Out:    os.Stdout,
		ErrOut: os.Stderr,
		In:     os.Stdin,
	}
	url := req.KubeClient.API.RESTClient().Post().Resource("pods").Namespace(req.Namespace).Name(req.PodName).SubResource("portforward").Prefix("/api/v1").URL()
	transport, upgrader, err := spdy.RoundTripperFor(req.KubeClient.Config)
	if err != nil {
		return err
	}
	fmt.Println(url)

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, url)
	fw, err := portforward.New(dialer, []string{fmt.Sprintf("%d:%d", req.LocalPort, req.PodPort)}, req.stopCh, req.readyCh, req.streams.Out, req.streams.ErrOut)
	if err != nil {
		return err
	}
	return fw.ForwardPorts()
}

func (req *PodPortForwardA) Close() {
	close(req.stopCh)
}

// TODO fix this
func (req *PodPortForwardA) Ready() {
	<-req.readyCh
}
