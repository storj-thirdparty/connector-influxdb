# connector-influxdb

# Developed using uplink v1.0.5

To build from scratch, [install Go](https://golang.org/doc/install#install).

## Download Project

* To download the project inside GOPATH use the following command:
```
go get github.com/utropicmedia/storj-influxdb
```
**NOTE**: To know your GOPATH, run ```go env gopath``` command. Navigate to the downloaded folder inside GOPATH and build the project.

* To download the project at a desired location, navigate to the desired location and run the following command:
```
git clone https://github.com/utropicmedia/storj-influxdb.git
```

## Set-up Files
* Create a `db_property.json` file, with following contents about a IndluxDB instance:
	* hostName :- Host Name connect to InfluxDB
	* port :- Port connect to InfluxDB
	* username :- User Name of InfluxDB
	* password :- Password of InfluxDB
	* database :- InfluxDB Database Name
	* influxdExeutablePath:- Path to the influxd executable including the executble name
```
{
	"hostname": "influxdbHostName",
	"port":     "8086",
	"username": "username",
	"password": "password",
	"database": "influxDatabaseName"
	"influxdExeutablePath": "your/path/to/influxd-executable/including-executable-name"
}
```

* Create a `storj_config.json` file, with Storj network's configuration information in JSON format:
	* apiKey:- API key created in Storj satellite gui
	* satelliteURL:- Storj Satellite URL
	* encryptionPassphrase:- Storj Encryption Passphrase.
	* bucketName:- Name of the bucket to upload data into.
	* uploadPath:- Path on Storj Bucket to store data (optional) or "/"
	* serializedScope:- Serialized Scope Key shared while uploading data used to access bucket without API key
	* allowDownload:- Set true to create serialized scope key with restricted download
	* allowUpload:- Set true to create serialized scope key with restricted upload
	* allowList:- Set true to create serialized scope key with restricted list access
	* allowDelete:- Set true to create serialized scope key with restricted delete
	* notBefore:- Set time that is always before notAfter in the format YYYY-MM-DD_hh:mm:ss
	* notAfter:- Set time that is always after notBefore in the format YYYY-MM-DD_hh:mm:ss

```
{
	"apikey":     "change-me-to-the-api-key-created-in-satellite-gui",
	"satellite":  "us-central-1.tardigrade.io:7777",
	"bucket":     "change-me-to-desired-bucket-name",
	"uploadPath": "optionalpath/requiredfilename",
	"encryptionpassphrase": "you'll never guess this",
	"serializedScope": "change-me-to-the-api-key-created-in-encryption-access-apiKey",
	"allowDownload":	"true/false-to-allow-download",
	"allowUpload":	"true/false-to-allow-upload",
	"allowList":	"true/false-to-list-buckets",
	"allowDelete":	"true/false-to-allow-delete",
	"notBefore":	"0/time-set-always-before-notAfter",
	"notAfter":	"0/time-set-always-after-notBefore"
}
```

* Store both these files in a `config` folder. Sample files are already provided. Make respective changes and you are good to go. Filename command-line arguments are optional. Defualt locations are used.

## Description
* Overview
	* Storj-IndluxDB Connector connects to the user IndluxDB database using the credentials provided by them in the database configuration file, takes backup of the specified database and uploads the backup files on Storj network inside the path specified by the user in the Storj configutation file.
* User can choose from the following commands:
	* `Parse`:- This command works as follows:
		* Connect to the specidifed database using the user specified creadentions in the database configuration file(db_property).
		* Take backup of the database using a "backup" command provided by influxdb.
		* Read those backup files and store their contents inside a byte array and notify the user of successful read.
		* Print the size of the byte array if "debug" is provided as argument.
	* `Test`:- This command works as follows:
		* Connect to storj network.
		* Upload sample data to storj network inside the path specidifed by the user in the Storj configuration file(storj_config).
		* Download the sample data uploaded on Storj network on local disk if "debug" is provided as argument.
		* Generate a restricted serialized scope key if "restrict" is provided as argument.
	* `Store`:- This command works as follows:
		* Connect to the specidifed database using the user specified creadentions in the database configuration file(db_property).
		* Take backup of the database using a "backup" command provided by influxdb.
		* Connect to storj network.
		* Iterate through the backup files and upload each file in parts of buffer size 32KB to Storj network.
		* Download the backup files uploaded to Storj network same as the original backup files if "debug" is provided as argument.
		* Generate a restricted serialized scope key if restrict is provided as argument.

## Build Once
* To build the project, run the following command:
```
$ go build storj-influxdb.go
```
* or you cam simply run:
```
go build
```

**NOTE**: Please make sure Influxd server is running in order to connect to database successfully.

## Run the command-line tool
* Once you have built the project run the following commands as per your requirement:

* Get help
```
$ ./storj-influxdb -h
```

* Check version
```
$ ./storj-influxdb -v
```

* Create backup of desired InfluxDB instance and upload it to given Storj network bucket using serialized scope key. [note: Filename arguments are optional. Default locations are used.]
```
$ ./storj-influxdb store ./config/db_property.json ./config/storj_config.json
```

* Create backup of desired InfluxDB instance and upload it to given Storj network bucket API key and EncryptionPassPhrase from storj_config.json and creates an unrestricted shareable Serialized Scope Key. [note: Filename arguments are optional. Default locations are used.]
```
$ ./storj-influxdb store ./config/db_property.json ./config/storj_config.json  key
```

* Create backup of desired InfluxDB instance and upload it to given Storj network bucket using serialized scope key and creates a restricted shareable serialized scope key. [note: Filename arguments are optional. Default locations are used.]
```
$ ./storj-influxdb store ./config/db_property.json ./config/storj_config.json restrict
```

* Create backup of desired InfluxDB instance and upload it to given Storj network bucket API key and EncryptionPassPhrase from storj_config.json and creates a restricted shareable Serialized Scope Key. [note: Filename arguments are optional. Default locations are used.]
```
$ ./storj-influxdb store ./config/db_property.json ./config/storj_config.json  key restrict
```

* Create backup of desired InfluxDB instance and upload it to given Storj network bucket using serialized scope key in debug mode. [note: Filename arguments are optional. Default locations are used.]
```
$ ./storj-influxdb store debug ./config/db_property.json ./config/storj_config.json
```

* Read InfluxDB instance property from a desired JSON file and fetch its backup.
```
$ ./storj-influxdb parse   
```

* Read InfluxDB instance property in `debug` mode from a desired JSON file and fetch its backup.
```
$ ./storj-influxdb.go parse debug
```

* Read and parse Storj network's configuration, in JSON format, from a desired file and upload a sample object to given Storj network bucket using serialized scope key.
```
$ storj-influxdb test
```

* Read and parse Storj network's configuration, in JSON format, from a desired file and upload a sample object to given Storj network bucket using API key, encryption passphrase and satellite URL from storj_config.json and creates an unrestricted shareable serialized scope key.
```
$ storj-influxdb test key
```

* Read and parse Storj network's configuration, in JSON format, from a desired file and upload a sample object to given Storj network bucket API key, encryption passphrase and satellite URL from storj_config.json and creates a restricted shareable serialized scope key.
```
$ storj-influxdb test key restrict
```

* Read and parse Storj network's configuration, in JSON format, from a desired file and upload a sample object to given Storj network bucket using serialized scope key and creates a restricted shareable serialized scope key.
```
$ storj-influxdb test restrict
```

* Read and parse Storj network's configuration, in JSON format, from a desired file and upload a sample object to given Storj network bucket using serialized scope key in debug mode.
```
$ storj-influxdb test debug
```

##Testing
* The project has been tested on the following operating systems:
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
