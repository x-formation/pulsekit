pulse
=====

<img src="http://zutubi.com/site_media/images/zutubi.png" align="right"><p>Repository for tools / libraries being built around Zutubi Pulse for automation purposes. This most notable tools in this repository are:</p>

## cmd/pulsecli

Command-line tool for communicating with Zutubi Pulse server through its Remote API.

#### Installation

Just go get it:

```
~ $ go get github.com/x-formation/int-tools/pulse/cmd/pulsecli
```

Ensure you have `$GOPATH` set and `$GOBIN` (or `$GOPATH`/bin) is in your `$PATH`. Taking the oportunity you read this, I strongly encourage you to use [gvm](https://github.com/moovweb/gvm).

#### Usage

```
NAME:
   pulsecli - a command-line client for a Pulse server

USAGE:
   pulsecli [global options] command [command options] [arguments...]

VERSION:
   0.1.0

COMMANDS:
   trigger	trigger a build
   health	perform a health check
   projects	list all projcts
   agents	list all agents
   help, h	Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --addr 'http://pulse/xmlrpc'	Pulse Remote API endpoint
   --user ''			Pulse user name
   --pass ''			Pulse user password
   --agent, -a '.*'		Agent name patter
   --project, -p '.*'		Project name pattern
   --version, -v		print the version
   --help, -h			show help
```

#### Examples

* Performing a health check

```
~ $ pulsecli --user $USER --pass $PASS health
0:0:OK
```

* Triggering builds for all the projects

```
~ $ pulsecli --user $USER --pass $PASS trigger
0:0:OK
```

* Triggering builds for all LM-X tiers

```
~ $ pulsecli --user $USER --pass $PASS --project 'LM-X - Tier' trigger
0:0:OK
```

* Listing all the agents

```
~ $ pulsecli --user $USER --pass $PASS agents
AIX - 5.3@http://aix275:8090
FreeBSD 10 - x64@http://freebsd10_x64:8090
HPUX - IA64@http://hpuxia64:8090
Linux - ARM@http://pulse-arm:8090
Linux - CentOS 5.10 - Distrib - x64@http://centos5_x64:8090
...
Windows 8.1 - 6@http://pulse-win-6:8090
Windows 8.1 - 7@http://pulse-win-7:8090
Windows 8.1 - 8@http://pulse-win-8:8090
Windows 8.1 - 9@http://pulse-win-9:8090
```
