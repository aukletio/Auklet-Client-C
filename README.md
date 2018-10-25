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

After cloning this repo and setting up your Go environment, run this
command to enable pre-commit gofmt checking: `git config core.hookspath
.githooks`.

# Development Tools

`autobuild` is an optional script that can be run in a separate terminal
window.  When source files change, it runs `go install ./cmd/client`,
allowing the developer to find compile-time errors immediately without
needing an IDE.

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

`AUKLET_LOG_ERRORS=true` logs any unexpected but recoverable errors,
such as encoding, filesystem, and network protocol errors.

`AUKLET_LOG_INFO=true` logs significant information or events, such
as broker addresses, remotely acquired configuraton parameters, and
production of messages.

# Assign a Configuration

	. .env

# Run an App

To run an Auklet-enabled executable called `x` (an executable compiled with the
Auklet agent and properly released using the Auklet releaser), run

	client ./x

## Runtime dependencies

`client` assumes the following directory structure:

	./.auklet/
		datalimit.json
		message/

If this structure does not exist, it will be created.

## Remote logging

`client` opens an anonymous `SOCK_STREAM` Unix domain socket to which
newline-delimited JSON messages can be written.  If `client` confirms that the
executable has been released, the child process will inherit the socket as file
descriptor 3. Otherwise, the child process will not inherit the file descriptor.
Messages written to the socket are transported without checking for syntax
errors and will be accessible via the user interface.

Here's a C program demonstrating how to use the socket, assuming the compilation
flags `-std=c99 -pedantic -D_POSIX_C_SOURCE=200809L`:

	#include <fcntl.h>
	#include <stdio.h>
	#include <sys/stat.h>
	#include <sys/types.h>
	#include <unistd.h>

	/* getAukletLogFD checks if file descriptor 3 is valid. If so, it
	 * returns the file descriptor. Otherwise, it opens /dev/null and
	 * returns its file descriptor. */
	int
	getAukletLogFD()
	{
		struct stat buf;
		int fd = 3;
		if (-1 == fstat(fd, &buf)
			fd = open("/dev/null", O_WRONLY);
		return fd;
	}

	int
	main()
	{
		int logFD = getAukletLogFD();
		dprintf(logFD, "{\"message\":\"hello, auklet\"}\n");
		close(logFD);
	}

# Docker Setup

1. Install Docker for Mac Beta.
1. Build your environment with `docker-compose build`.
1. To ensure you have all the correct dependencies, run `docker-compose run auklet dep ensure`.
1. To build and install the client to `$GOPATH/bin`, run `docker-compose run auklet go install ./client`.
1. To test the client, run `docker-compose run auklet go test ./client`.
