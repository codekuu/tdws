# Temporal.io Dynamic Worker Spawner (TDWS)

Temporal.io Dynamic Worker Spawner (TDWS) is a temporal worker that downloads workflows and activities from a remote git repository and loads them into the worker.

## Content

- [Reason for TDWS](#reason-for-tdws)
- [How it works](#how-it-works)
- [Configuration](#configuration)
- [Build a module](#build-a-module)
- [Build the TDWS binary](#build-the-tdws-binary)

## Reason for TDWS

We wanted a tool that could spawn a worker but use different workflows and activities based on the environment.
Since this would require us to write a separate worker for each environment, we decided to write a tool that could spawn a worker and download the workflows and activities from a remote git repository, convert them into plugins and then load them into the worker.

## How it works

1. TDWS is started with a configuration file that contains the "modules" that the worker should load. A module is a git repository that contains workflows and activities. (If you don't have a git repository, you can place the modules in the storage directory that is specified in the configuration file (default is ./tdws-storage))
2. TDWS clones the git repository and builds a plugin out of main.go in the repository.
3. TDWS loads the plugin and calls the TdwsRegister function with the worker as an argument.
4. The TdwsRegister function registers the workflows and activities with the worker.
5. TDWS then starts the worker.

## Configuration

TDWS is configured using a configuration file. The configuration file is a json file called tdws.json, but this can be set using the TDWS_CONFIG_FILE environment variable.
See [config.go](internal/config/config.go) for the configuration struct.

## Build a module

See [tdws-demo-module-go](https://github.com/codekuu/tdws-demo-module-go) for example and information.

## Build the TDWS binary (Will be provided in each release in the future)

`go build -o tdws cmd/tdws/main.go`
