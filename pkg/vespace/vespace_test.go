package vespace

import (
	"fmt"
	"testing"
)

const (
	//	URL      = "http://192.168.16.56:20001"
	URL      = "http://192.168.3.62:8081"
	USERNAME = "admin"
	PASSWORD = "admin"
)

func TestVMSummary(t *testing.T) {
	var authargs AuthArgs
	authargs.Password = PASSWORD
	authargs.UserName = USERNAME
	vc, err := NewVespaceClient(URL, authargs)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	vc.Debug()
	t.Log(vc.CurrentCluster())

	var vgetopt VolumeGetOption
	vgetopt.Name = "test2"
	vgetopt.Namespace = "default"
	vgetopt.PoolName = "default"
	vgetopt.ClusterUuid = vc.CurrentCluster().Uuid

	fmt.Println("test volume .....")
	vol, err := vc.GetVolume(vgetopt)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	fmt.Println(vol)

	if vol.Mounted != nil {
		err := vc.UmountVolume(vol.Mounted, vgetopt)
		if err != nil {
			t.Log(err)
			t.Fail()
		}
	}
	err = vc.DeleteVolume(vgetopt)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	/*
		var vaddopt VolumeAddOption
		vaddopt.VolumeGetOption = vgetopt
		vaddopt.Capacity = "128M"
		vaddopt.SnapCapacity = "128M"

		var va VolumeAttribute
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
		err = vc.AddVolume(vaddopt)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		fmt.Println("create volume success")
		//	var vgetopt VolumeGetOption
		vgetopt.Name = "test2"
		vgetopt.Namespace = "default"
		vgetopt.PoolName = "default"
		vgetopt.ClusterUuid = vc.CurrentCluster().Uuid
		t.Log(vgetopt)

		fmt.Println("test volume Get.....")
		vol, err = vc.GetVolume(vgetopt)
		if err != nil {
			t.Log(err)
			t.Fail()
		}
		fmt.Println(vol)

		fmt.Println("test volume Mount.....")
		hosts := make([]string, 0)
		for k, _ := range vol.ControllerHosts.Default {
			hosts = append(hosts, k)
		}

		err = vc.MountVolume(hosts, vgetopt)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		fmt.Println("test volume Get.....")
		vol, err = vc.GetVolume(vgetopt)
		if err != nil {
			t.Log(err)
			t.Fail()
		}
		fmt.Println(vol)
	*/

}
