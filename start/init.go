package start

import (
	"fmt"
	"vespace-provisioner/pkg/volume"

	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func Init(vespacehost string, username string, password string, config *rest.Config, provisioner string) error {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("Failed to create client: %v", err)
	}

	serverVersion, err := clientset.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("Error getting server version: %v", err)
	}

	vespaceProvisioner, err := volume.NewVespaceProvisioner(vespacehost, username, password)
	if err != nil {
		return fmt.Errorf("Error create vespace provisioner: %v", err)
	}
	pc := controller.NewProvisionController(
		clientset,
		provisioner,
		vespaceProvisioner,
		serverVersion.GitVersion,
	)
	go pc.Run(wait.NeverStop)
	return nil

}
