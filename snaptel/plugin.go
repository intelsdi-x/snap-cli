/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

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
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/intelsdi-x/snap-client-go/client/operations"
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

	params := operations.NewLoadPluginParams()
	f, err := os.Open(strings.Join(paths, "/"))
	if err != nil {
		fmt.Printf("No plugin to load: %v", err)
	}
	defer f.Close()
	params.SetPluginData(f)

	resp, err := getOperationsClient().LoadPlugin(params)
	if err != nil {
		return fmt.Errorf("Error loading plugin:\n%v", err.Error())
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
	pType := ctx.Args().Get(0)
	pName := ctx.Args().Get(1)
	pVer, err := strconv.Atoi(ctx.Args().Get(2))

	if pType == "" {
		return newUsageError("Must provide plugin type", ctx)
	}
	if pName == "" {
		return newUsageError("Must provide plugin name", ctx)
	}
	if err != nil {
		return newUsageError("Can't convert version string to integer", ctx)
	}
	if pVer < 1 {
		return newUsageError("Must provide plugin version", ctx)
	}

	if getOperationsClient() == nil {
		return errors.New(errNoClient)
	}

	params := operations.NewUnloadPluginParams()
	params.SetPname(pName)
	params.SetPtype(pType)
	params.SetPversion(int64(pVer))

	_, err = getOperationsClient().UnloadPlugin(params)
	if err != nil {
		return fmt.Errorf("Error unloading plugin:\n%v", err.Error())
	}

	fmt.Println("Plugin unloaded")
	fmt.Printf("Name: %s\n", pName)
	fmt.Printf("Version: %d\n", pVer)
	fmt.Printf("Type: %s\n", pType)

	return nil
}

func listPlugins(ctx *cli.Context) error {
	if getOperationsClient() == nil {
		return errors.New(errNoClient)
	}

	params := operations.NewGetPluginsParams()
	resp, err := getOperationsClient().GetPlugins(params)
	if err != nil {
		return fmt.Errorf("Error: %v", err.Error())
	}

	if len(resp.Payload.Plugins) == 0 {
		fmt.Println("No plugins found. Have you loaded a plugin?")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	printFields(w, false, 0, "NAME", "VERSION", "TYPE", "SIGNED", "STATUS", "LOADED TIME")
	for _, lp := range resp.Payload.Plugins {
		printFields(w, false, 0, lp.Name, lp.Version, lp.Type, lp.Signed, lp.Status, time.Unix(lp.LoadedTimestamp, 0).Format(time.RFC1123))
	}
	w.Flush()

	return nil
}
