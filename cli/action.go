/*
 * Copyright 2020 The SealABC Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package cli

import (
	"SealABC/cli/cliFlags"
	"errors"
	cliV2 "github.com/urfave/cli/v2"
)

func SetAction(app *cliV2.App) {
    app.Action = func(c *cliV2.Context) error {
		cfgFile := c.String(cliFlags.Config)

		//config file
		if "" == cfgFile {
			return errors.New("must set config file")
		}

		Parameters.ConfigFile = cfgFile

		return nil
	}
}
