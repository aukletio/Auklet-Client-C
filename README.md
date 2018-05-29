# Auklet Client

Auklet's IoT client (`client`) is a command-line program that runs any program
compiled with the Auklet agent and continuously sends live profile data to the
Auklet backend. The client is built to run on any POSIX operating system. It has
been validated on:

- Ubuntu 16.04

# Go Setup

`client` needs at least Go 1.8 and [dep][godep] 0.3.2. See the
[getting started page][gs] to download Go. Then see [How to Write Go Code -
Organization][org] to set up your system.

[godep]: https://github.com/golang/dep
[gs]: https://golang.org/doc/install
[org]: https://golang.org/doc/code.html#Organization

Conventionally, your `~/.profile` should contain the following:

	export GOPATH=$HOME/go
	export PATH=$PATH:$GOPATH/bin

The first line tells Go where your workspace is located. The second makes sure
that the shell will know about executables built with `go install`.

After setting up Go on your system, install `dep` by running:

	curl -L -s https://github.com/golang/dep/releases/download/v0.3.2/dep-linux-amd64 -o $GOPATH/bin/dep
	chmod +x $GOPATH/bin/dep

If you want to build `client` on Mac OS X, you can install `dep` via
Homebrew by running `brew install dep`, or by changing the above `curl` command
to download `dep-darwin-amd64`.

# Development Tools

`autobuild` is an optional script that can be run in a separate terminal window.
When source files change, it runs `go install ./cmd/client`, allowing the developer to find
compile-time errors immediately without needing an IDE.

`autobuild` requires [entr](http://www.entrproject.org/).

# Build

To ensure you have all the correct dependencies, run

	dep ensure

To build and install the client to `$GOPATH/bin`, run

	go install ./cmd/client

To run unit tests on the client, run

	go test ./config

# Configure

An Auklet configuration is defined by the following environment variables.

	AUKLET_APP_ID
	AUKLET_API_KEY
	AUKLET_BASE_URL
	AUKLET_LOG_INFO
	AUKLET_LOG_ERRORS

To view your current configuration, run `env | grep AUKLET`.

To make it easier to manage configurations, it is suggested to define the
environment variables in a shell script named after the configuration; for
example, `.env.staging`.

	$ cat .env.staging
	export AUKLET_APP_ID=5171dbff-c0ea-98ee-e70e-dd0af1f9fcdf
	export AUKLET_API_KEY=SM49BAMCA0...
	export AUKLET_LOG_INFO=true

## Base URL

`AUKLET_BASE_URL` defines the endpoint against which the client makes API calls.
If not defined, the default production endpoint is used.

Its format is a URL **without a trailing slash or path.** For example:

	AUKLET_BASE_URL=https://api-staging.auklet.io


## Console Logging

Console logging to stdout is disabled by default. There are two logging levels,
which are controlled by dedicated environment variables. To enable a logging
level, set its environment variable to `true`. To disable it, `unset` the
variable.

`AUKLET_LOG_ERRORS=true` logs any unexpected but recoverable errors, including but
not limited to

- JSON syntax errors
- unexpected HTTP responses
- insufficient filesystem permissions

`AUKLET_LOG_INFO=true` logs significant information or events, including but not
limited to

- Kafka broker list
- remotely acquired configuraton parameters
- production of Kafka messages

# Assign a Configuration

	. .env

# Run an App

To run an Auklet-enabled executable called `x` (an executable compiled with the
Auklet agent and properly released using the Auklet releaser), run

	client ./x

# Docker Setup

1. Install Docker for Mac Beta.
1. Build your environment with `docker-compose build`.
1. To ensure you have all the correct dependencies, run `docker-compose run auklet dep ensure`.
1. To build and install the client to `$GOPATH/bin`, run `docker-compose run auklet go install ./client`.
1. To test the client, run `docker-compose run auklet go test ./client`.
