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

package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"

	openapiclient "github.com/go-openapi/runtime/client"
	"github.com/golang/glog"
	"github.com/intelsdi-x/snap-cli/snaptel"
	"github.com/intelsdi-x/snap-client-go/client"
	"github.com/urfave/cli"
)

var (
	gitversion string
)

type tlsClientOptions struct {
	insecureSkipVerify bool
}

func main() {
	app := cli.NewApp()
	app.Name = "snaptel"
	app.Version = gitversion
	app.Usage = "The open telemetry framework"
	app.Flags = []cli.Flag{snaptel.FlURL, snaptel.FlSecure, snaptel.FlAPIVer, snaptel.FlPassword, snaptel.FlConfig, snaptel.FlTimeout}
	app.Commands = snaptel.Commands
	sort.Sort(ByCommand(app.Commands))
	app.Before = beforeAction

	app.Setup()
	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		if ue, ok := err.(snaptel.UsageError); ok {
			ue.Help()
		}
		os.Exit(1)
	}
}

// Run before every command
func beforeAction(ctx *cli.Context) error {
	snaptel.FlURL.Value = ctx.String("url")
	snaptel.FlAPIVer.Value = ctx.String("api-version")

	u, err := url.Parse(snaptel.FlURL.Value)
	if err != nil {
		glog.Fatal(err)
	}

	tlcOpts := tlsClientOptions{insecureSkipVerify: ctx.Bool("insecure")}
	tlcClient := tlsClient(tlcOpts)
	rt := openapiclient.NewWithClient(u.Host, snaptel.FlAPIVer.Value, []string{u.Scheme}, tlcClient)
	c := client.New(rt, nil)
	snaptel.SetClient(c)
	snaptel.SetScheme(u.Scheme)
	snaptel.SetAuthInfo(snaptel.BasicAuth(ctx))

	return nil
}

// tlsClient creates a http.Client
func tlsClient(opts tlsClientOptions) *http.Client {
	transport := tlsTransport(opts)
	return &http.Client{Transport: transport}
}

func tlsTransport(opts tlsClientOptions) http.RoundTripper {
	cfg := &tls.Config{}
	cfg.InsecureSkipVerify = opts.insecureSkipVerify
	cfg.BuildNameToCertificate()
	return &http.Transport{TLSClientConfig: cfg}
}

// ByCommand contains array of CLI commands.
type ByCommand []cli.Command

func (s ByCommand) Len() int {
	return len(s)
}
func (s ByCommand) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByCommand) Less(i, j int) bool {
	if s[i].Name == "help" {
		return false
	}
	if s[j].Name == "help" {
		return true
	}
	return s[i].Name < s[j].Name
}
