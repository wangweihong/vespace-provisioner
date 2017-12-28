package vespace

import (
	"encoding/json"
	"fmt"
)

const (
	defaultMountPort = 9888
)

type VolumeAttribute struct {
	ComponentShift  string
	DataType        string
	DevType         string
	DriveType       string
	Encrypto        string
	ReadBytesLimit  string
	ReadIOPSLimit   string
	Safety          string
	TargetACL       string
	ThinProvision   string
	WriteBytesLimit string
	WriteIOPSLimit  string
}

type VolumeAddOption struct {
	VolumeGetOption
	Capacity     string          `json:"capacity"`
	SnapCapacity string          `json:"snap_capacity"`
	Attribute    VolumeAttribute `json:"attribute"`
}

//添加并挂载
func (vc *vespaceClient) AddVolume(opt VolumeAddOption) error {
	urlpath := "/v1/volume/add"

	resp, err := vc.httpPost(urlpath, opt)
	if err != nil {
		return err
	}

	if resp.Ecode == OK {
		return nil
	}

	return fmt.Errorf("Vespace Error: Code:%v, Message:%v", resp.Ecode, resp.Message)
	//挂载iscsi

}

type MountOption struct {
	ClusterUuid string `json:"clusteruuid"`

	Name         string       `json:"name"` //卷名
	Namespace    string       `json:"namespace"`
	PoolName     string       `json:"poolname"`
	TargetedPort string       `json:"targetport"`
	Address      MountAddress `json:"address"`
	Attr         MountAttr    `json:"attr"`
}

type MountAddress struct {
	IP   string `json:"ip"`
	Port int    `json:"port"`
}

type MountAttr struct {
	NFSAcl  string
	NFSArgs string
}

type VolumeGetOption struct {
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	PoolName    string `json:"poolname"`
	ClusterUuid string `json:"clusteruuid"`
}

//如果有挂载先卸载
func (vc *vespaceClient) DeleteVolume(opt VolumeGetOption) error {
	/*
		vol, err := vc.GetVolume(opt)
		if err != nil {
			return nil
		}
	*/

	urlpath := "/v1/volume/delete"

	resp, err := vc.httpPost(urlpath, opt)
	if err != nil {
		return err
	}

	switch resp.Ecode {
	case OK:
		return nil
	default:
		return fmt.Errorf("Vespace Error: Code:%v, Message:%v", resp.Ecode, resp.Message)
	}
}

type MountHost struct {
	Ip   string `json:"ip"`
	Port int    `json:"port"`
}

type VolumeUmountOption struct {
	//	Attribute MountAttr `json:"attr"`
	Addresses []MountHost `json:"addresses"`
	VolumeGetOption
}
type VolumeMountOption struct {
	Attribute MountAttr   `json:"attr"`
	Addresses []MountHost `json:"addresses"`
	VolumeGetOption
}

func (vc *vespaceClient) UmountVolume(hosts []string, volgetopt VolumeGetOption) error {
	if len(hosts) == 0 {
		return nil
	}

	var vuo VolumeUmountOption
	vuo.VolumeGetOption = volgetopt
	vuo.Addresses = make([]MountHost, 0)
	for _, v := range hosts {
		var addr MountHost
		addr.Ip = v
		addr.Port = defaultMountPort
		vuo.Addresses = append(vuo.Addresses, addr)
	}

	urlpath := "/v1/volume/unmap"

	resp, err := vc.httpPost(urlpath, vuo)
	if err != nil {
		return err
	}

	if resp.Ecode == OK {
		return nil
	}

	return fmt.Errorf("Vespace Error: Code:%v, Message:%v", resp.Ecode, resp.Message)
}

func (vc *vespaceClient) MountVolume(hosts []string, volgetopt VolumeGetOption) error {
	if len(hosts) == 0 {
		return fmt.Errorf("need to provide host to mount")
	}

	var vuo VolumeMountOption
	vuo.VolumeGetOption = volgetopt

	vuo.Addresses = make([]MountHost, 0)
	for _, v := range hosts {
		var addr MountHost
		addr.Ip = v
		addr.Port = defaultMountPort
		vuo.Addresses = append(vuo.Addresses, addr)
	}
	vuo.Attribute.NFSAcl = "*"
	vuo.Attribute.NFSArgs = "rw@sync@no_root_squash"
	urlpath := "/v1/volume/map"

	resp, err := vc.httpPost(urlpath, vuo)
	if err != nil {
		return err
	}

	if resp.Ecode == OK {
		return nil
	}

	return fmt.Errorf("Vespace Error: Code:%v, Message:%v", resp.Ecode, resp.Message)
}

type Volume struct {
	Name            string          `json:"name"`
	Mounted         []string        `json:"mounted"`
	AccessPath      []string        `json:"accesspath"`
	ControllerHosts ControllerHosts `json:"controllerhosts"`
}

type ControllerHosts struct {
	Default map[string]bool `json:"default"`
}

func (vc *vespaceClient) GetVolume(opt VolumeGetOption) (*Volume, error) {
	urlpath := "/v1/volume/get" + "?name=" + opt.Name + "&namespace=" + opt.Namespace + "&poolname=" + opt.PoolName + "&clusteruuid=" + opt.ClusterUuid

	resp, err := vc.httpGet(urlpath)
	if err != nil {
		return nil, err
	}

	switch resp.Ecode {
	case OK:
		var vol Volume
		err := json.Unmarshal(resp.Data, &vol)
		if err != nil {
			return nil, err
		}

		return &vol, nil
	default:
		return nil, fmt.Errorf("Vespace Error: Code:%v, Message:%v", resp.Ecode, resp.Message)

	}
}
