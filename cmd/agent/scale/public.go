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
	"github.com/joyent/tsg-cli/cmd/config"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type AgentComputeClient struct {
	client *tcc.ComputeClient
}

func NewComputeClient(cfg *config.TritonClientConfig) (*AgentComputeClient, error) {
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
		instancesToRemove := runningInstances - expectedInstances
		for i := 1; i < instancesToRemove+1; i++ {
			instance := instances[len(instances)-1]

			err := c.DeleteInstance(instance.ID)
			if err != nil {
				log.Error().
					Str("account_name", c.client.Client.AccountName).
					Str("tsg_name", config.GetTsgName()).
					Str("status", "failed").
					Str("notification_type", "TSG_INSTANCE_TERMINATE_ERROR").
					Str("description", fmt.Sprintf("Error deleting instance %s", instance.ID)).
					Err(err)
				return err
			}

			log.Log().
				Str("account_name", c.client.Client.AccountName).
				Str("tsg_name", config.GetTsgName()).
				Str("status", "successful").
				Str("notification_type", "TSG_INSTANCE_TERMINATE").
				Str("description", fmt.Sprintf("Terminating instance %s", instance.ID)).
				Msgf("An instance was deleted due to a difference between the expected and actual instance count")

			instances = instances[:len(instances)-1]
		}

	} else if scaleCount > 0 {
		for i := 0; i < scaleCount; i++ {

			templateID := config.GetTsgTemplateID()

			instance, err := c.CreateInstance(templateID)
			if err != nil {
				log.Error().
					Str("account_name", c.client.Client.AccountName).
					Str("tsg_name", config.GetTsgName()).
					Str("status", "failed").
					Str("notification_type", "TSG_INSTANCE_LAUNCH_ERROR").
					Str("description", "Error launching new instance").
					Err(err)
				return err
			}

			err = c.TagInstance(instance.ID, templateID)
			if err != nil {
				log.Error().
					Str("account_name", c.client.Client.AccountName).
					Str("tsg_name", config.GetTsgName()).
					Str("status", "failed").
					Str("notification_type", "TSG_INSTANCE_LAUNCH_ERROR").
					Str("description", "Error launching new instance").
					Err(err)
				return err
			}

			log.Info().
				Str("account_name", c.client.Client.AccountName).
				Str("tsg_name", config.GetTsgName()).
				Str("status", "successful").
				Str("notification_type", "TSG_INSTANCE_LAUNCH").
				Str("description", fmt.Sprintf("Launching new instance %s", instance.ID)).
				Msgf("An instance was created due to a difference between the expected and actual instance count")

			instances = append(instances, instance)
		}
	} else {
		log.Info().
			Str("account_name", c.client.Client.AccountName).
			Str("tsg_name", config.GetTsgName()).
			Str("status", "successful").
			Str("notification_type", "TSG_INSTANCE_NO_OP").
			Str("description", fmt.Sprintf("Expected %d instances in TSG: %q - found %d instances", expectedInstances, config.GetTsgName(), runningInstances)).
			Msgf("TSG is healthy")
	}

	return nil
}

func (c *AgentComputeClient) TagInstance(instanceID string, templateID string) error {
	params := &tcc.AddTagsInput{
		ID: instanceID,
	}

	t := make(map[string]string, 0)
	t["name"] = formulateInstanceNameTag(templateID, instanceID)

	params.Tags = t

	err := c.client.Instances().AddTags(context.Background(), params)
	if err != nil {
		return err
	}

	return nil
}

func formulateInstanceNameTag(templateID string, instanceID string) string {
	return fmt.Sprintf("tsg-%s-%s", templateID[:8], instanceID[:8])
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
	return c.client.Instances().Delete(context.Background(), &tcc.DeleteInstanceInput{
		ID: instanceID,
	})
}

func (c *AgentComputeClient) CreateInstance(templateID string) (*tcc.Instance, error) {
	params := &tcc.CreateInstanceInput{
		FirewallEnabled: config.GetMachineFirewall(),
	}

	md := make(map[string]string, 0)
	tags := make(map[string]string, 0)
	tags["tsg.template"] = templateID

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
