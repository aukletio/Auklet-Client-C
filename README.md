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
When source files change, it runs `go install ./client`, allowing the developer to find
compile-time errors immediately without needing an IDE.

`autobuild` requires [entr](http://www.entrproject.org/).

# Build

To ensure you have all the correct dependencies, run

	dep ensure

To build and install the client to `$GOPATH/bin`, run

	go install ./client

To run unit tests on the client, run

	go test ./client

# Configure

An Auklet configuration is defined by the following environment variables.

	AUKLET_APP_ID
	AUKLET_API_KEY
	AUKLET_BASE_URL
	AUKLET_BROKERS
	AUKLET_PROF_TOPIC
	AUKLET_EVENT_TOPIC
	AUKLET_LOG_TOPIC

To view your current configuration, run `env | grep AUKLET`.

To make it easier to manage multiple configurations, it is suggested to define
the envars in a shell script named after the configuration; for example,
`.env.staging`.

The variables `AUKLET_API_KEY` and `AUKLET_APP_ID` are likely to be different
among developers, so it is suggested that they be defined in a separate
file, `.auklet`, and sourced from within `.env.staging`. For example:

	$ cat .auklet
	export AUKLET_APP_ID=5171dbff-c0ea-98ee-e70e-dd0af1f9fcdf
	export AUKLET_API_KEY=SM49BAMCA0...

	$ cat .env.staging
	. .auklet
	export AUKLET_BASE_URL=https://api-staging.auklet.io/v1
	export AUKLET_BROKERS=broker1,broker2,broker3
	export AUKLET_PROF_TOPIC=z8u1-profiler
	export AUKLET_EVENT_TOPIC=z8u1-events
	export AUKLET_LOG_TOPIC=z8u1-logs

## `AUKLET_BROKERS`

A comma-delimited list of Kafka broker addresses. For example:

	broker1,broker2,broker3

## `AUKLET_EVENT_TOPIC`, `AUKLET_PROF_TOPIC` `AUKLET_LOG_TOPIC`

Kafka topics to which `client` should send event, profile, and log data, respectively.

## `AUKLET_BASE_URL`

A URL, without a trailing slash, to be used when checking releases.
For example:

	https://api-staging.auklet.io/v1

If not defined, this defaults to the production endpoint.

# Assign a Configuration

	. .env.staging

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
