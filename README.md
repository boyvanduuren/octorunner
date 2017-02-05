# octorunner

Octorunner is a tool that serves as the endpoint for Github webhooks.
Right now it supports handling of the `push` event, and triggers the running of a simple script in a Docker container, which can be
used for running tests. The status of the commit that was pushed is set according to the exit code of that test.

## Configuration

### Octorunner configuration

Configuration is done via a config file or environment variables. If configured with a file, a file called `config` needs to be placed
in the same directory as the `octorunner` binary. The extension needs to reflect the language used in the configuration file.
JSON, TOML, YAML, HCL, and Java properties are supported (thanks to the [Viper](https://github.com/spf13/viper) library).

The following configuration settings can be set.

* `loglevel`, sets the log level, either `debug`, `error`, `fatal`, or `info` (default: `info`)
* `web.server`, sets the address octorunner should bind on (default: `127.0.0.1`)
* `web.port`, the port octorunner should bind on (default: `8080`)
* `web.path`, the pathname of the payload URL (default: `payload`)

In case you'd like to configure `octorunner` using environment variables, you should capitalize the configuration key, prefix it with `OCTORUNNER_`
and replace `.` with `_` (e.g. `WEB_PORT=8000`)

### Repository configuration

You'll want to configure access tokens and secrets for the repositories that'll use `octorunner` as webhook endpoint. The access tokens
will be used to set statuses on commits, and secrets are used to make sure an event actually originates from Github, not some unknown
entity that happens to know your endpoint. (see [securing your webhooks](https://developer.github.com/webhooks/securing/))

Repository configuration should work via the environment as well, but I'm having some trouble with that right now. You can provide
repository configuration as follows (example in `yaml`):

```yaml
repositories:
  boyvanduuren/octorunner:
    token: YOUR_ACCESS_TOKEN
    secret: YOUR_SECRET
```

## Adding a test to your repository

Tests are quite simple right now. You can specify which docker image should be used for your container, and you can specify
a few commands that need to run and succeed in order for your test to pass.
These settings are specified in a file called `octorunner.yaml` or `octorunner.yml` which should be present at the root of your
repository.
A simple configuration would look like this:

```yaml
image: maven:3-jdk-8
script:
  - mvn test
```
Multiple commands can be configured under the `script` key. Octorunner joins them as `command1 && command2 && ... && commandN`, so
they should all return 0 in order for the test to succeed.

## TODO

* Make sure repository config can be passed as environment variables
* Handle pull request events
* Store test output
* Write more and proper tests
