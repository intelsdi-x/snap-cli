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
	"strings"
	"text/tabwriter"

	"github.com/urfave/cli"
)

var (
	// Commands defines a list of Snap CLI commands.
	Commands = []cli.Command{
		{
			Name: "task",
			Subcommands: []cli.Command{
				{
					Name:        "create",
					Description: "Creates a new task in the snap scheduler",
					Usage:       "There are two ways to create a task.\n\t1) Use a task manifest with [--task-manifest]\n\t2) Provide a workflow manifest and schedule details.\n\n\t* Note: Start, stop date/time, and count are optional.\n\t* Using `task create -h` to see options.\n",
					Action:      createTask,
					Flags: []cli.Flag{
						flTaskManifest,
						flWorkfowManifest,
						flTaskSchedInterval,
						flTaskSchedCount,
						flTaskSchedStartDate,
						flTaskSchedStartTime,
						flTaskSchedStopDate,
						flTaskSchedStopTime,
						flTaskName,
						flTaskSchedDuration,
						flTaskSchedNoStart,
						flTaskDeadline,
						flTaskMaxFailures,
					},
				},
				{
					Name:   "list",
					Usage:  "list or list --verbose",
					Action: listTask,
					Flags: []cli.Flag{
						flTaskManifest,
						flWorkfowManifest,
						flTaskSchedInterval,
						flTaskSchedCount,
						flTaskSchedStartDate,
						flTaskSchedStartTime,
						flTaskSchedStopDate,
						flTaskSchedStopTime,
						flTaskName,
						flTaskSchedDuration,
						flTaskSchedNoStart,
						flTaskDeadline,
						flTaskMaxFailures,
					},
				},
				{
					Name:   "start",
					Usage:  "start <task_id>",
					Action: startTask,
				},
				{
					Name:   "stop",
					Usage:  "stop <task_id>",
					Action: stopTask,
				},
				{
					Name:   "remove",
					Usage:  "remove <task_id>",
					Action: removeTask,
				},
				{
					Name:   "export",
					Usage:  "export <task_id>",
					Action: exportTask,
				},
				{
					Name:   "watch",
					Usage:  "watch <task_id> or watch <task_id> --verbose",
					Action: watchTask,
					Flags: []cli.Flag{
						flVerbose,
					},
				},
				{
					Name:   "enable",
					Usage:  "enable <task_id>",
					Action: enableTask,
				},
			},
		},
		{
			Name: "plugin",
			Subcommands: []cli.Command{
				{
					Name:   "load",
					Usage:  "load <plugin_path> [--plugin-cert=<plugin_cert_path> --plugin-key=<plugin_key_path> --plugin-ca-certs=<ca_cert_paths>]",
					Action: loadPlugin,
					Flags: []cli.Flag{
						flPluginAsc,
						flPluginCert,
						flPluginKey,
						flPluginCACerts,
					},
				},
				{
					Name:   "unload",
					Usage:  "unload <plugin_type> <plugin_name> <plugin_version>",
					Action: unloadPlugin,
				},
				{
					Name:   "list",
					Usage:  "list or list --running",
					Action: listPlugins,
					Flags: []cli.Flag{
						flRunning,
					},
				},
				{
					Name: "config",
					Subcommands: []cli.Command{
						{
							Name:   "get",
							Usage:  "get <plugin_type>:<plugin_name>:<plugin_version> or get -t <plugin_type> -n <plugin_name> -v <plugin_version>",
							Action: getConfig,
							Flags: []cli.Flag{
								flPluginName,
								flPluginType,
								flPluginVersion,
							},
						},
					},
				},
			},
		},
		{
			Name: "metric",
			Subcommands: []cli.Command{
				{
					Name:   "list",
					Usage:  "list",
					Action: listMetrics,
					Flags: []cli.Flag{
						flMetricVersion,
						flMetricNamespace,
						flVerbose,
					},
				},
				{
					Name:   "get",
					Usage:  `get -m <namespace> or get -m <namespace> -v <version>`,
					Action: getMetric,
					Flags: []cli.Flag{
						flMetricVersion,
						flMetricNamespace,
					},
				},
			},
		},
	}
)

func printFields(tw *tabwriter.Writer, indent bool, width int, fields ...interface{}) {
	var argArray []interface{}
	if indent {
		argArray = append(argArray, strings.Repeat(" ", width))
	}
	for i, field := range fields {
		if field != nil {
			argArray = append(argArray, field)
		} else {
			argArray = append(argArray, "")
		}
		if i < (len(fields) - 1) {
			argArray = append(argArray, "\t")
		}
	}
	fmt.Fprintln(tw, argArray...)
}
