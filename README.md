# Auklet for C

<a href="https://www.apache.org/licenses/LICENSE-2.0" alt="Apache page link -- Apache 2.0 License"><img src="https://img.shields.io/pypi/l/auklet.svg" /></a>
[![Maintainability](https://api.codeclimate.com/v1/badges/5d3e8a3cc277bef22f5f/maintainability)](https://codeclimate.com/github/aukletio/Auklet-Client-C/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/5d3e8a3cc277bef22f5f/test_coverage)](https://codeclimate.com/github/aukletio/Auklet-Client-C/test_coverage)

This is the C client for Auklet. It officially supports C and C++, and runs
on most POSIX-based operating systems (Debian, Ubuntu Core, Raspbian, QNX, etc).

## Features

[auklet_site]: https://app.auklet.io
[auklet_releaser]: https://github.com/aukletio/Auklet-Releaser-C
[auklet_agent]: https://github.com/aukletio/Auklet-Agent-C
[mail_auklet]: mailto:hello@auklet.io

- Automatic crash reporting
- Automatic function performance issue reporting
- Location, system architecture, and system metrics identification for all
issues
- Ability to define data usage restrictions

## Device Requirements

Auklet's C/C++ client is built to run on any POSIX operating system. If
you don't see the OS or CPU architecture you are using for your application
listed below, and are wondering if Auklet will be compatible, please hit us
up at [hello@auklet.io][mail_auklet].

##### Validated OSes:

- Debian 8.6
- Fedora 24
- Linaro 4.4.23
- OpenWRT 3.8.3
- Rasbian Jessie 4.9.30
- Rasbian Stretch 4.14.71
- Ubuntu 16.04
- Yocto 2.2-r2

##### Validated CPU Architectures:

- x86-64
- ARM7
- ARM64
- MIPS

### Networking
Auklet is built to work in network-constrained environments. It can operate
while devices are not connected to the internet and then upload data once
connectivity is reestablished. Auklet can also work in non-IP-based
environments as well. For assistance with getting Auklet running in a
non-IP-based environment contact [hello@auklet.io][mail_auklet].

## Prerequisites

Before an application can send data to Auklet, it needs to be integrated with
the Auklet library, **libauklet.a**, and then released to Auklet. See the
README for the [Auklet Agent][auklet_agent] for integration instructions, and
the README for the [Auklet Releaser][auklet_releaser] for releasing
instructions.

The Auklet client assumes two things:
- It is the parent process of your program.
- The user running the Auklet integrated app has permissions to read and
write the current directory.

## Quickstart

### Ready to Go Architectures

If you don't see your architecture listed, it doesn't mean we can't support it,
so please reach out to [hello@auklet.io][mail_auklet].

- [ARM7](https://s3.amazonaws.com/auklet/client/latest/auklet-client-linux-arm-latest)  
- [ARM64](https://s3.amazonaws.com/auklet/client/latest/auklet-client-linux-arm64-latest)
- [MIPS](https://s3.amazonaws.com/auklet/client/latest/auklet-client-linux-mips-latest)
- [MIPS64](https://s3.amazonaws.com/auklet/client/latest/auklet-client-linux-mips64-latest)
- [MIPS64le](https://s3.amazonaws.com/auklet/client/latest/auklet-client-linux-mips64le-latest)    
- [MIPSle](https://s3.amazonaws.com/auklet/client/latest/auklet-client-linux-mipsle-latest)
- [x86-64](https://s3.amazonaws.com/auklet/client/latest/auklet-client-linux-amd64-latest)

### Getting Started

1. Download the appropriate client from the list above, based on your
   architecture, and add it to your deployment package.
1. Follow the [C/C++ Agent Quickstart Guide][auklet_agent] to integrate the
   C/C++ agent.
1. Configure the systems to which you are deploying with the following
   environment variables (the same ones used with the
   [Auklet    Releaser][auklet_releaser]):
   - `AUKLET_APP_ID`
   - `AUKLET_API_KEY`
1. Deploy your updated package and execute your application using the Auklet
   client:

        ./path/to/Auklet-Client ./path/to/<InsertYourApplication>

And with that, Auklet is ready to go!

## Questions? Problems? Ideas?

To get support, report a bug or suggest future ideas for Auklet, go to
https://help.auklet.io and click the blue button in the lower-right corner to
send a message to our support team.
