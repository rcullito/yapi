> We should have some ways of connecting programs like garden hose--screw in 
> another segment when it becomes when it becomes necessary to massage data in another way. 
> This is the way of IO also. ~ M. Douglas McIlroy - October 11, 1964  

## yapi - Yet Another Pipe Implementation

[yapi](http://github.com/cmfatih/yapi) is an application inspired by Unix pipeline. 
Currently it can execute remote system commands using ssh protocol. See [examples](#examples)  

For your information; yapi is **still** under heavy development. 
Feedbacks and pull requests are welcome.  

### Installation

#### Binary distributions

Latest release **v0.3.2** - see all [releases](https://github.com/cmfatih/yapi/releases)

| Linux | Windows | Mac OSX | FreeBSD | Source | Android |
|:---:|:---:|:---:|:---:|:---:|:---:|
| [64bit](https://github.com/cmfatih/yapi/releases/download/v0.3.2/yapi-linux-amd64.tar.gz) | [64bit](https://github.com/cmfatih/yapi/releases/download/v0.3.2/yapi-windows-amd64.zip) | [64bit](https://github.com/cmfatih/yapi/releases/download/v0.3.2/yapi-darwin-amd64.tar.gz)* | [64bit](https://github.com/cmfatih/yapi/releases/download/v0.3.2/yapi-freebsd-amd64.tar.gz)* | [tar.gz](https://github.com/cmfatih/yapi/archive/v0.3.2.zip) | - |
| [32bit](https://github.com/cmfatih/yapi/releases/download/v0.3.2/yapi-linux-386.tar.gz)* | [32bit](https://github.com/cmfatih/yapi/releases/download/v0.3.2/yapi-windows-386.zip)* | [32bit](https://github.com/cmfatih/yapi/releases/download/v0.3.2/yapi-darwin-386.tar.gz)* | [32bit](https://github.com/cmfatih/yapi/releases/download/v0.3.2/yapi-freebsd-386.tar.gz)* | [zip](https://github.com/cmfatih/yapi/archive/v0.3.2.zip) | - |
| [arm](https://github.com/cmfatih/yapi/releases/download/v0.3.2/yapi-linux-arm.tar.gz)* | - | - | [arm](https://github.com/cmfatih/yapi/releases/download/v0.3.2/yapi-freebsd-arm.tar.gz)* | - | [arm](#android) |

The binary files marked with `*` are compiled on Linux (64bit) with Go cross compiling support. 
Please [compile](#compile-source) your own yapi binary for better performance.
See [known issues](#known-issues)

#### Compile source

[Download the Go distribution](http://golang.org/doc/install) 
or 
[compile the Go source](http://golang.org/doc/install/source)

```
git clone https://github.com/cmfatih/yapi.git
cd yapi/ && go build yapi.go
```

### Getting started

* Do not forget to make necessary changes in `pipe.json` file before use `yapi` 
  See [config](#config)  
* Add the path of yapi binary file to the `PATH` environment variable or 
  use `./yapi` on Unix-like systems.

-
![yapi-figure-rsce](docs/img/figure-yapi-rsceoy-ccem.png "Remote System Command Execution on yapi")
-

### Usage

#### Help

```
./yapi --help
```

#### Options

```
  -cc       : Client command that will be executed.

  -cn       : Client name(s) those will be connected.
              Use comma (,) for multi-client.

  -cg       : Client group name(s) those will be connected. 
              Use comma (,) for multi-group.

  -ccem     : Execution method for client command. Default; serial
              Possible values; serial (~), parallel (//)

  -pc       : Pipe configuration file. Default; pipe.json

  -ssh      : Simple SSH client command execution.
              It uses the current/given username and HOME/.ssh/id_rsa 
              for the private key file.
              Syntax: [user@]host[:22]

  -help     : Display help and exit.
  -h

  -version  : Display version information and exit.
  -v
```

#### Examples

```
yapi -cc ls
```
It loads `pipe.json` file in the current working directory. Determine the client by `isDefault` value, 
executes `ls` command on the remote system and displays output.

-

```
yapi -cc "top -b -n 1" | grep ssh
```
It executes `top -b -n 1` command on the **remote system**,transfer result to the **host system**, 
executes `grep ssh` on the **host system** and displays output.

-

```
yapi -cc "tail -F /var/log/syslog" -ccem parallel
```
It tails `/var/log/syslog` file on the remote system and wait until the host process exit.

-

```
yapi -cc hostname -cn "client1,client2" -ccem parallel
```
It executes `hostname` command on the **remote systems** `client1` and `client2`,
and displays output on the **host system**.

-

```
yapi -cc hostname -cg group1 -ccem parallel
```
It executes `hostname` command on the **remote systems** which are part of the `group1` group, 
and displays output on the **host system**.

-

```
yapi -cc "ps aux" -cn client1 | yapi -cc "wc -l" -cn client2
```
It executes `ps aux` command on the **remote system** `client1`, 
transfer result to the **remote system** `client2`, counts the lines (`wc -l`)
and displays output on the **host system**.

-

##### Examples for `-ssh` option
`-ssh` option doesn't require `pipe.json` file. 
```
yapi -ssh localhost -cc ls
yapi -ssh user@localhost:22 -cc ls
yapi -ssh host1,host2 -cc ls -ccem parallel
```

-

##### Examples for stdin
```
// Unix-like systems
ls | yapi -cc "wc -l"
echo hello | yapi -cc "wc -c"
yapi -cc "wc -w" < README.md
yapi -cc "wc -w" << EOF
yapi -cc "wc -c" <<< hello

// Windows
dir | yapi -cc "wc -l"
echo hello | yapi -cc "wc -c"
yapi -cc "wc -w" < README.md
```

#### Config

yapi checks `-pc` option or `pipe.json` file in the current working directory 
for the pipe configuration file. The pipe configuration file contains information about 
clients which is used for remote system connections. Here is the default `pipe.json` file.

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

For ssh clients; `name` and `address` should be defined. Address can be `host` or `host:port`
If `username` is not defined then current user will be used for authentication.
`password` and `keyfile` are optional and can be used individually or together.  

If you want to use a PuTTY key (`.ppk`) then you have to 
[convert](https://www.google.com/search?q=how+to+convert+ppk+to+id_rsa) it to an OpenSSH key. 
Simply you can use `puttygen YOURKEYFILE.ppk -o id_rsa -O private-openssh` command.
If you get a `structure error` message due key file then please do not use `.ppk` key file.
Create your own SSH keys. See 
[how](https://www.digitalocean.com/community/articles/how-to-set-up-ssh-keys--2)

#### Android

Yapi works on Android systems;

1. You need [Android Debug Bridge](http://developer.android.com/tools/help/adb.html)
2. Extract the latest Linux arm binary ([yapi-linux-arm.tar.gz](https://github.com/cmfatih/yapi/releases)) 
   release file into a folder.
3. Push the files to the device;
   `adb push yapi /data/local/tmp/` and `adb push pipe.json /data/local/tmp/`
4. Connect to the device; `adb shell`
5. Go to the folder; `cd /data/local/tmp`
6. Change the file mode; `chmod 751 yapi`
7. `./yapi --help`

### Notes

* For issues see [Issues](https://github.com/cmfatih/yapi/issues)
* For coding and design goals see [CODING.md](https://github.com/cmfatih/yapi/blob/master/CODING.md)
* For all notable references see [REFERENCES.md](https://github.com/cmfatih/yapi/blob/master/REFERENCES.md)

### Known Issues

* yapi detects current username (if it is undefined) for SSH clients. But some of the yapi binaries
  are compiled on Linux (64bit) with Go cross compiling support so some platforms have issue (due cgo)
  with username detection. Please build your own yapi binary if you are using; 
  Mac OSX, FreeBSD or Linux (32bit)
  For more details see [golang issue 6376](https://code.google.com/p/go/issues/detail?id=6376)

### Changelog

For all notable changes see [CHANGELOG.md](https://github.com/cmfatih/yapi/blob/master/CHANGELOG.md)

### Contribution

Pull requests are welcome.

### License

Copyright (c) 2014 Fatih Cetinkaya (http://github.com/cmfatih/yapi)  
Licensed under The MIT License (MIT)  
For the full copyright and license information, please view the LICENSE.txt file.