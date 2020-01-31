# Storj-influxdb Changelog

## [1.0.2] - 31-01-2020
### Changelog:
* Added keyword `influxdb_exeutable_path`  in db_property.json to access influxd executable path.
* Removed the hardcoded windows influxd executable path.
* Made changes in the DEBUG function to rectify the `index out of range` error.

## [1.0.1] - 26-12-2019
### Changelog:
* Changes made according to latest libuplink v0.27.1
* Changes made according to updated cli package.
* Added Macroon functionality.
* Added option to access storj using Serialized Scope Key. 
* Added keyword `key` to access storj using API key rather than Serialized Scope Key (defalt).
* Added keyword `restrict` to apply restrictions on API key and provide shareable Serialized Scope Key for users.
* Error handling for various events.
