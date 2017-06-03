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
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	yaml "gopkg.in/yaml.v2"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/intelsdi-x/snap-client-go/client/tasks"
	"github.com/intelsdi-x/snap-client-go/models"
	"github.com/robfig/cron"
	"github.com/urfave/cli"
)

var (
	// padding to picking a time to start a "NOW" task
	createTaskNowPad = time.Second * 1
	timeParseFormat  = "3:04PM"
	dateParseFormat  = "1-02-2006"
	unionParseFormat = timeParseFormat + " " + dateParseFormat
)

// Constants used to truncate task hit and miss counts
// e.g. 1K(10^3), 1M(10^6, 1G(10^9) etc (not 1024^#). We do not
// use units larger than Gb to support 32 bit compiles.
const (
	K = 1000
	M = 1000 * K
	G = 1000 * M
)

func trunc(n int) string {
	var u string

	switch {
	case n >= G:
		u = "G"
		n /= G
	case n >= M:
		u = "M"
		n /= M
	case n >= K:
		u = "K"
		n /= K
	default:
		return strconv.Itoa(n)
	}
	return strconv.Itoa(n) + u
}

func createTask(ctx *cli.Context) error {
	var err error
	if ctx.IsSet("task-manifest") {
		err = createTaskUsingTaskManifest(ctx)
	} else if ctx.IsSet("workflow-manifest") {
		err = createTaskUsingWFManifest(ctx)
	} else {
		return newUsageError("Must provide either --task-manifest or --workflow-manifest arguments", ctx)
	}
	return err
}

func stringValToInt(val string) (int, error) {
	// parse the input (string) as an integer value (and return that integer value
	// to the caller or an error if the input value cannot be parsed as an integer)
	parsedField, err := strconv.Atoi(val)
	if err != nil {
		splitErr := strings.Split(err.Error(), ": ")
		errStr := splitErr[len(splitErr)-1]
		// return a value of zero and the error encountered during string parsing
		return 0, fmt.Errorf("Value '%v' cannot be parsed as an integer (%v)", val, errStr)
	}
	// return the integer equivalent of the input string and a nil error (indicating success)
	return parsedField, nil
}

// Parses the command-line parameters (if any) and uses them to override the underlying
// schedule for this task or to set a schedule for that task (if one is not already defined,
// as is the case when we're building a new task from a workflow manifest).
//
// Note: in this method, any of the following types of time windows can be specified:
//
//    +---------------------------...  (start with no stop and no duration; no end time for window)
//    |
//  start
//
//    ...---------------------------+  (stop with no start and no duration; no start time for window)
//                                  |
//                                 stop
//
//    +-----------------------------+  (start with a duration but no stop)
//    |                             |
//    |---------------------------->|
//  start        duration
//
//    +-----------------------------+  (stop with a duration but no start)
//    |                             |
//    |<----------------------------|
//               duration         stop
//
//    +-----------------------------+ (start and stop both specified)
//    |                             |
//    |<--------------------------->|
//  start                         stop
//
//    +-----------------------------+ (only duration specified, implies start is the current time)
//    |                             |
//    |---------------------------->|
//  Now()        duration
//
func setWindowedSchedule(start *time.Time, stop *time.Time, duration *time.Duration, t *models.Task) error {
	// if there is an empty schedule already defined for this task, then set the
	// type for that schedule to 'windowed'
	if *t.Schedule.Type == "" {
		*t.Schedule.Type = "windowed"
	} else if *t.Schedule.Type != "windowed" {
		// else if the task's existing schedule is not a 'windowed' schedule,
		// then return an error
		return fmt.Errorf("Usage error (schedule type mismatch); cannot replace existing schedule of type '%v' with a new, 'windowed' schedule", *t.Schedule.Type)
	}
	// if a duration was passed in, determine the start and stop times for our new
	// 'windowed' schedule from the input parameters
	if duration != nil {
		// if start and stop were both defined, then return an error (since specifying the
		// start, stop, *and* duration for a 'windowed' schedule is not allowed)
		if start != nil && stop != nil {
			return fmt.Errorf("Usage error (too many parameters); the window start, stop, and duration cannot all be specified for a 'windowed' schedule")
		}
		// if start is set and stop is not then use duration to create stop
		if start != nil && stop == nil {
			newStop := start.Add(*duration)
			t.Schedule.StartTimestamp = start
			t.Schedule.StopTimestamp = &newStop
			return nil
		}
		// if stop is set and start is not then use duration to create start
		if stop != nil && start == nil {
			newStart := stop.Add(*duration * -1)
			t.Schedule.StartTimestamp = &newStart
			t.Schedule.StopTimestamp = stop
			return nil
		}
		// otherwise, the start and stop are both undefined but a duration was passed in,
		// so use the current date/time (plus the 'createTaskNowPad' value) as the
		// start date/time and construct a stop date/time from that start date/time
		// and the duration
		newStart := time.Now().Add(createTaskNowPad)
		newStop := newStart.Add(*duration)
		t.Schedule.StartTimestamp = &newStart
		t.Schedule.StopTimestamp = &newStop
		return nil
	}
	// if a start date/time was specified, we will use it to replace
	// the current schedule's start date/time
	if start != nil {
		t.Schedule.StartTimestamp = start
	}
	// if a stop date/time was specified, we will use it to replace the
	// current schedule's stop date/time
	if stop != nil {
		t.Schedule.StopTimestamp = stop
	}
	// if we get this far, then just return a nil error (indicating success)
	return nil
}

// parse the command-line options and use them to setup a new schedule for this task
func setScheduleFromCliOptions(ctx *cli.Context, t *models.Task) error {
	// check the start, stop, and duration values to see if we're looking at a windowed schedule (or not)
	// first, get the parameters that define the windowed schedule
	start := mergeDateTime(
		strings.ToUpper(ctx.String("start-time")),
		strings.ToUpper(ctx.String("start-date")),
	)
	stop := mergeDateTime(
		strings.ToUpper(ctx.String("stop-time")),
		strings.ToUpper(ctx.String("stop-date")),
	)
	// Grab the duration string (if one was passed in) and parse it
	durationStr := ctx.String("duration")
	var duration *time.Duration
	if ctx.IsSet("duration") || durationStr != "" {
		d, err := time.ParseDuration(durationStr)
		if err != nil {
			return fmt.Errorf("Usage error (bad duration format); %v", err)
		}
		duration = &d
	}
	// Grab the interval for the schedule (if one was provided). Note that if an
	// interval value was not passed in and there is no interval defined for the
	// schedule associated with this task, it's an error
	interval := ctx.String("interval")
	if t.Schedule == nil && !ctx.IsSet("interval") && interval == "" && *(t.Schedule.Interval) == "" {
		return fmt.Errorf("Usage error (missing interval value); when constructing a new task schedule an interval must be provided")
	}
	// if a start, stop, or duration value was provided, or if the existing schedule for this task
	// is 'windowed', then it's a 'windowed' schedule
	isWindowed := (start != nil || stop != nil || duration != nil || (t.Schedule.Type != nil && *(t.Schedule.Type) == "windowed"))
	// if an interval was passed in, then attempt to parse it (first as a duration,
	// then as the definition of a cron job)
	isCron := false
	if interval != "" {
		// first try to parse it as a duration
		_, err := time.ParseDuration(interval)
		if err != nil {
			// if that didn't work, then try parsing the interval as cron job entry
			_, e := cron.Parse(interval)
			if e != nil {
				return fmt.Errorf("Usage error (bad interval value): cannot parse interval value '%v' either as a duration or a cron entry", interval)
			}
			// if it's a 'windowed' schedule, then return an error (we can't use a
			// cron entry interval with a 'windowed' schedule)
			if isWindowed {
				return fmt.Errorf("Usage error; cannot use a cron entry ('%v') as the interval for a 'windowed' schedule", interval)
			}
			isCron = true
		}
		t.Schedule.Interval = &interval
	}
	// if it's a 'windowed' schedule, then create a new 'windowed' schedule and add it to
	// the current task; the existing schedule (if on exists) will be replaced by the new
	// schedule in this method (note that it is an error to try to replace an existing
	// schedule with a new schedule of a different type, so an error will be returned from
	// this method call if that is the case)
	if isWindowed {
		return setWindowedSchedule(start, stop, duration, t)
	}
	// if it's not a 'windowed' schedule, then set the schedule type based on the 'isCron' flag,
	// which was set above.
	if isCron {
		// make sure the current schedule type (if there is one) matches; if not it is an error
		if t.Schedule.Type != nil && *(t.Schedule.Type) != "cron" {
			return fmt.Errorf("Usage error; cannot replace existing schedule of type '%v' with a new, 'cron' schedule", t.Schedule.Type)
		}
		*t.Schedule.Type = "cron"
		return nil
	}
	// if it wasn't a 'windowed' schedule and it's not a 'cron' schedule, then it must be a 'simple'
	// schedule, so first make sure the current schedule type (if there is one) matches; if not
	// then it's an error
	if t.Schedule.Type != nil && *(t.Schedule.Type) != "simple" {
		return fmt.Errorf("Usage error; cannot replace existing schedule of type '%v' with a new, 'simple' schedule", t.Schedule.Type)
	}

	countValStr := ctx.String("count")
	if ctx.IsSet("count") || countValStr != "" {
		count, err := stringValToUint64(countValStr)
		if err != nil {
			return fmt.Errorf("Usage error (bad count format); %v", err)
		}
		t.Schedule.Count = count

	}

	// if we get this far set the schedule type and return a nil error (indicating success)
	ty := "simple"
	t.Schedule.Type = &ty

	return nil
}

// stringValToUint64 parses the input (string) as an unsigned integer value (and returns that uint value
// to the caller or an error if the input value cannot be parsed as an unsigned integer)
func stringValToUint64(val string) (uint64, error) {
	parsedField, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		splitErr := strings.Split(err.Error(), ": ")
		errStr := splitErr[len(splitErr)-1]
		// return a value of zero and the error encountered during string parsing
		return 0, fmt.Errorf("Value '%v' cannot be parsed as an unsigned integer (%v)", val, errStr)
	}
	// return the unsigned integer equivalent of the input string and a nil error (indicating success)
	return uint64(parsedField), nil
}

// merge the command-line options into the current task
func mergeCliOptions(ctx *cli.Context, t *models.Task) error {
	st := !ctx.IsSet("no-start")
	t.Start = st

	// set the name of the task (if a 'name' was provided in the CLI options)
	name := ctx.String("name")
	if ctx.IsSet("name") || name != "" {
		t.Name = name
	}

	// set the deadline of the task (if a 'deadline' was provided in the CLI options)
	deadline := ctx.String("deadline")
	if ctx.IsSet("deadline") || deadline != "" {
		t.Deadline = deadline
	}
	// set the MaxFailures for the task (if a 'max-failures' value was provided in the CLI options)
	maxFailuresStrVal := ctx.String("max-failures")
	if ctx.IsSet("max-failures") || maxFailuresStrVal != "" {
		maxFailures, err := stringValToInt(maxFailuresStrVal)
		if err != nil {
			return err
		}
		t.MaxFailures = int64(maxFailures)
	}
	// set the schedule for the task from the CLI options (and return the results
	// of that method call, indicating whether or not an error was encountered while
	// setting up that schedule)
	return setScheduleFromCliOptions(ctx, t)
}

func convert(i interface{}) interface{} {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2i := map[string]interface{}{}
		for k, v := range x {
			m2i[k.(string)] = convert(v)
		}
		return m2i
	case []interface{}:
		for idx, v := range x {
			x[idx] = convert(v)
		}
	}
	return i
}

func createTaskUsingTaskManifest(ctx *cli.Context) error {
	// get the task manifest file to use
	path := ctx.String("task-manifest")
	ext := filepath.Ext(path)
	file, e := ioutil.ReadFile(path)
	if e != nil {
		return fmt.Errorf("File error [%s] - %v", ext, e)
	}

	bts := []byte(os.ExpandEnv(string(file)))

	var tsk *models.Task
	switch ext {
	case ".yaml", ".yml":
		t, err := taskYamlToJSON(ctx, bts)
		if err != nil {
			return err
		}
		tsk = t
	case ".json":
		t, err := taskJSONToJSON(ctx, bts)
		if err != nil {
			return err
		}
		tsk = t
	default:
		return fmt.Errorf("Unsupported file type %s", ext)
	}

	// Request parameters
	params := tasks.NewAddTaskParamsWithTimeout(FlTimeout.Value)
	params.SetTask(tsk)

	resp, err := client.Tasks.AddTask(params)
	if err != nil {
		return getErrorDetail(err, ctx)
	}

	res := resp.Payload
	fmt.Println("Task created")
	fmt.Printf("ID: %s\n", res.ID)
	fmt.Printf("Name: %s\n", res.Name)
	fmt.Printf("State: %s\n", res.TaskState)

	return nil
}

func createTaskUsingWFManifest(ctx *cli.Context) error {
	// Get the workflow manifest filename from the command-line
	path := ctx.String("workflow-manifest")
	ext := filepath.Ext(path)
	file, e := ioutil.ReadFile(path)
	if e != nil {
		return fmt.Errorf("File error [%s] - %v", ext, e)
	}

	// check to make sure that an interval was specified using the appropriate command-line flag
	interval := ctx.String("interval")
	if !ctx.IsSet("interval") || interval == "" {
		return fmt.Errorf("Workflow manifest requires that an interval be set via a command-line flag")
	}

	var tsk *models.Task
	switch ext {
	case ".yaml", ".yml":
		t, err := wfYamlToJSON(ctx, file)
		if err != nil {
			return err
		}
		tsk = t
	case ".json":
		t, err := wfJSONtoJSON(ctx, file)
		if err != nil {
			return err
		}
		tsk = t
	default:
		return fmt.Errorf("Unsupported file type %s", ext)
	}

	params := tasks.NewAddTaskParamsWithTimeout(FlTimeout.Value)
	params.SetTask(tsk)

	resp, err := client.Tasks.AddTask(params)
	if err != nil {
		return getErrorDetail(err, ctx)
	}
	res := resp.Payload
	fmt.Println("Task created")
	fmt.Printf("ID: %s\n", res.ID)
	fmt.Printf("Name: %s\n", res.Name)
	fmt.Printf("State: %s\n", res.TaskState)

	return nil
}

func mergeDateTime(tm, dt string) *time.Time {
	reTm := time.Now().Add(createTaskNowPad)
	if dt == "" && tm == "" {
		return nil
	}
	if dt != "" {
		t, err := time.Parse(dateParseFormat, dt)
		if err != nil {
			fmt.Printf("Error creating task:\n%v\n", err)
			os.Exit(1)
		}
		reTm = t
	}

	if tm != "" {
		_, err := time.ParseInLocation(timeParseFormat, tm, time.Local)
		if err != nil {
			fmt.Printf("Error creating task:\n%v\n", err)
			os.Exit(1)
		}
		reTm, err = time.ParseInLocation(unionParseFormat, fmt.Sprintf("%s %s", tm, reTm.Format(dateParseFormat)), time.Local)
		if err != nil {
			fmt.Printf("Error creating task:\n%v\n", err)
			os.Exit(1)
		}
	}
	return &reTm
}

func listTask(ctx *cli.Context) error {
	params := tasks.NewGetTasksParamsWithTimeout(FlTimeout.Value)
	resp, err := client.Tasks.GetTasks(params)
	if err != nil {
		return getErrorDetail(err, ctx)
	}

	termWidth, _, _ := terminal.GetSize(int(os.Stdout.Fd()))
	verbose := ctx.Bool("verbose")

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	if len(resp.Payload.Tasks) == 0 {
		fmt.Println("No task found. Have you created a task?")
		return nil
	}
	printFields(w, false, 0,
		"ID",
		"NAME",
		"STATE",
		"HIT",
		"MISS",
		"FAIL",
		"CREATED",
		"LAST FAILURE",
	)
	for _, task := range resp.Payload.Tasks {
		//165 is the width of the error message from ID - LAST FAILURE inclusive.
		//If the header row wraps, then the error message will automatically wrap too
		if termWidth < 165 {
			verbose = true
		}
		printFields(w, false, 0,
			task.ID,
			fixSize(verbose, task.Name, 41),
			task.TaskState,
			trunc(int(task.HitCount)),
			trunc(int(task.MissCount)),
			trunc(int(task.FailedCount)),
			time.Unix(task.CreationTimestamp, 0).Format(time.RFC1123),
			/*153 is the width of the error message from ID up to LAST FAILURE*/
			fixSize(verbose, task.LastFailureMessage, termWidth-153),
		)
	}
	w.Flush()

	return nil
}

func fixSize(verbose bool, msg string, width int) string {
	if len(msg) < width {
		for i := len(msg); i < width; i++ {
			msg += " "
		}
	} else if len(msg) > width && !verbose {
		return msg[:width-3] + "..."
	}
	return msg
}

func startTask(ctx *cli.Context) error {
	if len(ctx.Args()) != 1 {
		return newUsageError("Incorrect usage", ctx)
	}

	id := ctx.Args().First()

	params := tasks.NewUpdateTaskStateParamsWithTimeout(FlTimeout.Value)
	params.SetID(id)
	params.SetAction("start")

	_, err := client.Tasks.UpdateTaskState(params)
	if err != nil {
		return getErrorDetail(err, ctx)
	}

	fmt.Println("Task started:")
	fmt.Printf("ID: %s\n", id)

	return nil
}

func stopTask(ctx *cli.Context) error {
	if len(ctx.Args()) != 1 {
		return newUsageError("Incorrect usage", ctx)
	}

	id := ctx.Args().First()

	params := tasks.NewUpdateTaskStateParamsWithTimeout(FlTimeout.Value)
	params.SetID(id)
	params.SetAction("stop")

	_, err := client.Tasks.UpdateTaskState(params)
	if err != nil {
		return getErrorDetail(err, ctx)
	}

	fmt.Println("Task stopped:")
	fmt.Printf("ID: %s\n", id)

	return nil
}

func removeTask(ctx *cli.Context) error {
	if len(ctx.Args()) != 1 {
		return newUsageError("Incorrect usage", ctx)
	}

	id := ctx.Args().First()

	params := tasks.NewRemoveTaskParamsWithTimeout(FlTimeout.Value)
	params.SetID(id)

	_, err := client.Tasks.RemoveTask(params)
	if err != nil {
		return getErrorDetail(err, ctx)
	}

	fmt.Println("Task removed:")
	fmt.Printf("ID: %s\n", id)

	return nil
}

func enableTask(ctx *cli.Context) error {
	if len(ctx.Args()) != 1 {
		return newUsageError("Incorrect usage", ctx)
	}

	id := ctx.Args().First()

	params := tasks.NewUpdateTaskStateParamsWithTimeout(FlTimeout.Value)
	params.SetID(id)
	params.SetAction("enable")

	_, err := client.Tasks.UpdateTaskState(params)
	if err != nil {
		return getErrorDetail(err, ctx)
	}

	fmt.Println("Task enabled:")
	fmt.Printf("ID: %s\n", id)
	return nil
}

func exportTask(ctx *cli.Context) error {
	if len(ctx.Args()) != 1 {
		return newUsageError("Incorrect usage", ctx)
	}
	id := ctx.Args().First()
	params := tasks.NewGetTaskParamsWithTimeout(FlTimeout.Value)
	params.SetID(id)

	resp, err := client.Tasks.GetTask(params)
	if err != nil {
		return getErrorDetail(err, ctx)
	}

	tb, err := json.Marshal(resp.Payload)
	if err != nil {
		return fmt.Errorf("Error exporting task:\n%v", err)
	}
	fmt.Println(string(tb))
	return nil
}

func sortTags(tags map[string]string) []string {
	var tagSlice []string
	var keys []string
	for k := range tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		tagSlice = append(tagSlice, k+"="+tags[k])
	}
	return tagSlice
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func validateTask(t *models.Task) error {
	if err := validateScheduleExists(t.Schedule); err != nil {
		return err
	}
	if t.Version != 1 {
		return fmt.Errorf("Error: Invalid version provided for task manifest")
	}
	return nil
}

func validateScheduleExists(schedule *models.Schedule) error {
	if schedule == nil {
		return fmt.Errorf("Error: Task manifest did not include a schedule")
	}
	if *schedule == (models.Schedule{}) {
		return fmt.Errorf("Error: Task manifest included an empty schedule. Task manifests need to include a schedule")
	}
	return nil
}

func wfYamlToJSON(ctx *cli.Context, bts []byte) (*models.Task, error) {
	b, err := yamlToJSON(bts)
	if err != nil {
		return nil, err
	}

	wf := models.WorkflowMap{}
	err = json.Unmarshal(b, &wf)
	if err != nil {
		return nil, fmt.Errorf("Error parsing Workflow JSON: %v", err)
	}

	t := models.Task{
		Version:  1,
		Schedule: &models.Schedule{},
	}
	t.Workflow = &wf
	return toTaskJSON(ctx, t)
}

func taskYamlToJSON(ctx *cli.Context, bts []byte) (*models.Task, error) {
	b, err := yamlToJSON(bts)
	if err != nil {
		return nil, err
	}

	t := models.Task{}
	err = json.Unmarshal(b, &t)
	if err != nil {
		return nil, fmt.Errorf("Error parsing JSON file: %v", err)
	}
	return toTaskJSON(ctx, t)
}

func yamlToJSON(bts []byte) ([]byte, error) {
	var body interface{}
	if err := yaml.Unmarshal(bts, &body); err != nil {
		return nil, fmt.Errorf("Unmarshal YMAL file error: %v", err)
	}

	body = convert(body)
	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("Marshal interface to bytes error: %v", err)
	}
	return b, nil
}

func toTaskJSON(ctx *cli.Context, t models.Task) (*models.Task, error) {
	// merge any CLI options specified by the user (if any) into the current task;
	// if an error is encountered, return it
	if err := mergeCliOptions(ctx, &t); err != nil {
		return nil, err
	}

	// Validate task manifest includes schedule, workflow, and version
	if err := validateTask(&t); err != nil {
		return nil, err
	}
	return &t, nil
}

func taskJSONToJSON(ctx *cli.Context, bts []byte) (*models.Task, error) {
	t := models.Task{}

	if err := json.Unmarshal(bts, &t); err != nil {
		return nil, fmt.Errorf("Error parsing JSON file: %v", err)
	}
	return toTaskJSON(ctx, t)
}

func wfJSONtoJSON(ctx *cli.Context, bts []byte) (*models.Task, error) {
	wf := models.WorkflowMap{}
	err := json.Unmarshal(bts, &wf)
	if err != nil {
		return nil, fmt.Errorf("Error parsing Workflow JSON file: %v", err)
	}

	t := models.Task{
		Version:  1,
		Schedule: &models.Schedule{},
	}
	t.Workflow = &wf
	return toTaskJSON(ctx, t)
}
