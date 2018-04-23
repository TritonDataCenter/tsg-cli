//
//  Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
//  This Source Code Form is subject to the terms of the Mozilla Public
//  License, v. 2.0. If a copy of the MPL was not distributed with this
//  file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package main

import (
	"os"

	"github.com/joyent/tsg-cli/cmd/tsg-cli/cmd"
	"github.com/rs/zerolog/log"
	"github.com/sean-/conswriter"
)

func main() {
	defer func() {
		p := conswriter.GetTerminal()
		p.Wait()
	}()

	if err := cmd.Execute(); err != nil {
		//logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
		log.Error().Err(err).Msg("unable to run")
		os.Exit(1)
	}
}
