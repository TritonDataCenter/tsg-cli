//
//  Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
//  This Source Code Form is subject to the terms of the Mozilla Public
//  License, v. 2.0. If a copy of the MPL was not distributed with this
//  file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package config

const (
	KeyAccount        = "general.account"
	KeyUrl            = "general.url"
	KeySshKeyMaterial = "general.key-material"
	KeySshKeyID       = "general.key-id"

	KeyTsgGroupName = "compute.tsg.name"

	KeyInstanceName         = "compute.instance.name"
	KeyInstanceCount        = "compute.instance.count"
	KeyInstanceNamePrefix   = "compute.instance.name-prefix"
	KeyInstanceFirewall     = "compute.instance.firewall"
	KeyInstanceState        = "compute.instance.state"
	KeyInstanceBrand        = "compute.instance.brand"
	KeyInstanceNetwork      = "compute.instance.networks"
	KeyInstanceTag          = "compute.instance.tag"
	KeyInstanceMetadata     = "compute.instance.metadata"
	KeyInstanceAffinityRule = "compute.instance.affinity"
	KeyInstanceUserdata     = "compute.instance.userdata"

	KeyPackageId = "compute.package.id"

	KeyImageId = "compute.image.id"
)

