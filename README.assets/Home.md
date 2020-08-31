## Flow Diagram

![](https://github.com/utropicmedia/storj-influxdb/blob/master/README.assets/arch.drawio.png)

## Config Files

There are two config files that contain Storj network and InfluxDB connection information.  The tool is designed so you can specify a config file as part of your tooling/workflow.  

##### `db_property.json`

Inside the `./config` directory there is a  `db_property.json` file, with following information about your InfluxDB instance:

* `hostName`- Host Name connect to InfluxDB
* `port` - Port connect to InfluxDB
* `username` - User Name of InfluxDB
* `password` - Password of InfluxDB
* `database` - InfluxDB Database Name
* `influxdExeutablePath`- Path to the influxd executable including the executble name

##### `storj_config.json`

Inside the `./config` directory a `storj_config.json` file, with Storj network configuration information in JSON format:

* `apiKey` - API Key created in Storj Satellite GUI
* `satelliteURL` - Storj Satellite URL
* `encryptionPassphrase` - Storj Encryption Passphrase.
* `bucketName` - Name of the bucket to upload data into.
* `uploadPath` - Path on Storj Bucket to store data (optional) or "" or "/". (mandatory)
* `serializedAccess` - Serialized access shared while uploading data used to access bucket without API Key
* `allowDownload` - Set *true* to create serialized access with restricted download
* `allowUpload` - Set *true* to create serialized access with restricted upload
* `allowList` - Set *true* to create serialized access with restricted list access
* `allowDelete` - Set *true* to create serialized access with restricted delete
* `notBefore` - Set time that is always before *notAfter*
* `notAfter` - Set time that is always after *notBefore*

## Run

Backups are iterated through and upload in 32KB chunks to the Storj network.

The following flags  can be used with the `store` command:

* `accesskey` - Connects to Storj network using instead of Serialized Access Key instead of API key, satellite url and encryption passphrase .
* `shared` - Generates a restricted shareable serialized access with the restrictions specified in the Storj configuration file.

Once you have built the project you can run the following:

##### Get help

```
$ ./connector-influxdb --help
```

##### Check version

```
$ ./connector-influxdb --version
```

##### Create backup from InfluxDB and upload them to Storj

```
$ ./connector-influxdb store --influx <path_to_influx_config_file> --storj <path_to_storj_config_file>
```

##### Create backup files from InfluxDB and upload them to Storj bucket using Access Key

```
$ ./connector-influxdb store --accesskey
```

##### Create backup files from InfluxDB and upload them to Storj and generate a Shareable Access Key based on restrictions in `storj_config.json`.

```
$ ./connector-influxdb store --share
```

##### Create backup files from InfluxDB and upload them to Storj, then download them to `./debug` folder.

```
$ ./connector-influxdb store --debug --influx <path_to_influx_config_file> --storj <path_to_storj_config_file>
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
		* InfluxDB version: v1.8.2

	* macOS Catalina
		* Version: 10.15.4
		* Processor: 2.5 GHz Dual-Core Intel Core i5
		* InfluxDB version: v1.8.2

	* ubuntu
		* Version: 16.04 LTS
		* Processor: AMD A6-7310 APU with AMD Radeon R4 Graphics × 4
		* InfluxDB version: v1.8.2
```
