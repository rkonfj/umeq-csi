package umeq

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

var httpCli http.Client = http.Client{}

type AgentService struct {
	AgentServer string
}

func NewAgentService(agentServer string) *AgentService {
	return &AgentService{
		AgentServer: agentServer,
	}
}

func (u *AgentService) UnpublishVolume(volumeId, nodeId string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/disk/%s/publish/%s", u.AgentServer, volumeId, nodeId), nil)
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

func (u *AgentService) PublishVolume(volumeId, nodeId string) error {
	res, err := http.Post(fmt.Sprintf("%s/disk/%s/publish/%s", u.AgentServer, volumeId, nodeId), "application/x-www-form-urlencoded", nil)
	if err != nil {
		return fmt.Errorf("publish disk err:%s", err.Error())
	}
	defer res.Body.Close()
	b, _ := ioutil.ReadAll(res.Body)
	if res.StatusCode != 200 {
		return errors.New(string(b))
	}
	log.Println("publishVolume resp:", string(b))
	return nil
}

func (u *AgentService) CreateVolume(volumeId string, requiredBytes int64) error {
	res, err := http.Post(fmt.Sprintf("%s/disk/%s/%d", u.AgentServer, volumeId, requiredBytes), "application/x-www-form-urlencoded", nil)
	if err != nil {
		return fmt.Errorf("create disk err:%s", err.Error())
	}
	defer res.Body.Close()
	b, _ := ioutil.ReadAll(res.Body)
	log.Println("createVolume resp:", string(b))
	return nil
}

func (u *AgentService) ExpandVolume(volumeId string, requiredBytes int64) error {
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/disk/%s/%d", u.AgentServer, volumeId, requiredBytes), nil)
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

func (u *AgentService) DeleteVolume(volumeId string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/disk/%s", u.AgentServer, volumeId), nil)
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

func (u *AgentService) GetDevPath(volumeId string) (string, error) {
	res, err := http.Get(fmt.Sprintf("%s/dev-path/%s", u.AgentServer, volumeId))
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	b, _ := ioutil.ReadAll(res.Body)
	log.Println("get dev-path resp:", string(b))
	return string(b), nil
}

func (u *AgentService) GetCapacity() (*Capacity, error) {
	res, err := http.Get(u.AgentServer + "/capacity")
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
