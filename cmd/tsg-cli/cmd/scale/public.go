//
//  Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
//  This Source Code Form is subject to the terms of the Mozilla Public
//  License, v. 2.0. If a copy of the MPL was not distributed with this
//  file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package scale

import (
	"github.com/joyent/tsg-cli/cmd/agent/scale"
	tsgc "github.com/joyent/tsg-cli/cmd/config"
	"github.com/joyent/tsg-cli/cmd/internal/command"
	"github.com/joyent/tsg-cli/cmd/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var Cmd = &command.Command{
	Cobra: &cobra.Command{
		Args:         cobra.NoArgs,
		Use:          "scale",
		Short:        "scale triton service group",
		SilenceUsage: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := tsgc.New()
			if err != nil {
				return err
			}

			a, err := scale.NewComputeClient(c)
			if err != nil {
				return err
			}

			return a.MaintainInstanceCount()
		},
	},
	Setup: func(parent *command.Command) error {
		{
			const (
				key          = config.KeyTsgGroupName
				longName     = "tsg-name"
				defaultValue = ""
				description  = "TSG Name"
			)

			flags := parent.Cobra.Flags()
			flags.String(longName, defaultValue, description)
			viper.BindPFlag(key, flags.Lookup(longName))

			parent.Cobra.MarkFlagRequired(longName)
		}

		{
			const (
				key          = config.KeyTsgTemplateID
				longName     = "template-id"
				defaultValue = ""
				description  = "TSG Template ID"
			)

			flags := parent.Cobra.Flags()
			flags.String(longName, defaultValue, description)
			viper.BindPFlag(key, flags.Lookup(longName))

			parent.Cobra.MarkFlagRequired(longName)
		}

		{
			const (
				key          = config.KeyInstanceCount
				longName     = "count"
				shortName    = "c"
				defaultValue = ""
				description  = "Expected Instance Count"
			)

			flags := parent.Cobra.Flags()
			flags.StringP(longName, shortName, defaultValue, description)
			viper.BindPFlag(key, flags.Lookup(longName))

			parent.Cobra.MarkFlagRequired(longName)

		}

		{
			const (
				key         = config.KeyInstanceTag
				longName    = "tag"
				shortName   = "t"
				description = "Instance Tags. This flag can be used multiple times"
			)

			flags := parent.Cobra.Flags()
			flags.StringSliceP(longName, shortName, nil, description)
			viper.BindPFlag(key, flags.Lookup(longName))
		}

		{
			flags := parent.Cobra.PersistentFlags()
			flags.SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
				switch name {
				case "tag":
					name = "tags"
					break
				}

				return pflag.NormalizedName(name)
			})
		}

		{
			const (
				key          = config.KeyInstanceState
				longName     = "state"
				defaultValue = ""
				description  = "Instance state (e.g. running)"
			)

			flags := parent.Cobra.Flags()
			flags.String(longName, defaultValue, description)
			viper.BindPFlag(key, flags.Lookup(longName))
		}

		{
			const (
				key          = config.KeyPackageId
				longName     = "pkg-id"
				defaultValue = ""
				description  = "Package id (defaults to ''). This takes precedence over 'pkg-name'"
			)

			flags := parent.Cobra.Flags()
			flags.String(longName, defaultValue, description)
			viper.BindPFlag(key, flags.Lookup(longName))
		}

		{
			const (
				key          = config.KeyImageId
				longName     = "img-id"
				defaultValue = ""
				description  = "Image id (defaults to ''). This takes precedence over 'img-name'"
			)

			flags := parent.Cobra.Flags()
			flags.String(longName, defaultValue, description)
			viper.BindPFlag(key, flags.Lookup(longName))
		}

		{
			const (
				key          = config.KeyInstanceFirewall
				longName     = "firewall"
				defaultValue = false
				description  = "Enable Cloud Firewall on this instance (defaults to false)"
			)

			flags := parent.Cobra.Flags()
			flags.Bool(longName, defaultValue, description)
			viper.BindPFlag(key, flags.Lookup(longName))

			viper.SetDefault(key, defaultValue)
		}

		{
			const (
				key         = config.KeyInstanceNetwork
				longName    = "networks"
				shortName   = "N"
				description = "One or more comma-separated networks IDs. This option can be used multiple times."
			)

			flags := parent.Cobra.Flags()
			flags.StringSliceP(longName, shortName, nil, description)
			viper.BindPFlag(key, flags.Lookup(longName))
		}

		{
			const (
				key         = config.KeyInstanceMetadata
				longName    = "metadata"
				shortName   = "m"
				description = `Add metadata when creating the instance. Metadata are key/value
			       pairs available on the instance API object as the "metadata"
			       field, and inside the instance via the "mdata-*" commands. DATA
			       is one of: a "key=value" string (bool and numeric "value" are
				   converted to that type). This option can be used multiple times.`
			)

			flags := parent.Cobra.Flags()
			flags.StringSliceP(longName, shortName, nil, description)
			viper.BindPFlag(key, flags.Lookup(longName))
		}

		{
			const (
				key         = config.KeyInstanceAffinityRule
				longName    = "affinity"
				description = `Affinity rules for selecting a server for this instance. Rules
have one of the following forms: "instance==INST" (the new
instance must be on the same server as INST), "instance!=INST"
(new inst must *not* be on the same server as INST),
"instance==~INST"" (*attempt* to place on the same server as
INST), or "instance!=~INST" (*attempt* to place on a server
other than INST's). "INST" is an existing instance name or id.
There are two shortcuts: "inst" may be used instead of
"instance" and "instance==INST" can be shortened to just "INST".
This option can be used multiple times.`
			)

			flags := parent.Cobra.Flags()
			flags.StringSlice(longName, nil, description)
			viper.BindPFlag(key, flags.Lookup(longName))
		}

		{
			const (
				key          = config.KeyInstanceUserdata
				longName     = "userdata"
				defaultValue = ""
				description  = "A custom script which will be executed by the instance right after creation, and on every instance reboot."
			)

			flags := parent.Cobra.Flags()
			flags.String(longName, defaultValue, description)
			viper.BindPFlag(key, flags.Lookup(longName))
		}

		return nil
	},
}
