package main

import (
	"flag"
	"vespace-provisioner/start"

	"github.com/golang/glog"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	provisioner = flag.String("provisioner", "youruncloud.com/vespace", "Name of the provisioner. The provisioner will only provision volumes for claims that request a StorageClass with a provisioner field set equal to this name.")
	master      = flag.String("master", "", "Master URL to build a client config from. Either this or kubeconfig needs to be set if the provisioner is being run out of cluster.")
	kubeconfig  = flag.String("kubeconfig", "", "Absolute path to the kubeconfig file. Either this or master needs to be set if the provisioner is being run out of cluster.")
	username    = flag.String("user", "", "username of vespace")
	password    = flag.String("password", "", "password of vespace")
	vespacehost = flag.String("vespace", "", "vespace host url")
)

func main() {
	flag.Set("logtostderr", "true")
	flag.Parse()
	nevercall := make(chan struct{})

	outOfCluster := *master != "" || *kubeconfig != ""
	var config *rest.Config
	var err error
	if outOfCluster {
		config, err = clientcmd.BuildConfigFromFlags(*master, *kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		glog.Fatalf("Failed to create config: %v", err)
	}

	if *username == "" || *password == "" || *vespacehost == "" {
		glog.Fatalf("Invalid vespace client login info.")
	}
	err = start.Init(*vespacehost, *username, *password, config, *provisioner, nevercall)
	if err != nil {
		glog.Fatal(err)
	}

	<-nevercall

}
