<a href="#octopus"><img src="https://github.com/boyvanduuren/octorunner/raw/master/assets/logo/octorunner_01.png" width="250" /></a>

# octorunner [![Go Report Card](https://goreportcard.com/badge/github.com/boyvanduuren/octorunner)](https://goreportcard.com/report/github.com/boyvanduuren/octorunner)

Octorunner is a tool that serves as the endpoint for Github webhooks. It was inspired by [Gitlab's gitlab-runner](https://docs.gitlab.com/runner/).
Right now it can spin up a docker container using a user specified docker image, run some user specified test commands, and set a commit's status. This is all triggered by push events. It works for both public and private repositories.

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

### Docker configuration

The official docker client API is used, so configuration is done as described on [their documentation](https://godoc.org/github.com/docker/docker/client#NewEnvClient).
TLDR: `DOCKER_HOST=http://<ip>:<port>`, `DOCKER_API_VERSION=<version>`, `DOCKER_CERT_PATH=<path>`, `DOCKER_TLS_VERIFY=1`.

### Repository configuration

You'll want to configure access tokens and secrets for the repositories that use `octorunner` as webhook endpoint. The access tokens
will be used to set statuses on commits, and download private repository archives. Secrets are used to make sure an event actually originates from Github, not some unknown
entity that happens to know your endpoint. (see [securing your webhooks](https://developer.github.com/webhooks/securing/))

There are two methods of providing `octorunner` with this information.

Use the `config` file, which can be formatted in any of the previously named languages. (Example in `yaml`):

```yaml
repositories:
  boyvanduuren/octorunner:
    token: YOUR_ACCESS_TOKEN
    secret: YOUR_SECRET
```

Use env vars that are formatted as follows: `OCTORUNNER_account/repository_{TOKEN,SECRET}`. E.g. `OCTORUNNER_boyvanduuren/octorunner_SECRET=foobar`.

### Github configuration

When octorunner has been configured and is running, you'll want to set the configured URL as a webhook endpoint on your github repository. This can be done on `https://github.com/<username>/<project>/settings/hooks`.
[Github recommends ngrok](https://developer.github.com/webhooks/configuring/) to expose your endpoint on the internet, and I found
it works easy enough.

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

* ~~Make sure repository config can be passed as environment variables~~
* Handle pull request events
* Store test output
* Write more and proper tests
* Add an option to setup webhooks automatically

<hr />
<a name="octopus" />
Octopus made by [Stephan Kaffa](https://github.com/stephankaffa)
