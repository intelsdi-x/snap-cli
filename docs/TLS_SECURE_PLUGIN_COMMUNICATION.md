## Secure Plugin Communication

**Table of contents:**
  * [Overview](#overview)
  * [How to setup TLS plugin communication](#how-to-setup-tls-plugin-communication)
  * [Common mistakes](#common-mistakes)

### Overview
Snap framework communicates with plugins (collectors, processors and publishers) over gRPC protocol. This communication can be secured
by opening TLS channels and providing certificates to authenticate both sides: plugins and Snap daemon.

Snap CLI exposes the following flags to allow loading the plugin within passed paths to its key and certificate files.

```
$ snaptel plugin load --help

USAGE:
   snaptel plugin load [command options] [arguments...]

OPTIONS:
   --plugin-asc value, -a value       The plugin asc
   --plugin-cert value, -c value      The path to plugin certificate file
   --plugin-key value, -k value       The path to plugin private key file
   --plugin-ca-certs value, -r value  List of CA cert paths (directory/file) for plugin to verify TLS clients

```

**Notice** Setup TLS communication applies only to plugins which use gRPC protocol.

To find more information, read
[Secure Plugin Communication](https://github.com/intelsdi-x/snap/blob/master/docs/SECURE_PLUGIN_COMMUNICATION.md) for details.


### How to setup TLS plugin communication

#### 0. Pre-work: generate TLS certificates
The instruction how to generate TLS certificates is available here: [Setup TLS Certificates](https://github.com/intelsdi-x/snap/blob/master/docs/SETUP_TLS_CERTIFICATES.md)

#### 1. Start Snap daemon
Snap daemon is a client for all GRPC plugins. Start `snapteld` with flags `--tls-cert`, `--tls-key` at least.

Using `--ca-cert-paths` is optional, and if not specified, Snap loads CA certificates from your OS certificate trust store.

```sh
$ snapteld  -t 0 -l 1  --tls-cert snaptest-cli.crt --tls-key snaptest-cli.key --ca-cert-paths snaptest-ca.crt
```

#### 2. Load a plugin
Download a plugin, for example snap-plugin-collector-cpu:

```
$ wget  http://snap.ci.snap-telemetry.io/plugins/snap-plugin-collector-cpu/latest/linux/x86_64/snap-plugin-collector-cpu
$ chmod u+x snap-plugin-collector-cpu
```

Load a plugin providing paths to plugin-key and plugin-cert:

```sh
$ snaptel plugin load --plugin-cert=snaptest-srv.crt --plugin-key=snaptest-srv.key --plugin-ca-certs=snaptest-ca.crt snap-plugin-collector-cpu

Plugin loaded
Name: cpu
Version: 7
Type: collector
Signed: false
Loaded Time: Mon, 14 Aug 2017 22:25:16 PDT
```

### Common mistakes
Notice that only GRPC plugins are supported to setup TLS communication. There is also a requirement to use trusted CA and providing both plugin-cert and plugin-key.
Below common error messages are presented that you might receive if one of those requirements are not fulfilled.

#### Case 1: Missing plugin key

```sh
$ snaptel plugin load --plugin-cert=snaptest-srv.crt  --plugin-ca-certs=snaptest-ca.crt ../snap-plugin-lib-go/rand-collector

Error: Both plugin certification and key are mandatory.
Usage: load <plugin_path> [--plugin-cert=<plugin_cert_path> --plugin-key=<plugin_key_path> --plugin-ca-certs=<ca_cert_paths>]
```

#### Case 2: Using untrusted CA

```sh
$ snaptel plugin load --plugin-cert=snaptest-srv.crt --plugin-key=snaptest-srv.key --plugin-ca-certs=snaptest-ca.crt ../snap-plugin-lib-go/rand-collector

Error: rpc error: code = Internal desc = connection error: desc = "transport: authentication handshake failed: x509: certificate signed by unknown authority"
Usage: load <plugin_path> [--plugin-cert=<plugin_cert_path> --plugin-key=<plugin_key_path> --plugin-ca-certs=<ca_cert_paths>]
```

#### Case 3: Trying to set TLS GRPC communication for non-GRPC plugin

```sh
$ snaptel plugin load --plugin-cert snaptest-srv.crt --plugin-key snaptest-srv.key --plugin-ca-certs snaptest-ca.crt ../snap/snap-plugin-collector-mock1

Error: secure framework can't connect to insecure plugin; plugin_name: mock
Usage: load <plugin_path> [--plugin-cert=<plugin_cert_path> --plugin-key=<plugin_key_path> --plugin-ca-certs=<ca_cert_paths>]
```
