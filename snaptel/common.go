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

	snapClient "github.com/intelsdi-x/snap-client-go/client"
	"github.com/intelsdi-x/snap-client-go/client/plugins"
	"github.com/intelsdi-x/snap-client-go/client/tasks"
	"github.com/urfave/cli"
)

var client *snapClient.Snap

// UsageError defines the error message and CLI context
type UsageError struct {
	s   string
	ctx *cli.Context
}

// Error prints the usage error
func (ue UsageError) Error() string {
	return fmt.Sprintf("Error: %s \nUsage: %s", ue.s, ue.ctx.Command.Usage)
}

// Help displays the command help
func (ue UsageError) Help() {
	cli.ShowCommandHelp(ue.ctx, ue.ctx.Command.Name)
}

func newUsageError(s string, ctx *cli.Context) UsageError {
	return UsageError{s, ctx}
}

// SetClient provides a way to set the private snapClient in this package.
func SetClient(cl *snapClient.Snap) {
	client = cl
}

// GetFirstChar gets the first character of a giving string.
func GetFirstChar(s string) string {
	firstChar := ""
	for _, r := range s {
		firstChar = fmt.Sprintf("%c", r)
		break
	}
	return firstChar
}

func getErrorDetail(err error, ctx *cli.Context) error {
	switch err.(type) {
	case *plugins.GetMetricsNotFound:
		return newUsageError(fmt.Sprintf("%v", err.(*plugins.GetMetricsNotFound).Payload.ErrorMessage), ctx)
	case *plugins.GetMetricsInternalServerError:
		return newUsageError(fmt.Sprintf("%v", err.(*plugins.GetMetricsInternalServerError).Payload.ErrorMessage), ctx)
	case *plugins.LoadPluginBadRequest:
		return newUsageError(fmt.Sprintf("%v", err.(*plugins.LoadPluginBadRequest).Payload.ErrorMessage), ctx)
	case *plugins.LoadPluginConflict:
		return newUsageError(fmt.Sprintf("%v", err.(*plugins.LoadPluginConflict).Payload.ErrorMessage), ctx)
	case *plugins.LoadPluginInternalServerError:
		return newUsageError(fmt.Sprintf("%v", err.(*plugins.LoadPluginInternalServerError).Payload.ErrorMessage), ctx)
	case *plugins.LoadPluginUnsupportedMediaType:
		return newUsageError(fmt.Sprintf("\n%v", err.(*plugins.LoadPluginUnsupportedMediaType).Payload.ErrorMessage), ctx)
	case *plugins.UnloadPluginBadRequest:
		return newUsageError(fmt.Sprintf("%v", err.(*plugins.UnloadPluginBadRequest).Payload.ErrorMessage), ctx)
	case *plugins.UnloadPluginConflict:
		return fmt.Errorf("%v", err.(*plugins.UnloadPluginConflict).Payload.ErrorMessage)
	case *plugins.UnloadPluginInternalServerError:
		return newUsageError(fmt.Sprintf("%v", err.(*plugins.UnloadPluginInternalServerError).Payload.ErrorMessage), ctx)
	case *plugins.UnloadPluginNotFound:
		return newUsageError(fmt.Sprintf("%v", err.(*plugins.UnloadPluginNotFound).Payload.ErrorMessage), ctx)
	case *plugins.GetPluginBadRequest:
		return newUsageError(fmt.Sprintf("%v", err.(*plugins.GetPluginBadRequest).Payload.ErrorMessage), ctx)
	case *plugins.GetPluginInternalServerError:
		return newUsageError(fmt.Sprintf("%v", err.(*plugins.GetPluginInternalServerError).Payload.ErrorMessage), ctx)
	case *plugins.GetPluginNotFound:
		return newUsageError(fmt.Sprintf("%v", err.(*plugins.GetPluginNotFound).Payload.ErrorMessage), ctx)
	case *plugins.GetPluginConfigItemBadRequest:
		return newUsageError(fmt.Sprintf("%v", err.(*plugins.GetPluginConfigItemBadRequest).Payload.ErrorMessage), ctx)
	case *tasks.GetTaskNotFound:
		return newUsageError(fmt.Sprintf("%v", err.(*tasks.GetTaskNotFound).Payload.ErrorMessage), ctx)
	case *tasks.AddTaskInternalServerError:
		return newUsageError(fmt.Sprintf("%v", err.(*tasks.AddTaskInternalServerError).Payload.ErrorMessage), ctx)
	case *tasks.UpdateTaskStateBadRequest:
		return newUsageError(fmt.Sprintf("%v", err.(*tasks.UpdateTaskStateBadRequest).Payload.ErrorMessage), ctx)
	case *tasks.UpdateTaskStateConflict:
		return newUsageError(fmt.Sprintf("%v", err.(*tasks.UpdateTaskStateConflict).Payload.ErrorMessage), ctx)
	case *tasks.UpdateTaskStateInternalServerError:
		return newUsageError(fmt.Sprintf("%v", err.(*tasks.UpdateTaskStateInternalServerError).Payload.ErrorMessage), ctx)
	default:
		return newUsageError(fmt.Sprintf("Error: %v", err), ctx)
	}
}
