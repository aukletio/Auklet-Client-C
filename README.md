# Auklet for C

<a href="https://www.apache.org/licenses/LICENSE-2.0" alt="Apache page link -- Apache 2.0 License"><img src="https://img.shields.io/pypi/l/auklet.svg" /></a>
<a href="https://codeclimate.com/repos/5a96d367b192b3261b0003ce/maintainability"><img src="https://api.codeclimate.com/v1/badges/418ddb355b1b344f8c6e/maintainability" /></a>
<a href="https://codeclimate.com/repos/5a96d367b192b3261b0003ce/test_coverage"><img src="https://api.codeclimate.com/v1/badges/418ddb355b1b344f8c6e/test_coverage" /></a>

This is the official C client for Auklet. It officially supports C
and C++, and runs on most POSIX-based operating systems (Debian, 
Ubuntu Core, Raspbian, QNX, etc).

## Features

[auklet_site]: https://app.auklet.io
[auklet_releaser]: https://github.com/aukletio/Auklet-Releaser-C
[auklet_agent]: https://github.com/aukletio/Auklet-Agent-C
[mail_auklet]: mailto:hello@auklet.io

- Automatic crash reporting
- Automatic Function performance issue reporting
- Location, system architecture, and system metrics identification for all 
issues
- Ability to define data usage restriction

## Device Requirements

Auklet's IoT C/C++ agent is built to run on any POSIX operating system. If 
you don't see the OS or CPU architecture you are using for your application 
listed below, and are wondering if Auklet will be compatible, please hit us 
up at [hello@auklet.io][mail_auklet]. 

Validated OSes:

- Debian 8.6
- Fedora 24
- Linaro 4.4.23
- OpenWRT 3.8.3
- Rasbian Jessie 4.9.30 
- Rasbian Stretch 4.14.71
- Ubuntu 16.04
- Yocto 2.2-r2

Validated CPU architectures:

- x86-64
- ARM7
- ARM64
- MIPS

Lastly, don't forget to ensure that your device is connected to the Internet.


## Prerequisites

Before an application can send data to Auklet it needs to be integrated with 
the Auklet library, **libauklet.a**, and then released to Auklet. See the 
README for the [Auklet Agent][auklet_agent] for integration instructions, and
the README for the [Auklet Releaser][auklet_releaser] for releasing 
instructions.

The Auklet-Client assumes the following things: 
- Auklet-Client assumes it is the parent process of your program.
- Auklet-Client assumes that it has permission to create files in ./.auklet.
- Auklet-Client assumes that the user running the Auklet integrated app has 
permissions to read and write the current directory.
- Auklet-Client assumes that it has Internet access.

## Quickstart

### Ready to Go Architectures

If you don't see your architected listed, it doesn't mean we can't support it
so please reach out to [hello@auklet.io][mail_auklet].

- [ARM7](https://s3.amazonaws.com/auklet/client/latest/auklet-client-linux-arm-latest)  
- [ARM64](https://s3.amazonaws.com/auklet/client/latest/auklet-client-linux-arm64-latest)
- [MIPS](https://s3.amazonaws.com/auklet/client/latest/auklet-client-linux-mips-latest)
- [MIPS64](https://s3.amazonaws.com/auklet/client/latest/auklet-client-linux-mips64-latest)
- [MIPS64le](https://s3.amazonaws.com/auklet/client/latest/auklet-client-linux-mips64le-latest)    
- [MIPSle](https://s3.amazonaws.com/auklet/client/latest/auklet-client-linux-mipsle-latest)
- [x86-64](https://s3.amazonaws.com/auklet/client/latest/auklet-client-linux-amd64-latest)

### Getting Started

1. Download the appropriate client from the list above based on your 
   architecture (contact us if your architecture isn't listed) and add it to 
   your deployment package
1. Follow the [C/C++ Agent Quickstart Guide][auklet_agent] to integrate the 
   C/C++ agent that will monitor performance issues in your app and securely 
   transmit the information to the client
1. Configure the systems you are deploying your application to with the 
   following environment variables (the same ones used with the [Auklet 
   Releaser][auklet_releaser])
   - AUKLET_APP_ID
   - AUKLET_API_KEY
1. Deploy your updated package and execute your application using the Auklet 
   Client
   
        ./path/to/Auklet-Client ./path/to/executable
   
And with that, you should be seeing live data for your application come into 
the [Auklet website][auklet_site]!

## Advanced Settings

### Logging

The Auklet-Client opens an anonymous `SOCK_STREAM` Unix domain socket to which
newline-delimited JSON messages can be written.  If `Auklet-Client` confirms 
that the executable has been released, the child process will inherit the 
socket as file descriptor 3. Otherwise, the child process will not inherit 
the file descriptor. Messages written to the socket are transported without 
checking for syntax errors and will be accessible via the user interface.

Here's a C program demonstrating how to use the socket:

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
		if (-1 == fstat(fd, &buf))
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