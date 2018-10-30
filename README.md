# Auklet for C

<a href="https://www.apache.org/licenses/LICENSE-2.0" alt="Apache page link -- Apache 2.0 License"><img src="https://img.shields.io/pypi/l/auklet.svg" /></a>
<a href="https://codeclimate.com/repos/5a96d367b192b3261b0003ce/maintainability"><img src="https://api.codeclimate.com/v1/badges/418ddb355b1b344f8c6e/maintainability" /></a>
<a href="https://codeclimate.com/repos/5a96d367b192b3261b0003ce/test_coverage"><img src="https://api.codeclimate.com/v1/badges/418ddb355b1b344f8c6e/test_coverage" /></a>

Auklet is a profiler for IoT and embedded Linux apps. Like conventional 
benchtop C/C++ profilers, it is implemented as a library that you can link 
your program against. Unlike benchtop profilers, it is meant to be run in 
production and to continuously generate performance metrics. 


# Auklet Client

[auklet_site]: https://app.auklet.io
[auklet_releaser]: https://github.com/aukletio/Auklet-Releaser-C

Auklet's IoT client (`client`) is a command-line program that runs any program
compiled with the Auklet agent and continuously sends live profile data 
viewable on the [Auklet website][auklet_site].

## Device Requirements

Auklet's IoT C/C++ agent is built to run on any POSIX operating system. It
has been validated on:

- Debian 8.6
- Fedora 24
- Linaro 4.4.23
- OpenWRT 3.8.3
- Rasbian Jessie 4.9.30 
- Rasbian Stretch 4.14.71
- Ubuntu 16.04
- Yocto 2.2-r2

Auklet has also been validated for the following CPU architectures:

- ARM7
- ARM64
- MIPS
- x86-64

Lastly, don't forget to ensure that your device is connected to the Internet.


## Prerequisites

Before an application can send data to Auklet it needs to be integrated with 
the Auklet library, **libauklet.a**, and then released to Auklet. See the 
README for the [Auklet agent](https://github.com/aukletio/Auklet-Agent-C) for
integration instructions, and the README for the 
[Auklet releaser][auklet_releaser] for 
releasing instructions.

## Deploying Auklet

1. Download the client which matches the architecture of target devices

    ARM7
   
        curl https://s3.amazonaws.com/auklet/client/latest/auklet-client-linux-arm-latest > Auklet-Client         
     
    ARM64
    
        curl https://s3.amazonaws.com/auklet/client/latest/auklet-client-linux-arm64-latest > Auklet-Client    
    
    MIPS
    
        curl https://s3.amazonaws.com/auklet/client/latest/auklet-client-linux-mips-latest > Auklet-Client
    
    MIPS64
    
        curl https://s3.amazonaws.com/auklet/client/latest/auklet-client-linux-mips64-latest > Auklet-Client
    
    MIPS64le
    
        curl https://s3.amazonaws.com/auklet/client/latest/auklet-client-linux-mips64le-latest > Auklet-Client
    
    MIPSle
    
        curl https://s3.amazonaws.com/auklet/client/latest/auklet-client-linux-mipsle-latest > Auklet-Client
    
    x86-64
    
        curl https://s3.amazonaws.com/auklet/client/latest/auklet-client-linux-amd64-latest > Auklet-Client

1. Send the following files to the target devices
   - Application executable
   - Auklet-Client
   - A file with the Auklet environment variables (such as the **.env** file 
   from the [Auklet releaser][auklet_releaser] setup)

## Running with Auklet

Before running the application, assign an Auklet configuration using the 
Auklet environment variables sent to the device. For example

    . .env
*Note: You'll need to re-initialize your Auklet variables after a device is 
rebooted.
    
**Auklet-Client** assumes that it is the parent process of your program, and 
the instrument assumes that it is in the child process of the client. As 
such, run the application through the Auklet-Client

        ./path/to/Auklet-client ./path/to/executable arg1 arg2...

Auklet-Client assumes three other things: 
- Auklet-Client assumes that it has permission to create files in ./.auklet
- Auklet-Client assumes that the user running the Auklet integrated app has 
permissions to read and write the current directory.
- Auklet-Client assumes that it has Internet access.
   
And with that, you should be seeing live data for your application come into 
the [Auklet website][auklet_site]! :-)