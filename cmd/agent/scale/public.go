//
//  Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
//  This Source Code Form is subject to the terms of the Mozilla Public
//  License, v. 2.0. If a copy of the MPL was not distributed with this
//  file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package scale

import (
	"context"
	"fmt"

	"sort"

	"github.com/imdario/mergo"
	tcc "github.com/joyent/triton-go/compute"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/joyent/tsg-cli/cmd/config"
)

type AgentComputeClient struct {
	client *tcc.ComputeClient
}

func NewGetComputeClient(cfg *config.TritonClientConfig) (*AgentComputeClient, error) {
	computeClient, err := tcc.NewClient(cfg.Config)
	if err != nil {
		return nil, errors.Wrap(err, "Error Creating Triton Compute Client")
	}
	return &AgentComputeClient{
		client: computeClient,
	}, nil
}

func (c *AgentComputeClient) MaintainInstanceCount() error {
	instances, err := c.GetInstanceList()
	if err != nil {
		return err
	}

	runningInstances := len(instances)
	expectedInstances := config.GetExpectedMachineCount()
	scaleCount := expectedInstances - runningInstances

	if scaleCount < 0 {
		log.Log().Str("func", "MaintainInstanceCount").Msg("Scaling Instances Down")
		instancesToRemove := runningInstances - expectedInstances
		for i := 1; i < instancesToRemove+1; i++ {

			instance := instances[len(instances)-1]
			err := c.DeleteInstance(instance.ID)
			if err != nil {
				return err
			}

			instances = instances[:len(instances)-1]
		}

	} else if scaleCount > 0 {
		log.Log().Str("func", "MaintainInstanceCount").Msg("Scaling Instances Up")
		for i := 0; i < scaleCount; i++ {
			instance, err := c.CreateInstance()
			if err != nil {
				return err
			}

			instances = append(instances, instance)
		}
	} else {
		log.Log().Str("func", "MaintainInstanceCount").Msg("No-op")
	}

	return nil
}

func (c *AgentComputeClient) GetInstanceList() ([]*tcc.Instance, error) {
	params := &tcc.ListInstancesInput{}

	t := make(map[string]interface{}, 0)

	tsgName := config.GetTsgName()
	if tsgName != "" {
		t["tsg.name"] = tsgName
	}

	params.Tags = t

	instances, err := c.client.Instances().List(context.Background(), params)
	if err != nil {
		return nil, err
	}

	return sortInstances(instances), nil

}

func (c *AgentComputeClient) DeleteInstance(instanceID string) error {
	log.Log().Str("func", "DeleteInstance").Msg(fmt.Sprintf("Deleting Instance %q", instanceID))
	return c.client.Instances().Delete(context.Background(), &tcc.DeleteInstanceInput{
		ID: instanceID,
	})
}

func (c *AgentComputeClient) CreateInstance() (*tcc.Instance, error) {
	params := &tcc.CreateInstanceInput{
		FirewallEnabled: config.GetMachineFirewall(),
	}

	name := config.GetMachineName()
	if name != "" {
		params.Name = name
	}

	namePrefix := config.GetMachineNamePrefix()
	if namePrefix != "" {
		params.NamePrefix = namePrefix
	}

	md := make(map[string]string, 0)
	tags := make(map[string]string, 0)

	userdata := config.GetMachineUserdata()
	if userdata != "" {
		md["user-data"] = userdata
	}

	tsgName := config.GetTsgName()
	if tsgName != "" {
		tags["tsg.name"] = tsgName
	}

	networks := config.GetMachineNetworks()
	if len(networks) > 0 {
		params.Networks = networks
	}

	affinityRules := config.GetMachineAffinityRules()
	if len(affinityRules) > 0 {
		params.Affinity = affinityRules
	}

	t := config.GetMachineTags()
	if t != nil {
		mergo.Merge(&tags, t)
	}

	if tags != nil {
		params.Tags = tags
	}

	metadata := config.GetMachineMetadata()
	if metadata != nil {
		mergo.Merge(&md, metadata)
	}

	if len(md) > 0 {
		params.Metadata = md
	}

	pkgID := config.GetPkgID()
	if pkgID != "" {
		params.Package = pkgID
	}

	imgID := config.GetImgID()
	if imgID != "" {
		params.Image = imgID
	}

	machine, err := c.client.Instances().Create(context.Background(), params)
	if err != nil {
		return nil, err
	}

	log.Log().Str("func", "CreateInstance").Msg(fmt.Sprintf("Created Instance %q", machine.ID))
	return machine, nil
}

type instanceSort []*tcc.Instance

func sortInstances(instances []*tcc.Instance) []*tcc.Instance {
	sortInstances := instances
	sort.Sort(sort.Reverse(instanceSort(sortInstances)))
	return sortInstances
}

func (a instanceSort) Len() int {
	return len(a)
}

func (a instanceSort) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a instanceSort) Less(i, j int) bool {
	itime := a[i].Created
	jtime := a[j].Created
	return itime.Unix() < jtime.Unix()
}

