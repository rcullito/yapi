## yapi - Yet Another Pipe Implementation

[yapi](http://github.com/cmfatih/yapi) is an application inspired by Unix pipeline.  

> We should have some ways of connecting programs like garden hose--screw in 
> another segment when it becomes when it becomes necessary to massage data in another way. 
> This is the way of IO also.  ~ M. D. McIlroy  

Currently yapi can execute remote system commands using ssh protocol. See [examples](#examples)  

For your information; yapi is **still** under heavy development. 
Feedbacks and pull requests are welcome.  

### Installation

#### From binary distributions

* **Version v0.2.5**
  * Linux : 
    [64bit](https://github.com/cmfatih/yapi/releases/download/v0.2.5/yapi-linux-amd64.tar.gz) , 
    [32bit](https://github.com/cmfatih/yapi/releases/download/v0.2.5/yapi-linux-386.tar.gz) , 
    [arm](https://github.com/cmfatih/yapi/releases/download/v0.2.5/yapi-linux-arm.tar.gz)
  * Mac OSX : 
    [64bit](https://github.com/cmfatih/yapi/releases/download/v0.2.5/yapi-darwin-amd64.tar.gz) , 
    [32bit](https://github.com/cmfatih/yapi/releases/download/v0.2.5/yapi-darwin-386.tar.gz)
  * FreeBSD : 
    [64bit](https://github.com/cmfatih/yapi/releases/download/v0.2.5/yapi-freebsd-amd64.tar.gz) , 
    [32bit](https://github.com/cmfatih/yapi/releases/download/v0.2.5/yapi-freebsd-386.tar.gz) , 
    [arm](https://github.com/cmfatih/yapi/releases/download/v0.2.5/yapi-freebsd-arm.tar.gz)
  * Windows 8/7/Vista/XP : 
    [64bit](https://github.com/cmfatih/yapi/releases/download/v0.2.5/yapi-windows-amd64.zip) , 
    [32bit](https://github.com/cmfatih/yapi/releases/download/v0.2.5/yapi-windows-386.zip)

#### From source

For HEAD version (or see [releases](https://github.com/cmfatih/yapi/releases))

```
git clone https://github.com/cmfatih/yapi.git
cd yapi/
go build yapi.go
```

### Usage

#### Test

```
./yapi --help
```
output...
```
Usage: yapi [OPTION]...

yapi - Yet Another Pipe Implementation - v0.2.5

Options:
  -cc         : Client command
  -cg         : Client group(s). Use comma (,) for multiple.
  -cn         : Client name(s). Use comma (,) for multiple.
  -h          : Display this help and exit
  -help       : Display this help and exit
  -pc         : Pipe configuration file
  -v          : Display version information and exit
  -version    : Display version information and exit
...
...
```

#### Getting started

* Do not forget to make necessary changes in `pipe.json` file before use `yapi` 
  See [config](#config)  
* Add the path of yapi binary file to the `PATH` environment variable or 
  use `./yapi` on Unix-like systems.

#### Examples

##### Example 1
```
yapi -cc ls
```
It loads `pipe.json` file in the current working directory. Determine the client by 
`isDefault` value, executes `ls` command on the remote system and displays output.

##### Example 2
```
yapi -pc /path/pipe.json -cc "tail -f /var/log/syslog"
```
It loads `/path/pipe.json`, tails `/var/log/syslog` file on the remote system and 
wait until the host process exit.

##### Example 3
```
yapi -cc "top -b -n 1" | grep ssh
```
It executes `top -b -n 1` command on the **remote system**,
transfer result to the **host system**, executes `grep ssh` on the **host system** 
and displays output.

##### Example 4
```
yapi -cn dev -cc "ps aux" | yapi -cn prod -cc "wc -l"
```
It executes `ps aux` command on the **remote system** `dev`,
transfer result to the **remote system** `prod`, counts the lines (`wc -l`)
and displays output on the **host system**.

##### Example 5
```
yapi -cc hostname -cn "client1,client2"
```
It executes `hostname` command on the **remote systems** `client1` and `client2`,
and displays output on the **host system**.

##### Example 6
```
yapi -cc hostname -cg group1
```
It executes `hostname` command on the **remote systems** which are part of the
`group1` group, and displays output on the **host system**.

#### Config

yapi checks `-pc` option or `pipe.json` file in the current working directory 
for pipe configuration file. Pipe configuration file contains information about 
clients which is used for remote system connections.  

Currently only ssh protocol is supported. For ssh clients; `address` and `username` 
should be defined. `password` and `keyfile` are optional and can be used individually 
or together.

Here is the default `pipe.json` file.

```
{
  "clients": [
    {
      "name": "sshtest",
      "groups": ["test"],
      "kind": "ssh",
      "address": "HOST",
      "auth": {
        "username": "USERNAME",
        "password": "PASSWORD",
        "keyfile": ""
      },
      "isDefault": true
    }
  ]
}
```

### Notes

* For issues see [Issues](https://github.com/cmfatih/yapi/issues)
* For coding and design goals see [CODING.md](https://github.com/cmfatih/yapi/blob/master/CODING.md)
* For all notable references see [REFERENCES.md](https://github.com/cmfatih/yapi/blob/master/REFERENCES.md)

### Changelog

For all notable changes see [CHANGELOG.md](https://github.com/cmfatih/yapi/blob/master/CHANGELOG.md)

### Contribution

Pull requests are welcome.

### License

Copyright (c) 2014 Fatih Cetinkaya (http://github.com/cmfatih/yapi)  
Licensed under The MIT License (MIT)  
For the full copyright and license information, please view the LICENSE.txt file.