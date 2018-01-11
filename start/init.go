package start

import (
	"fmt"
	"strings"
	"vespace-provisioner/pkg/volume"

	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func Init(vespacehost string, username string, password string, config *rest.Config, provisioner string, stopChan chan struct{}) error {
	if errs := validateProvisioner(provisioner, field.NewPath("provisioner")); len(errs) != 0 {
		return fmt.Errorf("Invalid provisioner specified: %v", errs)
	}

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
	go pc.Run(stopChan)
	return nil

}

// validateProvisioner tests if provisioner is a valid qualified name.
// https://github.com/kubernetes/kubernetes/blob/release-1.4/pkg/apis/storage/validation/validation.go
func validateProvisioner(provisioner string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if len(provisioner) == 0 {
		allErrs = append(allErrs, field.Required(fldPath, provisioner))
	}
	if len(provisioner) > 0 {
		for _, msg := range validation.IsQualifiedName(strings.ToLower(provisioner)) {
			allErrs = append(allErrs, field.Invalid(fldPath, provisioner, msg))
		}
	}
	return allErrs
}
