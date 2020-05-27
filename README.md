## connector-influxdb (uplink v1.0.5)

[![Go Report Card](https://goreportcard.com/badge/github.com/storj-thirdparty/connector-influxdb)](https://goreportcard.com/report/github.com/storj-thirdparty/connector-influxdb)

## Overview

The InfluxDB Connector connects to an InfluxDB database, takes a backup of the specified database and uploads the backup files on Storj network.

```bash
Usage:
  connector-influxdb [command] <flags>

Available Commands:
  help        Help about any command
  store       Command to upload data to a Storj V3 network
  version     Prints the version of the tool

```



`store` - Connect to the specified database (default: `db_property.json`). Back-up files of the database are generated using tooling provided by InfluxDB then uploaded to the Storj network. Connect to a Storj v3 network using the access specified in the Storj configuration file (default: `storj_config.json`).



Sample configuration files are provided in the `./config` folder.



## Requirements and Install

To build from scratch, [install the latest Go](https://golang.org/doc/install#install).

> Note: Ensure go modules are enabled (GO111MODULE=on)



#### Option #1: clone this repo (most common)

To clone the repo

```
git clone https://github.com/storj-thirdparty/connector-influxdb.git
```

Then, build the project using the following:

```
cd connector-influxdb
go build
```



#### Option #2:  ``go get`` into your gopath

To download the project inside your GOPATH use the following command:

```
go get github.com/storj-thridparty/connector-influxdb
```



## Run (short version)

Once you have built the project run the following commands as per your requirement:

##### Get help

```
$ ./connector-influxdb --help
```

##### Check version

```
$ ./connector-influxdb --version
```

##### Create backup from InfluxDB and upload to Storj

```
$ ./connector-influxdb store
```



## Documentation

For more information on runtime flags, configuration, testing, and diagrams, check out the [Detail](//github.com/storj-thirdparty/wiki/Detail) or jump to:

* [Config Files](//github.com/storj-thirdparty/connector-influxdb/wiki/#config-files)
* [Run (long version)](//github.com/storj-thirdparty/connector-influxdb/wiki/#run)
* [Testing](//github.com/storj-thirdparty/connector-influxdb/wiki/#testing)
* [Flow Diagram](//github.com/storj-thirdparty/connector-influxdb/wiki/#flow-diagram)
