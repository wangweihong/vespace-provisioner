package volume

import (
	"fmt"
	"strings"
	"vespace-provisioner/pkg/vespace"

	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	provisionName = "ufleet.com/vespace"
)

func NewVespaceProvisioner(host string, username string, password string) (controller.Provisioner, error) {
	opt := vespace.AuthArgs{}
	opt.UserName = username
	opt.Password = password
	vc, err := vespace.NewVespaceClient(host, opt)
	if err != nil {
		return nil, err
	}

	return &vespaceProvisioner{
		VespaceClient: vc,
		//		pvVolumeMap:   make(map[string]string),
	}, nil

}

type vespaceProvisioner struct {
	VespaceClient vespace.VespaceClient
	pvVolumeMap   map[string]string
}

var _ controller.Provisioner = &vespaceProvisioner{}

func (p *vespaceProvisioner) createVolume(options controller.VolumeOptions) (string, string, error) {
	//创建一个以namespace-pvc
	vc := p.VespaceClient
	vgetopt := vespace.VolumeGetOption{}
	vgetopt.Name = options.PVName
	vgetopt.Namespace = "default"
	vgetopt.PoolName = "default"
	vgetopt.ClusterUuid = vc.CurrentCluster().Uuid

	var vaddopt vespace.VolumeAddOption
	vaddopt.VolumeGetOption = vgetopt
	cap := options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]
	vaddopt.Capacity = (&cap).String()
	vaddopt.SnapCapacity = vaddopt.Capacity
	var va vespace.VolumeAttribute
	va.ComponentShift = "27"
	va.DataType = "linear"
	va.DevType = "target"
	va.DriveType = "HDD"
	va.Encrypto = "off"
	va.ReadBytesLimit = "0"
	va.ReadIOPSLimit = "0"
	va.Safety = "first"
	va.TargetACL = "ALL"
	va.ThinProvision = "on"
	va.WriteBytesLimit = "0"
	va.WriteIOPSLimit = "0"
	vaddopt.Attribute = va

	err := vc.AddVolume(vaddopt)
	if err != nil {
		return "", "", err
	}

	vol, err := vc.GetVolume(vgetopt)
	if err != nil {
		return "", "", err
	}

	hosts := make([]string, 0)
	for k, _ := range vol.ControllerHosts.Default {
		hosts = append(hosts, k)
	}

	err = vc.MountVolume(hosts, vgetopt)
	if err != nil {
		return "", "", err
	}

	vol, err = vc.GetVolume(vgetopt)
	if err != nil {
		return "", "", err
	}

	if len(vol.AccessPath) == 0 {
		return "", "", fmt.Errorf("cannot get volume access path")
	}

	s := strings.Split(vol.AccessPath[0], " ")
	if len(s) != 2 {
		return "", "", fmt.Errorf("cannpt parse volume access path")
	}
	return s[0], s[1], nil
}

func (p *vespaceProvisioner) deleteVolume(volume *v1.PersistentVolume) error {
	vc := p.VespaceClient
	var vgetopt vespace.VolumeGetOption
	vgetopt.Name = volume.Name
	vgetopt.Namespace = "default"
	vgetopt.PoolName = "default"
	vgetopt.ClusterUuid = vc.CurrentCluster().Uuid

	vol, err := vc.GetVolume(vgetopt)
	if err != nil {
		return err
	}

	if vol.Mounted != nil {
		err := vc.UmountVolume(vol.Mounted, vgetopt)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	err = vc.DeleteVolume(vgetopt)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (p *vespaceProvisioner) Provision(options controller.VolumeOptions) (*v1.PersistentVolume, error) {

	portals, iqn, err := p.createVolume(options)
	if err != nil {
		return nil, err
	}
	lun := 1

	annotations := make(map[string]string)
	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:        options.PVName,
			Labels:      map[string]string{},
			Annotations: annotations,
		},
		Spec: v1.PersistentVolumeSpec{
			PersistentVolumeReclaimPolicy: options.PersistentVolumeReclaimPolicy,
			AccessModes:                   options.PVC.Spec.AccessModes,
			Capacity: v1.ResourceList{
				v1.ResourceName(v1.ResourceStorage): options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)],
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				ISCSI: &v1.ISCSIVolumeSource{
					//					TargetPortal: options.Parameters["targetPortal"],
					TargetPortal: portals,
					Portals:      []string{portals},
					IQN:          iqn,
					//					ISCSIInterface:    options.Parameters["iscsiInterface"],
					Lun: int32(lun),
					//					ReadOnly:          getReadOnly(options.Parameters["readonly"]),
					//					FSType:            getFsType(options.Parameters["fsType"]),
					//					DiscoveryCHAPAuth: getBool(options.Parameters["chapAuthDiscovery"]),
					//					SessionCHAPAuth:   getBool(options.Parameters["chapAuthSession"]),
					//					SecretRef:         getSecretRef(getBool(options.Parameters["chapAuthDiscovery"]), getBool(options.Parameters["chapAuthSession"]), &v1.LocalObjectReference{Name: viper.GetString("provisioner-name") + "-chap-secret"}),
				},
			},
		},
	}

	return pv, nil

}

func (p *vespaceProvisioner) Delete(vol *v1.PersistentVolume) error {
	err := p.deleteVolume(vol)
	return err

}

func getSize(options controller.VolumeOptions) int64 {
	q := options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]
	return q.Value()
}

func (p *vespaceProvisioner) getVolumeName(options controller.VolumeOptions) string {
	return options.PVName
}
