/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2017 Intel Corporation

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
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/intelsdi-x/snap-client-go/client/plugins"
	"github.com/urfave/cli"
)

func loadPlugin(ctx *cli.Context) error {
	pAsc := ctx.String("plugin-asc")
	var paths []string
	if len(ctx.Args()) != 1 {
		return newUsageError("Incorrect usage:", ctx)
	}
	paths = append(paths, ctx.Args().First())
	if pAsc != "" {
		if !strings.Contains(pAsc, ".asc") {
			return newUsageError("Must be a .asc file for the -a flag", ctx)
		}
		paths = append(paths, pAsc)
	}

	params := plugins.NewLoadPluginParamsWithTimeout(FlTimeout.Value)
	f, err := os.Open(filepath.Join(paths...))
	if err != nil {
		return newUsageError("Cannot open the plugin", ctx)
	}
	defer f.Close()
	params.SetPluginData(f)

	resp, err := client.Plugins.LoadPlugin(params, authInfoWriter)
	if err != nil {
		return getErrorDetail(err, ctx)
	}

	p := resp.Payload
	fmt.Println("Plugin loaded")
	fmt.Printf("Name: %s\n", p.Name)
	fmt.Printf("Version: %d\n", p.Version)
	fmt.Printf("Type: %s\n", p.Type)
	fmt.Printf("Signed: %v\n", p.Signed)
	fmt.Printf("Loaded Time: %s\n\n", time.Unix(p.LoadedTimestamp, 0).Format(time.RFC1123))

	return nil
}

func unloadPlugin(ctx *cli.Context) error {
	if len(ctx.Args()) < 2 {
		return newUsageError("Incorrect usage:", ctx)
	}

	pType := ctx.Args().Get(0)
	pName := ctx.Args().Get(1)
	pVerStr := ctx.Args().Get(2)

	if pType == "" {
		return newUsageError("Must provide plugin type", ctx)
	}
	if pName == "" {
		return newUsageError("Must provide plugin name", ctx)
	}
	if pVerStr == "" {
		return newUsageError("Must provide plugin version", ctx)
	}

	pVer, err := strconv.Atoi(pVerStr)
	if err != nil {
		return newUsageError("Can't convert version string to integer", ctx)
	}
	if pVer < 1 {
		return newUsageError("Plugin version must be greater than zero", ctx)
	}

	params := plugins.NewUnloadPluginParamsWithTimeout(FlTimeout.Value)
	params.SetPname(pName)
	params.SetPtype(pType)
	params.SetPversion(int64(pVer))

	_, err = client.Plugins.UnloadPlugin(params, authInfoWriter)
	if err != nil {
		return getErrorDetail(err, ctx)
	}

	fmt.Println("Plugin unloaded")
	fmt.Printf("Name: %s\n", pName)
	fmt.Printf("Version: %d\n", pVer)
	fmt.Printf("Type: %s\n", pType)

	return nil
}

func listPlugins(ctx *cli.Context) error {
	running := ctx.Bool("running")
	params := plugins.NewGetPluginsParamsWithTimeout(FlTimeout.Value)
	if running {
		params.SetRunning(&running)
	}

	resp, err := client.Plugins.GetPlugins(params, authInfoWriter)
	if err != nil {
		return getErrorDetail(err, ctx)
	}

	lps := len(resp.Payload.Plugins)

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	if running {
		if lps == 0 {
			fmt.Println("No running plugins found. Have you started a task?")
			return nil
		}

		printFields(w, false, 0, "NAME", "HIT COUNT", "LAST HIT", "TYPE", "PPROF PORT")
		for _, rp := range resp.Payload.Plugins {
			printFields(w, false, 0, rp.Name, rp.HitCount, time.Unix(rp.LastHitTimestamp, 0).Format(time.RFC1123), rp.Type, rp.PprofPort)
		}
	} else {
		if lps == 0 {
			fmt.Println("No plugins found. Have you loaded a plugin?")
			return nil
		}
		printFields(w, false, 0, "NAME", "VERSION", "TYPE", "SIGNED", "STATUS", "LOADED TIME")
		for _, lp := range resp.Payload.Plugins {
			printFields(w, false, 0, lp.Name, lp.Version, lp.Type, lp.Signed, lp.Status, time.Unix(lp.LoadedTimestamp, 0).Format(time.RFC1123))
		}
	}
	w.Flush()

	return nil
}
