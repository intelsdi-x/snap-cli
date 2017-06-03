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
	"net/http"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/intelsdi-x/snap-client-go/models"
	"github.com/urfave/cli"
)

func watchTask(ctx *cli.Context) error {
	if len(ctx.Args()) != 1 {
		return newUsageError("Incorrect usage", ctx)
	}

	verbose := ctx.Bool("verbose")
	id := ctx.Args().First()
	url := fmt.Sprintf("%s/%s/tasks/%s/watch", FlURL.Value, FlAPIVer.Value, id)

	// Currently, there is no way to implement a proper idel timeout for streaming.
	// Therefore no timeout for this request.
	resp, err := http.Get(url)
	defer resp.Body.Close()
	if err != nil {
		return err
	}

	var tskEvent models.StreamedTaskEvent
	reader := bufio.NewReader(resp.Body)

	delim := []byte{':', ' '}

	fmt.Printf("Watching Task (%s):\n", id)

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	fields := []interface{}{"NAMESPACE", "DATA", "TIMESTAMP"}
	if verbose {
		fields = append(fields, "TAGS")
	}
	printFields(w, false, 0, fields...)
	w.Flush()

	for {
		bs, err := reader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return err
		}

		if len(bs) < 2 {
			continue
		}

		spl := bytes.Split(bs, delim)
		if len(spl) < 2 {
			continue
		}

		err = json.Unmarshal(bytes.TrimSpace(spl[1]), &tskEvent)
		if err != nil {
			return fmt.Errorf("Error unmarshal task stream: %v", err)
		}

		var lines int
		var extra int
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

		if err == io.EOF {
			break
		}
	}
	w.Flush()
	return nil
}
