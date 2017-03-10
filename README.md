<!--
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015-2017 Intel Corporation

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
A powerful telemetry framework CLI

## Usage
Either copy `snaptel` to `/usr/local/sbin/snaptel` and ensure `/usr/local/sbin` is in your path, or use fully qualified filepath to the `snaptel` binary:

```
$ snaptel [global options] command [command options] [arguments...]
```

### Global Options
```
--url, -u 'http://localhost:8181'    Sets the URL to use [$SNAP_URL]
--api-version, -a 'v1'               The Snap API version [$SNAP_API_VERSION]
--config, -c                         Path to a config file [$SNAPTEL_CONFIG_PATH]
--help, -h                           show help
--version, -v                        print the version
```

### Commands
```
metric
plugin
help, h      Shows a list of commands or help for one command
```

### Command Options

#### plugin
```
$ snaptel plugin command [command options] [arguments...]
```
```
load        load <plugin_path>
unload      unload <plugin_type> <plugin_name> <plugin_version>
list        list
help, h     Shows a list of commands or help for one command
```

#### metric
```
$ snaptel metric command [command options] [arguments...]
```
```
list         list
get          get details on a single metric
help, h      Shows a list of commands or help for one command
```

Example Usage
-------------

### Load and unload plugins, create and start a task

In one terminal window, run snapteld (log level is set to 1 and signing is turned off for this example):
```
$ snapteld --log-level 1 --log-path '' --plugin-trust 0
```

and then in another terminal:

1. load a collector plugin
2. load a processing plugin
3. load a publishing plugin
4. list the plugins
5. list loaded metrics
6. list loaded metrics with details
7. list a specific metric including all versions
8. list a specific metric with its version
9. list a plugin config
10. unload the plugins

```
$ snaptel plugin load /opt/snap/plugins/snap-plugin-collector-mock1
$ snaptel plugin load /opt/snap/plugins/snap-plugin-processor-passthru
$ snaptel plugin load /opt/snap/plugins/snap-plugin-publisher-mock-file
$ snaptel plugin list
$ snaptel metric list
$ snaptel metric list --verbose
$ snaptel metric get -m /intel/mock/foo
$ snaptel metric get -m /intel/mock/foo -v <version>
$ snaptel plugin config get collector:mock:<version>
$ snaptel plugin unload processor passthru <version>
$ snaptel plugin unload publisher publisher <version>
```

