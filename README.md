# syslog-redirector

A statically compiled tool to spawn an executable redirecting its
standard out and standard error to syslog.

Static compilation is important, at least in the intended Helios/Docker
use case, as we do not want to make any assumptions as to the contents
of the container and the availability or location of libc.  Specifically,
BusyBox and Ubuntu containers have libc in different places, and I'm sure
there are others.

# Build

Note: Requires docker.

```shell
make
```

# Installation

For Helios usage, it expects syslog-redirector to live in `/usr/lib/helios`,
so copy it there.  Otherwise, put it wherever you like.

# Developing

Prerequisites:

* syslog listening to a port somewhere you can talk to

It includes two slightly modified files from the golang library, as they
had bugs at the time with respect to reconnecting if I recall correctly.
Newer versions of golang may not have this problem, and so these copied and
modified files may be superfluous now, but I just haven't gotten around to
looking and testing, and I know this current one works and haven't had to
touch it, save for opensourcing in a great many moons.

Also, we submodule the go distribution, as we've run into issues
coaxing go to create a static executable.  Yes, it's supposed to just
do this, but it's not doing it reliably depending on exactly which go
version you had, and we wanted a reproducible build.  The go we
submodule is linux-amd64.  If care less about the static-buildness,
adjust the build command above.  If you find a better solution to
this, pull requests are most definitely welcome.

# Use With Helios

When starting the Helios Agent, pass the `--syslog-redirect-to` switch
with the `host:port` of where syslog is running on the agent machine.

# Use Outside of Helios
```shell
usage: ./syslog-redirector -h syslog_host:port -n name -- executable [arg ...]
  -h="": Host port of where to connect to the syslog daemon
  -n="": Name to log as
  -t=false: use TCP instead of UDP (the default) for syslog communication
  -tee=false: also write to stdout and stderr
```

Example:
```shell
./syslog-redirector -h 10.0.3.1:6514 -n test-ls-thingy -- ls -l
```
This would run the command
```shell
ls -l
```
and redirect it's stdout and stderr to TCP syslog running at `10.0.3.1:6514`
and the name that will show up in the syslogs is `test-ls-thingy`.

Normally, it uses TCP connections to syslog, but the `-t` switch can be used
to tell it to use UDP instead.

Standard out will be logged as `LOG_INFO` and standard error will be
logged as `LOG_ERROR`.


