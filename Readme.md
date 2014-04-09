pulseutil
=========

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
   status	list build status
   help, h	Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --addr 'http://pulse/xmlrpc'	Pulse Remote API endpoint
   --user ''			Pulse user name
   --pass ''			Pulse user password
   --agent, -a '.*'		Agent name patter
   --project, -p '.*'		Project name pattern
   --build, -b '0'		Build number
   --prtg			PRTG-friendly output
   --version, -v		print the version
   --help, -h			show help
```

#### PRTG output

Passing `--prtg` flag makes the output PRTG-friendly - when command exists with:

* exit code 0, the output is:

`0:0:OK`

* exit code different than 0, the output is:

`2:1:"<error message here>"`

#### Examples

###### Performing a health check

```
~ $ pulsecli --prtg --user $USER --pass $PASS health
0:0:OK
```

###### Triggering builds for all the projects

```
~ $ pulsecli --user $USER --pass $PASS trigger
```

###### Triggering builds for all LM-X tiers

```
~ $ pulsecli --user $USER --pass $PASS --project 'LM-X - Tier' trigger
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

###### Getting status of the `LM-X - Release Build - Tier 2` project

The `--build` or `-b` flag expects either: 

  * a real build number
  * 0 which means latest build number
```
~ $ pulsecli --user $USER --pass $PASS -p 'LM-X - Release Build - Tier 2' -b 0 status
LM-X - Release Build - Tier 2 (build 547):
- id: 547
  complete: true
  end: {}
  endunix: "1396963267179"
  errors: 0
  maturity: integration
  owner: LM-X - Release Build - Tier 2
  personal: false
  pinned: false
  progress: -1
  project: LM-X - Release Build - Tier 2
  revision: 887e88a5c4709e9bf260744d398d71dd7ef70050
  reason: manual trigger by rjeczalik
...
```
  * a negative number being an relative offset to the latest build number
```
~ $ pulsecli --user $USER --pass $PASS -p 'LM-X - Release Build - Tier 2' -b -10 status
LM-X - Release Build - Tier 2 (build 537):
- id: 537
  complete: true
  end: {}
  endunix: "1396372083402"
  errors: 8
  maturity: integration
  owner: LM-X - Release Build - Tier 2
  personal: false
  pinned: false
  progress: -1
  project: LM-X - Release Build - Tier 2
  revision: 22fc614bd290041778ad1a69fc66c97841c77177
...
```

#### TODO

* create project on JIRA to track the issues
* session management (save encrypted user/pass to `~/.pulsecli` to spare the user effort typing it each time)
* verbose logging, currently the output is PRTG-friendly, which should be made optional (e.g. using `--prtg` flag)
* `cmd/pulseclid` daemon for watching builds, which will be used for `github.com` bot
