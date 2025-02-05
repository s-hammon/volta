# Volta

[![license](https://img.shields.io/github/license/s-hammon/volta)](https://github.com/s-hammon/volta/blob/master/LICENSE)
[![Volta CI](https://github.com/s-hammon/volta/actions/workflows/ci.yaml/badge.svg)](https://github.com/s-hammon/volta/actions/workflows/ci.yaml)
[![Go report
Card](https://goreportcard.com/badge/github.com/s-hammon/volta)](https://goreportcard.com/report/github.com/s-hammon/volta)

Volta is an HL7 message parsing service.

### Overview

- WIP

# Install

- WIP

# Usage

<details>
<summary>Click to show <code>volta help</code> output</summary>

```
Usage:
  volta [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  serve       Start the Volta service

Flags:
  -h, --help   help for volta

Use "volta [command] --help" for more information about a command.
```

</details>

## serve

Starts the parsing service.

    $ volta serve -d $DATABASE_URL
    $ {"level":"info","host":"localhost","port":"8080","message":"service configuration"}

You can specify the hostname/port with the `-H`/`-p` flags, respectively. Otherwise, Volta will use the default `localhost:8080`. You must provide the database URI with `-d`.

# Application Default Credentials

This project feches HL7 messages from the [Google Cloud Healthcare API](https://cloud.google.com/healthcare-api/docs), which requires setting up Application Default Credentials (ADC) in the development and production environments. This service does not use/issue API keys, for reasons I'm sure that are related to SOC2 standards. To learn/review how to set up ADC, please check out [Set up Application Default Credentials](https://cloud.google.com/docs/authentication/provide-credentials-adc).

