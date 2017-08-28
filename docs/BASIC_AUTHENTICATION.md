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

## Basic Authentication

**Table of contents:**
  * [Overview](#overview)
  * [Examples](#examples)

### Overview
Basic authentication is an optional authentication handler for Snap CLI. Snap daemon running with set flag `--rest-auth` requires passing a password
for REST API authentication on every request, as well as giving your credentials to clients.


### Examples
In one terminal window, run snapteld (log level is set to 1, signing is turned off, specify --rest-auth flag).
You will be prompted for the password that you will use for authentication.
```
$ snapteld -l 1 -t 0 --rest-auth
...
What password do you want to use for authentication?
Password: <your_password>
```

In another terminal, use snap CLI `snaptel` with flag `--password` (or `-p`), for example, to list loaded plugins.
You will be prompted for providing the password to authenticate.
```
$ snaptel -p plugin list

Password: <your_password>
```

If Snap's client authentication is successful (provided password is correct), requested command will be executed and its output will be returned:
```
$ snaptel -p plugin list

Password: <your_password>

NAME     VERSION         TYPE            SIGNED          STATUS          LOADED TIME
cpu      7               collector       false           loaded          Fri, 25 Aug 2017 13:49:30 CEST
```

If Snap's client authentication is failing (provided password is incorrect), the following error message will be returned:

```
$ snaptel -p plugin list

Password: <wrong_password>

Error: Invalid credentials
```
