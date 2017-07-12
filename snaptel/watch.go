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
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"text/tabwriter"

	"github.com/intelsdi-x/snap-client-go/models"
	"github.com/urfave/cli"
)

type watchErrorResponse struct {
	Message string
}

func watchTask(ctx *cli.Context) error {
	if len(ctx.Args()) != 1 {
		return newUsageError("Incorrect usage", ctx)
	}

	verbose := ctx.Bool("verbose")
	id := ctx.Args().First()
	url := fmt.Sprintf("%s/%s/tasks/%s/watch", FlURL.Value, FlAPIVer.Value, id)

	// Currently, there is no way to implement a proper idel timeout for streaming.
	// Therefore no timeout for this request.
	req, err := http.NewRequest("GET", url, nil)
	req.SetBasicAuth("snap", password)
	cli := &http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var lines int
	// Catch interrupt signal so we can return to command line without formatting issues
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Printf("%sStopping task watch\n", strings.Repeat("\n", lines))
		resp.Body.Close()
		os.Exit(0)
	}()

	// Decode and display error message in case of error response

	if resp.StatusCode == 500 {
		errRespBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("An error occured while reading task watch error response: %s", err)
		}
		errResp := watchErrorResponse{}
		err = json.Unmarshal(errRespBody, &errResp)
		if err != nil {
			return fmt.Errorf("An error occured while unmarshalling task watch error response JSON data: %s", err)
		}
		return fmt.Errorf(errResp.Message)
	}

	var tskEvent models.StreamedTaskEvent
	reader := bufio.NewReader(resp.Body)

	fmt.Printf("Watching Task (%s):\n", id)

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	fields := []interface{}{"NAMESPACE", "DATA", "TIMESTAMP"}
	if verbose {
		fields = append(fields, "TAGS")
	}

	for {
		bs, err := reader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return err
		}

		if len(bs) < 2 {
			continue
		}

		bsData := bytes.TrimPrefix(bytes.TrimSpace(bs), []byte("data: "))
		if len(bsData) == 0 {
			continue
		}

		err = json.Unmarshal(bsData, &tskEvent)
		if err != nil {
			return fmt.Errorf("Error unmarshal task stream: %v", err)
		}

		var extra int

		// Print header fields if data received
		if len(tskEvent.Event) > 0 {
			printFields(w, false, 0, fields...)
			extra++
		}

		for _, e := range tskEvent.Event {
			fmt.Printf("\033[0J")
			eventFields := []interface{}{
				e.Namespace,
				e.Data,
				e.Timestamp,
			}
			if !verbose {
				printFields(w, false, 0, eventFields...)
				continue
			}
			tags := sortTags(e.Tags)
			if len(tags) <= 3 {
				eventFields = append(eventFields, strings.Join(tags, ", "))
				printFields(w, false, 0, eventFields...)
				continue
			}
			for i := 0; i < len(tags); i += 3 {
				tagSlice := tags[i:min(i+3, len(tags))]
				if i == 0 {
					eventFields = append(eventFields, strings.Join(tagSlice, ", ")+",")
					printFields(w, false, 0, eventFields...)
					continue
				}
				extra++
				if i+3 > len(tags) {
					printFields(w, false, 0,
						"",
						"",
						"",
						strings.Join(tagSlice, ", "),
					)
					continue
				}
				printFields(w, false, 0,
					"",
					"",
					"",
					strings.Join(tagSlice, ", ")+",",
				)
			}
		}
		lines = len(tskEvent.Event) + extra
		fmt.Fprintf(w, "\033[%dA\n", lines+1)
		w.Flush()

		if err == io.EOF {
			break
		}
	}
	return nil
}
