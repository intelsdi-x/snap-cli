/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015-2017 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package snaptel

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"text/tabwriter"

	"github.com/intelsdi-x/snap-client-go/snap"
	"github.com/urfave/cli"
)

func getConfig(ctx *cli.Context) error {
	pDetails := filepath.SplitList(ctx.Args().First())
	var ptyp string
	var pname string
	var pver int
	var err error

	if len(pDetails) == 3 {
		ptyp = pDetails[0]
		pname = pDetails[1]
		pver, err = strconv.Atoi(pDetails[2])
		if err != nil {
			return newUsageError("Can't convert version string to integer", ctx)
		}
	} else {
		ptyp = ctx.String("plugin-type")
		pname = ctx.String("plugin-name")
		pver = ctx.Int("plugin-version")
	}

	if ptyp == "" {
		return newUsageError("Must provide plugin type", ctx)
	}
	if pname == "" {
		return newUsageError("Must provide plugin name", ctx)
	}
	if pver < 1 {
		return newUsageError("Must provide plugin version", ctx)
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	defer w.Flush()

	params := snap.NewGetPluginConfigItemParams()
	params.SetPtype(ptyp)
	params.SetPname(pname)
	params.SetPversion(int64(pver))

	resp, err := snapClient.GetPluginConfigItem(params)
	if err != nil {
		return fmt.Errorf("Error requesting plugin config %v", err.Error())
	}

	printFields(w, false, 0,
		"NAME",
		"VALUE",
		"TYPE",
	)

	cfg := resp.Payload.(map[string]interface{})
	cfgm := cfg["config"].(map[string]interface{})

	for k, v := range cfgm {
		printFields(w, false, 0, k, v, reflect.TypeOf(v))
	}
	return nil
}
