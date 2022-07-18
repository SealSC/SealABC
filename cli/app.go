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
	"github.com/SealSC/SealABC/cli/cliFlags"
	cliV2 "github.com/urfave/cli/v2"
	"io"
	"os"
)

func Run() (app *cliV2.App) {

	oldHelpPrinter := cliV2.HelpPrinter
	cliV2.HelpPrinter = func(w io.Writer, templ string, data interface{}) {
		oldHelpPrinter(w, templ, data)
		os.Exit(0)
	}

	oldVersionPrinter := cliV2.VersionPrinter
	cliV2.VersionPrinter = func(c *cliV2.Context) {
		oldVersionPrinter(c)
		os.Exit(0)
	}

	app = cliV2.NewApp()
	app.Name = "SealABC"
	app.Version = "0.1"
	app.HelpName = "SealABC"
	app.Usage = "SealABC"
	app.UsageText = "SealABC [options] [args]"
	app.HideHelp = false
	app.HideVersion = false

	cliFlags.SetFlags(app)
	SetAction(app)

	//run
	err := app.Run(os.Args)
	if nil != err {
		os.Exit(-1)
	}

	return
}
