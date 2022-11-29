// Copyright 2022 rkonfj@fnla.io
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package umeq

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Println("[info] unpublishVolume", string(respBody))
	return nil
}

func (u *AgentService) PublishVolume(volumeId, nodeId string) error {
	res, err := http.Post(fmt.Sprintf("%s/disk/%s/publish/%s",
		u.AgentServer, volumeId, nodeId), "application/x-www-form-urlencoded", nil)
	if err != nil {
		return fmt.Errorf("publish disk err:%s", err.Error())
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	if res.StatusCode != 200 {
		return errors.New(string(b))
	}
	log.Println("[info] publishVolume", string(b))
	return nil
}

func (u *AgentService) CreateVolume(kind string, volumeId string, requiredBytes int64) error {
	res, err := http.Post(fmt.Sprintf("%s/kind/%s/disk/%s/%d",
		u.AgentServer, kind, volumeId, requiredBytes),
		"application/x-www-form-urlencoded", nil)
	if err != nil {
		return fmt.Errorf("create disk err:%s", err.Error())
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("[error] createVolume failed, caused %s", b)
	}
	log.Println("[info] createVolume", string(b))
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
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Println("[info] expandVolume", string(respBody))
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
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Println("[info] deleteVolume", string(respBody))
	return nil
}

func (u *AgentService) GetDevPath(volumeId string) (string, error) {
	res, err := http.Get(fmt.Sprintf("%s/dev-path/%s", u.AgentServer, volumeId))
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("get dev path error: %s", b)
	}
	log.Println("[info] devicePath", string(b))
	return string(b), nil
}

func (u *AgentService) GetCapacity() (*Capacity, error) {
	res, err := http.Get(u.AgentServer + "/capacity")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	log.Println("[info] GetCapacity", string(b))
	var cap Capacity
	err = json.Unmarshal(b, &cap)
	return &cap, err
}

func (u *AgentService) Probe() error {
	res, err := http.Get(u.AgentServer + "/probe")
	if err != nil {
		return fmt.Errorf("may be agent server not started yet? %w", err)
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	log.Println("[info] Probe", string(b))
	return nil
}

type Capacity struct {
	Available         int64
	MaximumVolumeSize int64
	MinimumVolumeSize int64
}
