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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"golang.org/x/crypto/ssh/terminal"

	log "github.com/Sirupsen/logrus"
	"github.com/go-openapi/runtime"
	openapiclient "github.com/go-openapi/runtime/client"
	snapClient "github.com/intelsdi-x/snap-client-go/client"
	"github.com/intelsdi-x/snap-client-go/client/plugins"
	"github.com/intelsdi-x/snap-client-go/client/tasks"
	"github.com/urfave/cli"
)

var (
	client         *snapClient.Snap
	authInfoWriter runtime.ClientAuthInfoWriter
	password       string
	scheme         string
)

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

// SetClient sets the private HTTP Client in this package.
func SetClient(cl *snapClient.Snap) {
	client = cl
}

// SetAuthInfo sets the runtime ClientAuthInfoWriter.
func SetAuthInfo(aw runtime.ClientAuthInfoWriter) {
	authInfoWriter = aw
}

// SetScheme sets the request protocol.
func SetScheme(s string) {
	scheme = s
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
	case *plugins.GetMetricsUnauthorized:
		return newUsageError(fmt.Sprintf("%v", err.(*plugins.GetMetricsUnauthorized).Payload.Message), ctx)
	case *plugins.LoadPluginBadRequest:
		return newUsageError(fmt.Sprintf("%v", err.(*plugins.LoadPluginBadRequest).Payload.ErrorMessage), ctx)
	case *plugins.LoadPluginConflict:
		return newUsageError(fmt.Sprintf("%v", err.(*plugins.LoadPluginConflict).Payload.ErrorMessage), ctx)
	case *plugins.LoadPluginInternalServerError:
		return newUsageError(fmt.Sprintf("%v", err.(*plugins.LoadPluginInternalServerError).Payload.ErrorMessage), ctx)
	case *plugins.LoadPluginUnsupportedMediaType:
		return newUsageError(fmt.Sprintf("%v", err.(*plugins.LoadPluginUnsupportedMediaType).Payload.ErrorMessage), ctx)
	case *plugins.LoadPluginUnauthorized:
		return newUsageError(fmt.Sprintf("%v", err.(*plugins.LoadPluginUnauthorized).Payload.Message), ctx)
	case *plugins.UnloadPluginBadRequest:
		return newUsageError(fmt.Sprintf("%v", err.(*plugins.UnloadPluginBadRequest).Payload.ErrorMessage), ctx)
	case *plugins.UnloadPluginConflict:
		return fmt.Errorf("%v", err.(*plugins.UnloadPluginConflict).Payload.ErrorMessage)
	case *plugins.UnloadPluginInternalServerError:
		return newUsageError(fmt.Sprintf("%v", err.(*plugins.UnloadPluginInternalServerError).Payload.ErrorMessage), ctx)
	case *plugins.UnloadPluginNotFound:
		return newUsageError(fmt.Sprintf("%v", err.(*plugins.UnloadPluginNotFound).Payload.ErrorMessage), ctx)
	case *plugins.UnloadPluginUnauthorized:
		return newUsageError(fmt.Sprintf("%v", err.(*plugins.UnloadPluginUnauthorized).Payload.Message), ctx)
	case *plugins.GetPluginBadRequest:
		return newUsageError(fmt.Sprintf("%v", err.(*plugins.GetPluginBadRequest).Payload.ErrorMessage), ctx)
	case *plugins.GetPluginInternalServerError:
		return newUsageError(fmt.Sprintf("%v", err.(*plugins.GetPluginInternalServerError).Payload.ErrorMessage), ctx)
	case *plugins.GetPluginNotFound:
		return newUsageError(fmt.Sprintf("%v", err.(*plugins.GetPluginNotFound).Payload.ErrorMessage), ctx)
	case *plugins.GetPluginUnauthorized:
		return newUsageError(fmt.Sprintf("%v", err.(*plugins.GetPluginUnauthorized).Payload.Message), ctx)
	case *plugins.GetPluginsUnauthorized:
		return newUsageError(fmt.Sprintf("%v", err.(*plugins.GetPluginsUnauthorized).Payload.Message), ctx)
	case *plugins.GetPluginConfigItemBadRequest:
		return newUsageError(fmt.Sprintf("%v", err.(*plugins.GetPluginConfigItemBadRequest).Payload.ErrorMessage), ctx)
	case *plugins.GetPluginConfigItemUnauthorized:
		return newUsageError(fmt.Sprintf("%v", err.(*plugins.GetPluginConfigItemUnauthorized).Payload.Message), ctx)
	case *tasks.GetTaskNotFound:
		return newUsageError(fmt.Sprintf("%v", err.(*tasks.GetTaskNotFound).Payload.ErrorMessage), ctx)
	case *tasks.GetTaskUnauthorized:
		return newUsageError(fmt.Sprintf("%v", err.(*tasks.GetTaskUnauthorized).Payload.Message), ctx)
	case *tasks.GetTasksUnauthorized:
		return newUsageError(fmt.Sprintf("%v", err.(*tasks.GetTasksUnauthorized).Payload.Message), ctx)
	case *tasks.AddTaskInternalServerError:
		return newUsageError(fmt.Sprintf("%v", err.(*tasks.AddTaskInternalServerError).Payload.ErrorMessage), ctx)
	case *tasks.AddTaskUnauthorized:
		return newUsageError(fmt.Sprintf("%v", err.(*tasks.AddTaskUnauthorized).Payload.Message), ctx)
	case *tasks.UpdateTaskStateBadRequest:
		return newUsageError(fmt.Sprintf("%v", err.(*tasks.UpdateTaskStateBadRequest).Payload.ErrorMessage), ctx)
	case *tasks.UpdateTaskStateConflict:
		return newUsageError(fmt.Sprintf("%v", err.(*tasks.UpdateTaskStateConflict).Payload.ErrorMessage), ctx)
	case *tasks.UpdateTaskStateInternalServerError:
		return newUsageError(fmt.Sprintf("%v", err.(*tasks.UpdateTaskStateInternalServerError).Payload.ErrorMessage), ctx)
	case *tasks.UpdateTaskStateUnauthorized:
		return newUsageError(fmt.Sprintf("%v", err.(*tasks.UpdateTaskStateUnauthorized).Payload.Message), ctx)
	default:
		// this is a hack
		if strings.Contains(err.Error(), "tls: oversized record") || strings.Contains(err.Error(), "malformed HTTP response") {
			return newUsageError(extractError(err.Error()), ctx)
		}
		return newUsageError(fmt.Sprintf("Error: %v", err), ctx)
	}
}

type config struct {
	RestAPI restAPIConfig `json:"rest"`
}
type restAPIConfig struct {
	Password *string `json:"rest-auth-pwd"`
}

// checkForAuth Checks for authentication flags and returns a username/password
// from the specified settings
func checkForAuth(ctx *cli.Context) (username, password string) {
	if ctx.Bool("password") {
		username = "snap" // for now since username is unused but needs to exist for basicAuth
		// Prompt for password
		fmt.Print("Password:")
		pass, err := terminal.ReadPassword(0)
		if err != nil {
			password = ""
		} else {
			password = string(pass)
		}
		// Go to next line after password prompt
		fmt.Println()
	}

	if ctx.IsSet("config") {
		cfg := &config{}
		if err := cfg.loadConfig(ctx.String("config")); err != nil {
			fmt.Println(err)
		}
		if cfg.RestAPI.Password != nil {
			password = *cfg.RestAPI.Password
		} else {
			fmt.Println("Error config password field 'rest-auth-pwd' is empty")
		}
	}
	return username, password
}

func (c *config) loadConfig(path string) error {
	log.WithFields(log.Fields{
		"_module":     "snaptel-config",
		"_block":      "loadConfig",
		"config_path": path,
	}).Warning("The snaptel configuration file will be deprecated. Find more information here: https://github.com/intelsdi-x/snap/issues/1539")
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("Unable to read config. File might not exist")
	}
	err = json.Unmarshal(b, &c)
	if err != nil {
		return fmt.Errorf("Invalid config")
	}
	return nil
}

// BasicAuth returns the instance of runtime.ClientAuthInfoWriter.
func BasicAuth(ctx *cli.Context) runtime.ClientAuthInfoWriter {
	if ctx.IsSet("password") || ctx.IsSet("config") {
		u, p := checkForAuth(ctx)
		password = p
		return openapiclient.BasicAuth(u, p)
	}
	return nil
}

// extractError is a hack for SSL/TLS handshake error.
func extractError(m string) string {
	ts := strings.Split(m, "\"")

	var tss []string
	if len(ts) > 0 {
		tss = strings.Split(ts[0], "malformed")
	}

	errMsg := "Error connecting to API. Do you have an http/https mismatching API request?"
	if len(tss) > 0 {
		errMsg = tss[0] + errMsg
	}
	return errMsg
}
