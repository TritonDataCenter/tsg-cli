//
//  Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
//  This Source Code Form is subject to the terms of the Mozilla Public
//  License, v. 2.0. If a copy of the MPL was not distributed with this
//  file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package cmd

import (
	"github.com/joyent/tsg-cli/cmd/internal/command"
	"github.com/joyent/tsg-cli/cmd/internal/config"
	"github.com/joyent/tsg-cli/cmd/tsg-cli/cmd/scale"
	"github.com/sean-/conswriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var subCommands = []*command.Command{
	scale.Cmd,
}

var rootCmd = &command.Command{
	Cobra: &cobra.Command{
		Use:   "tsg",
		Short: "Joyent Triton Service Groups CLI",
	},
	Setup: func(parent *command.Command) error {
		{
			const (
				key          = config.KeyAccount
				longName     = "account"
				shortName    = "A"
				defaultValue = ""
				description  = "Account (login name). If not specified, the environment variable TRITON_ACCOUNT or SDC_ACCOUNT will be used"
			)

			flags := parent.Cobra.PersistentFlags()
			flags.StringP(longName, shortName, defaultValue, description)
			viper.BindPFlag(key, flags.Lookup(longName))
		}

		{
			const (
				key          = config.KeyUrl
				longName     = "url"
				shortName    = "U"
				defaultValue = ""
				description  = "CloudAPI URL. If not specified, the environment variable TRITON_URL or SDC_URL will be used"
			)

			flags := parent.Cobra.PersistentFlags()
			flags.StringP(longName, shortName, defaultValue, description)
			viper.BindPFlag(key, flags.Lookup(longName))
		}

		{
			const (
				key          = config.KeySshKeyID
				longName     = "key-id"
				shortName    = "K"
				defaultValue = ""
				description  = "This is the fingerprint of the public key matching the key specified in key_path. It can be obtained via the command ssh-keygen -l -E md5 -f /path/to/key. It can be provided via the SDC_KEY_ID or TRITON_KEY_ID environment variables."
			)

			flags := parent.Cobra.PersistentFlags()
			flags.StringP(longName, shortName, defaultValue, description)
			viper.BindPFlag(key, flags.Lookup(longName))
		}

		{
			const (
				key          = config.KeySshKeyMaterial
				longName     = "key-material"
				defaultValue = ""
				description  = "This is the private key of an SSH key associated with the Triton account to be used. If this is not set, the private key corresponding to the fingerprint in key_id must be available via an SSH Agent. It can be provided via the SDC_KEY_MATERIAL or TRITON_KEY_MATERIAL environment variables."
			)

			flags := parent.Cobra.PersistentFlags()
			flags.String(longName, defaultValue, description)
			viper.BindPFlag(key, flags.Lookup(longName))
		}

		return nil
	},
}

func Execute() error {

	rootCmd.Setup(rootCmd)

	conswriter.UsePager(false)
	//
	//if err := logger.Setup(); err != nil {
	//	return err
	//}

	for _, cmd := range subCommands {
		rootCmd.Cobra.AddCommand(cmd.Cobra)
		cmd.Setup(cmd)
	}

	if err := rootCmd.Cobra.Execute(); err != nil {
		return err
	}

	return nil
}
