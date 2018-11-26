# How to Contribute

## Standards

Auklet is an edge first application performance monitor; therefore, starting
with version 1.0.0 the following compliance levels are to be maintained:

- Automotive Safety Integrity Level B (ASIL B)

## Submissions

If you have found a bug, please go to https://help.auklet.io and click the blue
button in the lower-right corner to report it to our support team.

We are not accepting outside contributions at this time. If you have a feature
request or idea, please open a new issue.

If you've found a security related bug, please do not create an issue or PR.
Instead, email our team directly at [security@auklet.io](mailto:security@auklet.io).

# Working on the Auklet C Client
## Go Setup

The Auklet client needs at least Go 1.8 and [dep][godep] 0.3.2. See the
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

If you want to build the client on Mac OS X, you can install `dep` via
Homebrew by running `brew install dep`, or by changing the above `curl` command
to download `dep-darwin-amd64`.

After cloning this repo and setting up your Go environment, run this
command to enable pre-commit gofmt checking: `git config core.hookspath
.githooks`.

## Build

To ensure you have all the correct dependencies, run:

	dep ensure

To build and install the client to `$GOPATH/bin`, run:

	go install ./cmd/client

To run unit tests on the client, run:

	go test ./config

## Configure

An Auklet configuration is defined by the following environment variables.

	AUKLET_APP_ID
	AUKLET_API_KEY
	AUKLET_LOG_INFO
	AUKLET_LOG_ERRORS

To view your current configuration, run `env | grep AUKLET`.

To make it easier to manage configurations, it is suggested to define the
environment variables in a shell script named after the configuration; for
example, `.env`.

	$ cat .env
	export AUKLET_APP_ID=ABCDEF1234...
	export AUKLET_API_KEY=ABCDEF1234...
	export AUKLET_LOG_INFO=true

### Console Logging

Console logging to stdout is disabled by default. There are two logging levels,
which are controlled by dedicated environment variables. To enable a logging
level, set its environment variable to `true`. To disable it, `unset` the
variable.

`AUKLET_LOG_ERRORS=true` logs any unexpected but recoverable errors,
such as encoding, filesystem, and network protocol errors.

`AUKLET_LOG_INFO=true` logs significant information or events, such
as broker addresses, remotely acquired configuraton parameters, and
production of messages.

## Assign a Configuration

	. .env

## Run an App

To run an Auklet-enabled executable called `x` (an executable compiled with the
Auklet agent and properly released using the Auklet releaser), run:

	client ./x

## Runtime dependencies

The Auklet client assumes the following directory structure:

	./.auklet/
		datalimit.json
		message/

If this structure does not exist, it will be created.

## Docker Setup

1. Install [Docker](www.docker.com/products/docker-desktop).
1. Build your environment with `docker-compose build`.
1. To ensure you have all the correct dependencies, run `docker-compose run auklet dep ensure`.
1. To build and install the client to `$GOPATH/bin`, run `docker-compose run auklet go install ./client`.
1. To test the client, run `docker-compose run auklet go test ./client`.