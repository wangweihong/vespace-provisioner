package vespace

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"vespace-provisioner/util/request"
)

type VespaceClient interface {
	//	VolumeInterface
	//	LoginInteface
	AuthInterface
	NamespaceInterface
	VolumeInterface
	CurrentCluster() ClusterInfo
	Debug()
}

type VolumeInterface interface {
	AddVolume(opt VolumeAddOption) error
	GetVolume(opt VolumeGetOption) (*Volume, error)
	DeleteVolume(opt VolumeGetOption) error
	UmountVolume(hosts []string, volgetopt VolumeGetOption) error
	MountVolume(hosts []string, volgetopt VolumeGetOption) error
}

type AuthInterface interface {
	RefreshToken() error
	Logout() error
}

//使用current cluster/default namespace/ default pool来创建卷
//default namespace/default pool一定存在.
type NamespaceInterface interface {
	ListNamespaces() ([]NamespaceInfo, error)
}

type vespaceClient struct {
	token          string
	username       string
	password       string
	master         string
	schema         string
	port           string
	currentCluster ClusterInfo
}

type AuthArgs struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

func NewVespaceClient(host string, opt AuthArgs) (VespaceClient, error) {
	vc := &vespaceClient{}
	vc.username = opt.UserName
	h := md5.New()
	h.Write([]byte(opt.Password))
	ciperStr := h.Sum(nil)
	vc.password = hex.EncodeToString(ciperStr)

	switch {
	case strings.HasPrefix(host, "http://"):
		ipport := strings.TrimPrefix(host, "http://")
		s := strings.Split(ipport, ":")
		switch len(s) {
		case 2:
			vc.master = s[0]
			vc.port = s[1]
		default:
			return nil, fmt.Errorf("Cannot parse host '%v'", host)
		}
		vc.schema = "http://"
	case strings.HasPrefix(host, "https://"):
		ipport := strings.TrimPrefix(host, "https://")
		s := strings.Split(ipport, ":")
		switch len(s) {
		case 2:
			vc.master = s[0]
			vc.port = s[1]
		default:
			return nil, fmt.Errorf("Cannot parse host '%v'", host)
		}
		vc.schema = "https://"

	default:
		s := strings.Split(host, ":")
		switch len(s) {
		case 2:
			vc.master = s[0]
			vc.port = s[1]
		default:
			return nil, fmt.Errorf("Cannot parse host '%v'", host)
		}
		vc.schema = "http://"
	}

	opt.UserName = vc.username
	opt.Password = vc.password
	err := vc.Login(opt)
	if err != nil {
		return nil, err
	}

	return vc, nil
}

type Response struct {
	Ecode   int             `json:"ecode"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type MasterCandidateInfo struct {
	Leader     string   `json:"leader"`
	Candidates []string `json:"candidates"`
	Ready      bool     `json:"ready"`
}

type UserInfo struct {
	Token          string      `json:"token"`
	CurrentCluster ClusterInfo `json:"currentcluster"`
}

type ClusterInfo struct {
	IsAuth bool   `json:"isauth"`
	Name   string `json:"name"`

	Uuid string `json:"uuid"`
}

func (vc *vespaceClient) Login(opt AuthArgs) error {
	urlpath := "/v1/authentication/login"
	resp, err := vc.httpPost(urlpath, opt)
	if err != nil {
		return err
	}

	var ui UserInfo
	switch resp.Ecode {
	case OK:
		err = json.Unmarshal([]byte(resp.Data), &ui)
		if err != nil {
			return err
		}

		vc.token = ui.Token
		vc.currentCluster = ui.CurrentCluster
	default:
		return fmt.Errorf("Vespace Error: %v", resp.Message)
	}

	return nil
}

func (vc *vespaceClient) Logout() error {
	urlpath := "/v1/authentication/logout"
	resp, err := vc.httpPost(urlpath, nil)
	if err != nil {
		return err
	}

	if resp.Ecode == OK {
		return nil
	}

	return fmt.Errorf("Vespace Error: %v" + resp.Message)
}

func (vc *vespaceClient) RefreshToken() error {
	urlpath := "/v1/authentication/refreshtoken"
	resp, err := vc.httpPost(urlpath, nil)
	if err != nil {
		return err
	}

	var ui UserInfo

	switch resp.Ecode {
	case OK:
		err = json.Unmarshal([]byte(resp.Data), &ui)
		if err != nil {
			return err
		}

		vc.token = ui.Token
	default:
		return fmt.Errorf("Vespace Error: %v", resp.Message)
	}
	return nil
}

type NamespaceList struct {
	List       []NamespaceInfo
	TotalCount int
}

type NamespaceInfo struct {
	Name          string `json:"name"`
	PoolCount     int    `json:"poolcount"`
	VolumeCount   int    `json:"volumecount"`
	StrategyCount int    `json:"strategycount"`
	Quota         int    `json:"quota"`
	Allocated     int    `json:"allocated"`
}

func (vc *vespaceClient) CurrentCluster() ClusterInfo {
	return vc.currentCluster
}

func (vc *vespaceClient) ListNamespaces() ([]NamespaceInfo, error) {

	urlpath := "/v1/namespace/list"
	resp, err := vc.httpGet(urlpath)
	if err != nil {
		return nil, err
	}

	switch resp.Ecode {
	case OK:
		var nl NamespaceList
		err := json.Unmarshal(resp.Data, &nl)
		if err != nil {
			return nil, err
		}
		return nl.List, nil

	default:
		return nil, fmt.Errorf("Vespace Error: %v", resp.Message)
	}

}
func (vc *vespaceClient) Debug() {
	fmt.Println("master:", vc.master)
	fmt.Println("user:", vc.username)
	fmt.Println("password:", vc.password)
	fmt.Println("token:", vc.token)
	fmt.Println("schema:", vc.schema)
	fmt.Println("port:", vc.port)
	fmt.Println("current cluster:", vc.currentCluster)
}

func (vc *vespaceClient) httpPost(urlpath string, data interface{}) (*Response, error) {
	url := vc.url(urlpath)

	httpResp, err := request.Post(url, vc.token, data)
	if err != nil {
		return nil, err
	}

	var resp Response
	var ci MasterCandidateInfo
checkResp:
	err = json.Unmarshal(httpResp, &resp)
	if err != nil {
		return nil, err
	}

	switch resp.Ecode {
	case ErrnoNotMaster:
		err = json.Unmarshal([]byte(resp.Data), &ci)
		if err != nil {
			return nil, err
		}

		vc.master = ci.Leader

		url = vc.url(urlpath)
		httpResp, err = request.Post(url, vc.token, data)
		if err != nil {
			return nil, err
		}
		goto checkResp
	}

	return &resp, nil
}

func (vc *vespaceClient) url(urlpath string) string {
	return vc.schema + vc.master + ":" + vc.port + urlpath
}

func (vc *vespaceClient) httpGet(urlpath string) (*Response, error) {
	url := vc.url(urlpath)

	httpResp, err := request.Get(url, vc.token)
	if err != nil {
		return nil, err
	}

	var resp Response
	var ci MasterCandidateInfo
checkResp:
	err = json.Unmarshal(httpResp, &resp)
	if err != nil {
		return nil, err
	}

	switch resp.Ecode {
	case ErrnoNotMaster:
		err = json.Unmarshal([]byte(resp.Data), &ci)
		if err != nil {
			return nil, err
		}

		vc.master = ci.Leader

		url = vc.url(urlpath)
		httpResp, err = request.Get(url, vc.token)
		if err != nil {
			return nil, err
		}
		goto checkResp
	}

	return &resp, nil
}
