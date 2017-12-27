package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	defaultTimeout = 15 * time.Second
)

func Get(url string, token string) ([]byte, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("New request for url [%v] fail for %v", url, err)
	}
	//测试使用
	//	request.Header.Add("token", "1234567890987654321")
	request.Header.Add("Authorization", "Bearer: "+token)

	client := http.Client{Timeout: defaultTimeout}
	resp, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("request url [%v] fail for %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("request url [%v] fail for %v", url, resp.Status)
		} else {
			return nil, fmt.Errorf("request url [%v] fail for %v", url, string(body))
		}
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse body from url [%v] request's body fail for %v", url, err)
	}
	return body, nil
}

func Post(url string, token string, data interface{}) ([]byte, error) {
	var bs []byte

	if data != nil {
		byteContent, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}

		bs = byteContent
	}
	req := bytes.NewBuffer(bs)

	request, err := http.NewRequest("POST", url, req)
	if err != nil {
		return nil, fmt.Errorf("New request for url [%v] fail for %v", url, err)
	}
	//测试使用
	//	request.Header.Add("token", "1234567890987654321")
	//request.Header.Add("token", token)
	request.Header.Add("Authorization", "Bearer: "+token)

	client := http.Client{Timeout: defaultTimeout}
	resp, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("request url [%v] fail for %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("request url [%v] fail for %v", url, resp.Status)
		} else {
			return nil, fmt.Errorf("request url [%v] fail for %v", url, string(body))
		}
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse body from url [%v] request's body fail for %v", url, err)
	}
	return body, nil
}
