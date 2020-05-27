## Flow Diagram

![](https://github.com/utropicmedia/storj-influxdb/blob/master/README.assets/arch.drawio.png)

## Config Files

There are two config files that contain Storj network and InfluxDB connection information. The tool is designed so you can specify a config file as part of your tooling/workflow. 



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

* `apiKey` - API Key created in Storj Satellite GUI(mandatory)
* `satelliteURL` - Storj Satellite URL(mandatory)
* `encryptionPassphrase` - Storj Encryption Passphrase(mandatory)
* `bucketName` - Name of the bucket to upload data into(mandatory)
* `uploadPath` - Path on Storj Bucket to store data (optional) or "/" (mandatory)
* `serializedAccess` - Serialized access shared while uploading data used to access bucket without API Key (mandatory while using *accesskey* flag)
* `allowDownload` - Set *true* to create serialized access with restricted download (mandatory while using *share* flag)
* `allowUpload` - Set *true* to create serialized access with restricted upload (mandatory while using *share* flag)
* `allowList` - Set *true* to create serialized access with restricted list access
* `allowDelete` - Set *true* to create serialized access with restricted delete
* `notBefore` - Set time that is always before *notAfter*
* `notAfter` - Set time that is always after *notBefore*



## Run



Backups are iterated through and upload in 32KB chunks to the Storj network.

The following flags  can be used with the `store` command:

* `accesskey` - Connects to Storj network using instead of Serialized Access Key instead of API key, satellite url and encryption passphrase.
* `shared` - Generates a restricted shareable serialized access with the restrictions specified in the Storj configuration file.
* `debug` - Download the uploaded backup files to local disk inside project_folder/debug folder.



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

##### Create backup files from InfluxDB and upload them to Storj and generate a Shareable Access Key based on restrictions in `storj_config.json`

```
$ ./connector-influxdb store --share
```




## Testing

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