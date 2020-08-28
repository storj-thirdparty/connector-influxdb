# connector-influxdb (uplink v1.0.5)

[![Codacy Badge](https://api.codacy.com/project/badge/Grade/d0e808e60a4a4ab79c9fa0fd188b3171)](https://app.codacy.com/gh/storj-thirdparty/connector-influxdb?utm_source=github.com&utm_medium=referral&utm_content=storj-thirdparty/connector-influxdb&utm_campaign=Badge_Grade_Dashboard)
[![Go Report Card](https://goreportcard.com/badge/github.com/utropicmedia/storj-influxdb)](https://goreportcard.com/report/github.com/utropicmedia/storj-influxdb)
![Cloud Build](https://storage.googleapis.com/storj-utropic-services-badges/builds/connector-influxdb/branches/master.svg)

## Overview

The InfluxDB Connector connects to an InfluxDB database, takes a backup of the specified database and uploads the backup files on Storj network.

```bash
Usage:
  connector-influxdb [command] <flags>

Available Commands:
  help        Help about any command
  store       Command to upload data to a Storj V3 network.
  version     Prints the version of the tool

```

`store` - Connect to the specified database(default: `influxdb_property.json`). Back-up of the database is generated using tooling provided by influxdb and then uploaded to the Storj network. Connect to a Storj v3 network using the access specified in the Storj configuration file(default: `storj_config.json`).

Back-up files are iterated through and upload one by one to the Storj network.

The following flags  can be used with the `store` command:

* `accesskey` - Connects to the Storj network using a serialized access key instead of an API key, satellite url and encryption passphrase.
* `share` - Generates a restricted shareable serialized access with the restrictions specified in the Storj configuration file.

Sample configuration files are provided in the `./config` folder.

## Requirements and Install

To build from scratch, [install the latest Go](https://golang.org/doc/install#install).

> Note: Ensure go modules are enabled (GO111MODULE=on)

### Option #1: clone this repo (most common)

To clone the repo

```
git clone https://github.com/storj-thirdparty/connector-influxdb.git
```

Then, build the project using the following:

```
cd connector-influxdb
go build
```

### Option #2:  ``go get`` into your gopath

To download the project inside your GOPATH use the following command:

```
go get github.com/storj-thirdparty/connector-influxdb
```

## Run (short version)

Once you have built the project run the following commands as per your requirement:

### Get help

```
$ ./connector-influxdb --help
```

### Check version

```
$ ./connector-influxdb --version
```

### Create backup from InfluxDB and upload to Storj

```
$ ./connector-influxdb store
```

## Flow Diagram

![Flow Diagram](/_images/arch.drawio.png ':include :type=iframe width=100% height=1000px')
