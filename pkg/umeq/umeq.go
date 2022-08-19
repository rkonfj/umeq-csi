package umeq

import (
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
	fmt.Println("UnpublishVolume REST Resp:", string(respBody))
	return nil
}

func publishVolume(volumeId, nodeId string) error {
	res, err := http.Post(fmt.Sprintf("http://192.168.3.11:8080/disk/%s/publish/%s", volumeId, nodeId), "application/x-www-form-urlencoded", nil)
	if err != nil {
		return fmt.Errorf("publish disk err:%s", err.Error())
	}
	defer res.Body.Close()
	b, _ := ioutil.ReadAll(res.Body)
	log.Println("real publish resp:", string(b))
	return nil
}

func createVolume(volumeId string, requiredBytes int64) error {
	res, err := http.Post(fmt.Sprintf("http://192.168.3.11:8080/disk/%s/%d", volumeId, requiredBytes), "application/x-www-form-urlencoded", nil)
	if err != nil {
		return fmt.Errorf("create disk err:%s", err.Error())
	}
	defer res.Body.Close()
	b, _ := ioutil.ReadAll(res.Body)
	log.Println("real create disk resp:", string(b))
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
	fmt.Println("deleteVolume REST Resp:", string(respBody))
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
