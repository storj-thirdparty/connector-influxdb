## connector-influxdb (uplink v1.0.5)

[![Go Report Card](https://goreportcard.com/badge/github.com/utropicmedia/storj-influxdb)](https://goreportcard.com/report/github.com/utropicmedia/storj-influxdb)

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



`store` - Connect to the specified database (default: `db_property.json`).  Backups of the database are generated using tooling provided by InfluxDB then uploaded to the Storj network.  Connect to a Storj v3 network using the access specified in the Storj configuration file (default: `storj_config.json`). 

Sample configuration files are provided in the `./config` folder.  Backups are iterated through  and upload  in 32KB chunks to the Storj network.

The following flags  can be used with the `store` command:

* `accesskey` - Connects to Storj network using instead of Serialized Access Key instead of API key, satellite url and encryption passphrase .
* `shared` - Generates a restricted shareable serialized access with the restrictions specified in the Storj configuration file.
* `debug` - Download the uploaded backup files to local disk inside project_folder/debug folder.



## Requirements

To build from scratch, [install Go](https://golang.org/doc/install#install). 

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
go get github.com/utropicmedia/storj-influxdb
```


## Config Files

There are two config files that contain Storj network and InfluxDB connection information.  The tool is designed so you can specify a config file as part of your tooling/workflow.  

Inside the `./config` directory there is a  `db_property.json` file, with following information about your InfluxDB instance:

* `hostName`- Host Name connect to InfluxDB
* `port` - Port connect to InfluxDB
* `username` - User Name of InfluxDB
* `password` - Password of InfluxDB
* `database` - InfluxDB Database Name
* `influxdExeutablePath`- Path to the influxd executable including the executble name



Inside the `./config` directory a `storj_config.json` file, with Storj network configuration information in JSON format:

* `apiKey` - API Key created in Storj Satellite GUI
* `satelliteURL` - Storj Satellite URL
* `encryptionPassphrase` - Storj Encryption Passphrase.
* `bucketName` - Name of the bucket to upload data into.
* `uploadPath` - Path on Storj Bucket to store data (optional) or "/"
* `serializedAccess` - Serialized access shared while uploading data used to access bucket without API Key
* `allowDownload` - Set *true* to create serialized access with restricted download
* `allowUpload` - Set *true* to create serialized access with restricted upload
* `allowList` - Set *true* to create serialized access with restricted list access
* `allowDelete` - Set *true* to create serialized access with restricted delete
* `notBefore` - Set time that is always before *notAfter*
* `notAfter` - Set time that is always after *notBefore*



## Run 

Once you have built the project run the following commands as per your requirement:

##### Get help

```
$ ./storj-influxdb --help
```

##### Check version

```
$ ./storj-influxdb --version
```

##### Create backup from InfluxDB and upload them to Storj

```
$ ./storj-influxdb store --influx <path_to_influx_config_file> --storj <path_to_storj_config_file>
```

##### Create backup files from InfluxDB and upload them to Storj bucket using Access Key

```
$ ./storj-influxdb store --accesskey
```

##### Create backup files from InfluxDB and upload them to Storj and generate a Shareable Access Key based on restrictions in `storj_config.json`.

```
$ ./storj-influxdb store --share
```

##### Create backup files from InfluxDB and upload them to Storj, then download them to `./debug` folder.

```
$ ./storj-influxdb store --debug --influx <path_to_influx_config_file> --storj <path_to_storj_config_file>
```


> **NOTE**: To restore database from the downloaded backup files after running `store` command with       `--debug` flag, you can run the following command:

```
influxd restore -portable -db <old-database-name> -newdb <new-database-name> -host localhost:8088 <path_to_downloaded_backup_files>
```



##  Testing

The project has been tested on the following operating systems:

```
	* Windows
		* Version: 10 Pro
		* Processor: Intel(R) Core(TM) i3-5005U CPU @ 2.00GHz 2.00GHz

	* macOS Catalina
		* Version: 10.15.4
		* Processor: 2.5 GHz Dual-Core Intel Core i5

	* ubuntu
		* Version: 16.04 LTS
		* Processor: AMD A6-7310 APU with AMD Radeon R4 Graphics × 4
```
