<!--
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
-->

# snaptel
A Snap telemetry framework CLI

## Usage
Either copy `snaptel` to `/usr/local/sbin` and ensure `/usr/local/sbin` is in your path, or use fully qualified filepath to the `snaptel` binary:

```sh
$ snaptel 
```

### Global Options

```sh
NAME:
   snaptel - The open telemetry framework

USAGE:
   snaptel [global options] command [command options] [arguments...]

VERSION:
   all-clis-f9fa285

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --url value, -u value          Sets the URL to use (default: "http://localhost:8181") [$SNAP_URL]
   --insecure                     Ignore certificate errors when Snap API is running HTTPS [$SNAP_INSECURE]
   --api-version value, -a value  The Snap API version (default: "v2") [$SNAP_API_VERSION]
   --password, -p                 Require password for REST API authentication [$SNAP_REST_PASSWORD]
   --config value, -c value       Path to a config file [$SNAPTEL_CONFIG_PATH, $SNAPCTL_CONFIG_PATH]
   --timeout value, -t value      Timeout to be set on HTTP request to the server (default: 10s)
   --help, -h                     show help
   --version, -v                  print the version
```

### Commands
```
metric
plugin
task
help, h      Shows a list of commands or help for one command
```

### Command Options

#### plugin

```sh
$ snaptel plugin

NAME:
    -

USAGE:
    command [command options] [arguments...]

COMMANDS:
     load    load <plugin_path>
     unload  unload <plugin_type> <plugin_name> <plugin_version>
     list    list or list --running
     <unload_plugin_type>:<unload_plugin_name>:<unload_plugin_version> or swap <load_plugin_path> -t <unload_plugin_type> -n <unload_plugin_name> -v <unload_plugin_version>
     config

OPTIONS:
   --help, -h  show help

```

#### metric

```sh
$ snaptel metric

NAME:
    -

USAGE:
    command [command options] [arguments...]

COMMANDS:
     list  list
     get   get -m -v

OPTIONS:
   --help, -h  show help

```

#### task

```sh
$ snaptel task

NAME:
    -

USAGE:
    command [command options] [arguments...]

COMMANDS:
     create  There are two ways to create a task.
             1) Use a task manifest with [--task-manifest]
             2) Provide a workflow manifest and schedule details.

  * Note: Start and stop date/time are optional.

     list    list or list --verbose
     start   start <task_id>
     stop    stop <task_id>
     remove  remove <task_id>
     export  export <task_id>
     watch   watch <task_id> or watch <task_id> --verbose
     enable  enable <task_id>

OPTIONS:
   --help, -h  show help

```

```sh
$ snaptel task create -h

USAGE:
    create [command options] [arguments...]

DESCRIPTION:
   Creates a new task in the snap scheduler

OPTIONS:
   --task-manifest value, -t value      File path for task manifest to use for task creation.
   --workflow-manifest value, -w value  File path for workflow manifest to use for task creation
   --interval value, -i value           Interval for the task schedule [ex (simple schedule): 250ms, 1s, 30m (cron schedule): "0 * * * * *"]
   --count value                        The count of runs for the task schedule [defaults to 0 what means no limit, e.g. set to 1 determines a single run task]
   --start-date value                   Start date for the task schedule [defaults to today]
   --start-time value                   Start time for the task schedule [defaults to now]
   --stop-date value                    Stop date for the task schedule [defaults to today]
   --stop-time value                    Start time for the task schedule [defaults to now]
   --name value, -n value               Optional requirement for giving task names
   --duration value, -d value           The amount of time to run the task [appends to start or creates a start time before a stop]
   --no-start                           Do not start task on creation [normally started on creation]
   --deadline value                     The deadline for the task to be killed after started if the task runs too long (All tasks default to 5s)
   --max-failures value                 The number of consecutive failures before Snap disables the task

```

Example Usage
-------------

### Load and unload plugins, create and start a task

In one terminal window, run snapteld (log level is set to 1 and signing is turned off for this example):
```
$ snapteld --log-level 1 --log-path '' --plugin-trust 0
```

prepare a task manifest file, for example, task.json with following content:

```json
{
    "version": 1,
    "name": "sample",
    "schedule": {
        "type": "simple",
        "interval": "15s"
    },
    "workflow": {
        "collect": {
            "metrics": {
                "/intel/mock/foo": {},
                "/intel/mock/bar": {},
                "/intel/mock/*/baz": {}
            },
            "config": {
                "/intel/mock": {
                    "user": "root",
                    "password": "secret"
                }
            },
            "process": null,
            "publish": [
                {
                    "plugin_name": "file",                            
                    "config": {
                        "file": "/tmp/collected_swagger"
                    }
                }
            ]             
        }
    }
}
```

prepare a workflow manifest file, for example, workflow.json with the following content:
```json
{
    "collect": {
        "metrics": {
            "/intel/mock/foo": {}
        },
        "config": {
            "/intel/mock/foo": {
                "password": "testval"
            }
        },
        "process": [],
        "publish": [
            {
                "plugin_name": "file",
                "config": {
                    "file": "/tmp/rest.test"
                }
            }
        ]
    }
}
```

and then in another terminal:

1. load a collector plugin
2. load a processing plugin
3. load a publishing plugin
4. list the plugins
5. list running plugins
6. swap plugins
7. list loaded metrics
8. list loaded metrics with details
9. list a specific metric including all versions
10. list a specific metric with its version
11. list a plugin config
12. unload the plugins
13. create a task using task manifest
14. create a task using workflow
15. create a single run task
16. list tasks
17. watch a task
18. export a task
19. stop a task

```sh
$ snaptel plugin load /opt/snap/plugins/snap-plugin-collector-mock1
$ snaptel plugin load /opt/snap/plugins/snap-plugin-processor-passthru
$ snaptel plugin load /opt/snap/plugins/snap-plugin-publisher-mock-file
$ snaptel plugin list
$ snaptel plugin list --running
$ snaptel metric list
$ snaptel metric list --verbose
$ snaptel metric get -m /intel/mock/foo
$ snaptel metric get -m /intel/mock/foo -v <version>
$ snaptel plugin config get <plugin_type>:<plugin_name>:<plugin_version>
$ snaptel plugin config get -t <plugin_type> -n <plugin_name> -v <plugin_version>
$ snaptel plugin unload collector <plugin_name> <plugin_version>
$ snaptel task create -t mock-file.json
$ snaptel task create -w workflow.json -i 1s
$ snaptel task create -t mock-file.yml --count 1
$ snaptel task list
$ snaptel task watch <task_id>
$ snaptel task export <task_id>
$ snaptel task stop <task_id>
```

### Basic Authentication

In one terminal window, run snapteld (log level is set to 1, signing is turned off, specify --rest-auth flag):
```
$ snapteld -l 1 -t 0 --rest-auth
...
What password do you want to use for authentication?
Password:snap
```

snaptel will have to use the _`same password`_ that used to start snapteld. In another terminal:
1. list plugins
2. list metrics
3. load a plugin
4. create a task

```
$ snaptel -p plugin list
$ snaptel -p metric list
$ snaptel -p plugin load /opt/snap/plugins/snap-plugin-collector-mock1
$ snaptel -p task create -t mock-file.yml
```

### Secure GRPC plugins
Snap supports TLS for GRPC plugins. Referring to [secure plugin communication](https://github.com/intelsdi-x/snap/blob/master/docs/SECURE_PLUGIN_COMMUNICATION.md) for details. How to setup TLS on both server and client? The [Setup TLS Certificates](https://github.com/intelsdi-x/snap/blob/master/docs/SETUP_TLS_CERTIFICATES.md) has everything.

#### Examples

##### Definition of flags

| Flag | Description |
| ------ | ------ |
| plugin-cert | TLS server certificate |
| plugin-key | TLS server private key |
| plugin-ca-certs  | TLS server CA certificates |

##### Starting `snapteld` 

Snap is a client for all GRPC plugins. Note that Snap loads CA certificates from your OS certificate trust store if it's not specified.

```sh
$snapteld  -t 0 -l 1  --tls-cert snaptest-cli.crt --tls-key snaptest-cli.key --ca-cert-paths snaptest-ca.crt
```

##### Running `snaptel`

```sh
▶ snaptel plugin load --plugin-cert=snaptest-srv.crt --plugin-key=snaptest-srv.key --plugin-ca-certs=snaptest-ca.crt ../snap-plugin-lib-go/rand-collector
Plugin loaded
Name: test-rand-collector
Version: 1
Type: collector
Signed: false
Loaded Time: Mon, 14 Aug 2017 22:25:16 PDT
```

##### Error One

```sh
▶ snaptel plugin load --plugin-cert=snaptest-srv.crt  --plugin-ca-certs=snaptest-ca.crt ../snap-plugin-lib-go/rand-collector
Error: Both plugin certification and key are mandatory.
Usage: load <plugin_path> [--plugin-cert=<plugin_cert_path> --plugin-key=<plugin_key_path> --plugin-ca-certs=<ca_cert_paths>]
```

> What happened: Both `plugin-cert` and `plugin-key` are mandatory.


##### Error Two

```sh
▶ snaptel plugin load --plugin-cert=snaptest-srv.crt --plugin-key=snaptest-srv.key --plugin-ca-certs=snaptest-ca.crt ../snap-plugin-lib-go/rand-collector
Error: rpc error: code = Internal desc = connection error: desc = "transport: authentication handshake failed: x509: certificate signed by unknown authority"
Usage: load <plugin_path> [--plugin-cert=<plugin_cert_path> --plugin-key=<plugin_key_path> --plugin-ca-certs=<ca_cert_paths>]

```

> What happened: Did you start `snapteld` with CA cert or put the trusted CA in your OS/APP trust store?

##### Error Three

```sh
▶ snaptel plugin load --plugin-cert snaptest-srv.crt --plugin-key snaptest-srv.key --plugin-ca-certs snaptest-ca.crt ../snap/snap-plugin-collector-mock1
Error: secure framework can't connect to insecure plugin; plugin_name: mock
Usage: load <plugin_path> [--plugin-cert=<plugin_cert_path> --plugin-key=<plugin_key_path> --plugin-ca-certs=<ca_cert_paths>]
```

>What happened: The TLS is only supported for GRPC plugins. Restarting `snapteld` without TLS to load non-GRPC plugins.

