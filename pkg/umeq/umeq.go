package umeq

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

var httpCli http.Client = http.Client{}

func unpublishVolume(volumeId, nodeId string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("http://192.168.3.11:8080/disk/%s/publish/%s", volumeId, nodeId), nil)
	if err != nil {
		return err
	}

	// Fetch Request
	resp, err := httpCli.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read Response Body
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Println("unpublishVolume REST Resp:", string(respBody))
	return nil
}

func publishVolume(volumeId, nodeId string) error {
	res, err := http.Post(fmt.Sprintf("http://192.168.3.11:8080/disk/%s/publish/%s", volumeId, nodeId), "application/x-www-form-urlencoded", nil)
	if err != nil {
		return fmt.Errorf("publish disk err:%s", err.Error())
	}
	defer res.Body.Close()
	b, _ := ioutil.ReadAll(res.Body)
	log.Println("publishVolume resp:", string(b))
	return nil
}

func createVolume(volumeId string, requiredBytes int64) error {
	res, err := http.Post(fmt.Sprintf("http://192.168.3.11:8080/disk/%s/%d", volumeId, requiredBytes), "application/x-www-form-urlencoded", nil)
	if err != nil {
		return fmt.Errorf("create disk err:%s", err.Error())
	}
	defer res.Body.Close()
	b, _ := ioutil.ReadAll(res.Body)
	log.Println("createVolume resp:", string(b))
	return nil
}

func expandVolume(volumeId string, requiredBytes int64) error {
	req, err := http.NewRequest("PUT", fmt.Sprintf("http://192.168.3.11:8080/disk/%s/%d", volumeId, requiredBytes), nil)
	if err != nil {
		return err
	}

	// Fetch Request
	resp, err := httpCli.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read Response Body
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Println("expandVolume Resp:", string(respBody))
	return nil
}

func deleteVolume(volumeId string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("http://192.168.3.11:8080/disk/%s", volumeId), nil)
	if err != nil {
		return err
	}

	// Fetch Request
	resp, err := httpCli.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read Response Body
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Println("deleteVolume Resp:", string(respBody))
	return nil
}

func getDevPath(volumeId string) (string, error) {
	res, err := http.Get(fmt.Sprintf("http://192.168.3.11:8080/dev-path/%s", volumeId))
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	b, _ := ioutil.ReadAll(res.Body)
	log.Println("get dev-path resp:", string(b))
	return string(b), nil
}

func getCapacity() (*Capacity, error) {
	res, err := http.Get("http://192.168.3.11:8080/capacity")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	b, _ := ioutil.ReadAll(res.Body)
	log.Println("getCapacity resp:", string(b))
	var cap Capacity
	err = json.Unmarshal(b, &cap)
	return &cap, err
}

type Capacity struct {
	Available         int64
	MaximumVolumeSize int64
	MinimumVolumeSize int64
}
