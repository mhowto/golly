package consul

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type AgentServiceCheck struct {
	Script            string              `json:",omitempty"`
	DockerContainerID string              `json:",omitempty"`
	Shell             string              `json:",omitempty"` // Only supported for Docker.
	Interval          string              `json:",omitempty"`
	Timeout           string              `json:",omitempty"`
	TTL               string              `json:",omitempty"`
	HTTP              string              `json:",omitempty"`
	Header            map[string][]string `json:",omitempty"`
	Method            string              `json:",omitempty"`
	TCP               string              `json:",omitempty"`
	Status            string              `json:",omitempty"`
	Notes             string              `json:",omitempty"`
	TLSSkipVerify     bool                `json:",omitempty"`

	// In Consul 0.7 and later, checks that are associated with a service
	// may also contain this optional DeregisterCriticalServiceAfter field,
	// which is a timeout in the same Go time format as Interval and TTL. If
	// a check is in the critical state for more than this configured value,
	// then its associated service (and all of its associated checks) will
	// automatically be deregistered.
	DeregisterCriticalServiceAfter string `json:",omitempty"`
}

type AgentServiceChecks []*AgentServiceCheck

type AgentServiceRegistration struct {
	ID                string   `json:",omitempty"`
	Name              string   `json:",omitempty"`
	Tags              []string `json:",omitempty"`
	Port              int      `json:",omitempty"`
	Address           string   `json:",omitempty"`
	EnableTagOverride bool     `json:",omitempty"`
	Check             *AgentServiceCheck
	Checks            AgentServiceChecks
}

type AgentService struct {
	ID                string
	Service           string
	Tags              []string
	Port              int
	Address           string
	EnableTagOverride bool
	CreateIndex       uint64
	ModifyIndex       uint64
}

type ConsulClient struct {
	Address string
	c       *http.Client
}

func NewConsulClient(address string, client *http.Client) *ConsulClient {
	if client == nil {
		client = &http.Client{}
	}
	return &ConsulClient{
		Address: strings.TrimRight(address, "/"),
		c:       client,
	}
}

// RegisterService adds a new service, with an optional health check, to the agent.
func (cc *ConsulClient) RegisterService(service *AgentServiceRegistration) error {
	s, err := json.Marshal(service)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, cc.Address+"/agent/service/register", bytes.NewBuffer(s))
	if err != nil {
		return err
	}
	resp, err := cc.c.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("RegisterService's status code is not 200: %d", resp.StatusCode)
	}
	return nil
}

// DeregisterService removes a service from the agent. If the service does not exist, no action is taken.
func (cc *ConsulClient) DeregisterService(serviceID string) error {
	req, err := http.NewRequest(http.MethodPut, cc.Address+"/agent/service/deregister/"+serviceID, nil)
	if err != nil {
		return err
	}
	resp, err := cc.c.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("DeregisterService's status code is not 200: %d", resp.StatusCode)
	}
	return nil
}

// ListServices returns all the services that are registered with the local agent.
func (cc *ConsulClient) ListServices() (map[string]*AgentService, error) {
	resp, err := cc.c.Get(cc.Address + "/agent/services")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ListServices's status code is not 200: %d", resp.StatusCode)
	}

	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	services := make(map[string]*AgentService)
	err = json.Unmarshal(b, &services)
	if err != nil {
		return nil, err
	}
	return services, nil
}
