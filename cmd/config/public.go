//
//  Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
//  This Source Code Form is subject to the terms of the Mozilla Public
//  License, v. 2.0. If a copy of the MPL was not distributed with this
//  file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package config

import (
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/joyent/triton-go"
	"github.com/joyent/triton-go/authentication"
	"github.com/joyent/tsg-cli/cmd/internal/config"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type TritonClientConfig struct {
	Config *triton.ClientConfig
}

func New() (*TritonClientConfig, error) {
	viper.AutomaticEnv()

	var signer authentication.Signer
	var err error

	keyMaterial, err := GetTritonKeyMaterial()
	if err != nil {
		return nil, errors.Wrap(err, "error decoding key material from a Base64 encoded value")
	}

	if keyMaterial == "" {
		signer, err = authentication.NewSSHAgentSigner(authentication.SSHAgentSignerInput{
			KeyID:       GetTritonKeyID(),
			AccountName: GetTritonAccount(),
		})
		if err != nil {
			log.Error().Str("func", "initConfig").Msg(err.Error())
			return nil, err
		}
	} else {
		var keyBytes []byte
		if _, err = os.Stat(keyMaterial); err == nil {
			keyBytes, err = ioutil.ReadFile(keyMaterial)
			if err != nil {
				return nil, fmt.Errorf("error reading key material from %s: %s",
					keyMaterial, err)
			}
			block, _ := pem.Decode(keyBytes)
			if block == nil {
				return nil, fmt.Errorf(
					"failed to read key material '%s': no key found", keyMaterial)
			}

			if block.Headers["Proc-Type"] == "4,ENCRYPTED" {
				return nil, fmt.Errorf(
					"failed to read key '%s': password protected keys are\n"+
						"not currently supported. Please decrypt the key prior to use.", keyMaterial)
			}

		} else {
			keyBytes = []byte(keyMaterial)
		}

		signer, err = authentication.NewPrivateKeySigner(authentication.PrivateKeySignerInput{
			KeyID:              GetTritonKeyID(),
			PrivateKeyMaterial: keyBytes,
			AccountName:        GetTritonAccount(),
		})
		if err != nil {
			return nil, errors.Wrap(err, "Error Creating SSH Private Key Signer")
		}
	}

	config := &triton.ClientConfig{
		TritonURL:   GetTritonUrl(),
		AccountName: GetTritonAccount(),
		Signers:     []authentication.Signer{signer},
	}

	return &TritonClientConfig{
		Config: config,
	}, nil
}

func GetTritonUrl() string {
	return viper.GetString(config.KeyUrl)
}

func GetTritonKeyMaterial() (string, error) {
	value := viper.GetString(config.KeySshKeyMaterial)

	data, err := decodeBase64(value)
	if err != nil {
		return "", err
	}

	return data, nil
}

func GetTritonAccount() string {
	return viper.GetString(config.KeyAccount)
}

func GetTritonKeyID() string {
	return viper.GetString(config.KeySshKeyID)
}

func GetPkgID() string {
	return viper.GetString(config.KeyPackageId)
}

func GetImgID() string {
	return viper.GetString(config.KeyImageId)
}

func GetExpectedMachineCount() int {
	return viper.GetInt(config.KeyInstanceCount)
}

func GetTsgName() string {
	return viper.GetString(config.KeyTsgGroupName)
}

func GetTsgTemplateID() string {
	return viper.GetString(config.KeyTsgTemplateID)
}

func GetMachineFirewall() bool {
	return viper.GetBool(config.KeyInstanceFirewall)
}

func GetMachineNetworks() []string {
	if viper.IsSet(config.KeyInstanceNetwork) {
		var networks []string
		cfg := viper.GetStringSlice(config.KeyInstanceNetwork)
		for _, i := range cfg {
			networks = append(networks, i)
		}

		return networks
	}
	return nil
}

func GetMachineAffinityRules() []string {
	if viper.IsSet(config.KeyInstanceAffinityRule) {
		var rules []string
		cfg := viper.GetStringSlice(config.KeyInstanceAffinityRule)
		for _, i := range cfg {
			rules = append(rules, i)
		}

		return rules
	}
	return nil
}

func GetMachineTags() map[string]string {
	if viper.IsSet(config.KeyInstanceTag) {
		tags := make(map[string]string, 0)
		cfg := viper.GetStringSlice(config.KeyInstanceTag)
		for _, i := range cfg {
			m := strings.Split(i, "=")
			tags[m[0]] = m[1]
		}

		return tags
	}

	return nil
}

func GetMachineMetadata() (map[string]string, error) {
	if viper.IsSet(config.KeyInstanceMetadata) {
		metadata := make(map[string]string, 0)
		cfg := viper.GetStringSlice(config.KeyInstanceMetadata)
		for _, i := range cfg {
			data, err := decodeBase64(i)
			if err != nil {
				return nil, err
			}
			m := strings.Split(data, "=")
			metadata[m[0]] = m[1]
		}

		return metadata, nil
	}
	return nil, nil
}

func GetMachineUserdata() (string, error) {
	value := viper.GetString(config.KeyInstanceUserdata)

	data, err := decodeBase64(value)
	if err != nil {
		return "", err
	}
	return data, nil
}

func decodeBase64(s string) (string, error) {
	bytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err

	}
	return string(bytes), nil
}
