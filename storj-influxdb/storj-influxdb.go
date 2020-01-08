// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import (
	//"bytes"

	"bytes"
	"fmt"
	"io/ioutil"
	"strings"

	//"io/ioutil"
	"log"
	"os"

	//"time"
	"unsafe"

	"utropicmedia/influx_storj_interface/influx"
	"utropicmedia/influx_storj_interface/storj"

	"github.com/urfave/cli"
)

const dbConfigFile = "./config/db_property.json"
const storjConfigFile = "./config/storj_config.json"

var gbDEBUG = false

// Create command-line tool to read from CLI.
var app = cli.NewApp()

// SetAppInfo sets information about the command-line application.
func setAppInfo() {
	app.Name = "Storj InfluxDB Connector"
	app.Usage = "Backup your InfluxDB tables to the decentralized Storj network"
	app.Authors = []*cli.Author{{Name: "Shubham Shivam - Utropicmedia", Email: "development@utropicmedia.com"}}
	app.Version = "1.0.1"

}

// helper function to flag debug
func setDebug(debugVal bool) {
	gbDEBUG = debugVal
	influx.DEBUG = debugVal
	storj.DEBUG = debugVal
}

// setCommands sets various command-line options for the app.
func setCommands() {

	app.Commands = []*cli.Command{
		{
			Name:    "parse",
			Aliases: []string{"p"},
			Usage:   "Command to read and parse JSON information about InfluxDB instance properties and then fetch ALL its tables. ",
			//\narguments-\n\t  fileName [optional] = provide full file name (with complete path), storing InfluxDB properties; if this fileName is not given, then data is read from ./config/db_connector.json\n\t  example = ./storj-influxdb d ./config/db_property.json\n",
			Action: func(cliContext *cli.Context) error {
				var fullFileName = dbConfigFile

				// process arguments
				if len(cliContext.Args().Slice()) > 0 {
					for i := 0; i < len(cliContext.Args().Slice()); i++ {

						// Incase, debug is provided as argument.
						if cliContext.Args().Slice()[i] == "debug" {
							setDebug(true)
						} else {
							fullFileName = cliContext.Args().Slice()[i]
						}
					}
				}

				// Establish connection with InfluxDB and get io.Reader implementor.
				dbReader, err := influx.ConnectToDB(fullFileName)
				//
				if err != nil {
					fmt.Printf("Failed to establish connection with InfluxDB:")
					return err
				} else {
					// Connect to the Database and process data
					data, err := influx.FetchData(dbReader)

					if err != nil {
						fmt.Printf("influx.FetchData:")
						return err
					} else {
						fmt.Println("Reading ALL tables from the InfluxDB database...Complete!")
					}

					if gbDEBUG {
						fmt.Println("Size of fetched data from database: ", dbReader.ConfigInfluxDB.Database, unsafe.Sizeof(data))
					}
				}
				return err
			},
		},
		{
			Name:    "test",
			Aliases: []string{"t"},
			Usage:   "Command to read and parse JSON information about Storj network and upload sample JSON data",
			//\n arguments- 1. fileName [optional] = provide full file name (with complete path), storing Storj configuration information if this fileName is not given, then data is read from ./config/storj_config.json example = ./storj-influxdb s ./config/storj_config.json\n\n\n",
			Action: func(cliContext *cli.Context) error {

				// Default Storj configuration file name.
				var fullFileName = storjConfigFile
				var foundFirstFileName = false
				var foundSecondFileName = false
				var keyValue string
				var restrict string

				// process arguments
				if len(cliContext.Args().Slice()) > 0 {
					for i := 0; i < len(cliContext.Args().Slice()); i++ {

						// Incase, debug is provided as argument.
						if cliContext.Args().Slice()[i] == "debug" {
							setDebug(true)
						} else {
							if !foundFirstFileName {
								fullFileName = cliContext.Args().Slice()[i]
								foundFirstFileName = true
							} else {
								if !foundSecondFileName {
									keyValue = cliContext.Args().Slice()[i]
									foundSecondFileName = true
								} else {
									restrict = cliContext.Args().Slice()[i]
								}
							}
						}
					}
				}

				// Sample database name and data to be uploaded
				dbName := "testdb"
				DumpTSM := []byte("DROP TABLE IF EXISTS `HelloStorj`")
				var fileName string
				if gbDEBUG {
					fileName = "./dumptsm_.txt"
					data := []byte(DumpTSM)
					err := ioutil.WriteFile(fileName, data, 0644)
					if err != nil {
						fmt.Println("Error while writting to file ")
					}
				}
				fileName = dbName + ".txt"
				data := bytes.NewReader(DumpTSM)
				var fileNamesDEBUG []string
				// Connect to storj network.
				ctx, uplink, project, bucket, storjConfig, _, errr := storj.ConnectStorjReadUploadData(fullFileName, keyValue, restrict)

				// Upload sample data on storj network.
				fileNamesDEBUG = storj.ConnectUpload(ctx, bucket, data, fileName, fileNamesDEBUG, storjConfig, errr)

				// Close storj project.
				storj.CloseProject(uplink, project, bucket)
				//
				fmt.Println("\nUpload \"testdata\" on Storj: Successful!")
				return errr
			},
		},
		{
			Name:    "store",
			Aliases: []string{"s"},
			Usage:   "Command to connect and transfer ALL tables from a desired InfluxDB instance to given Storj Bucket as dumptsm",
			//\n    arguments-\n      1. fileName [optional] = provide full file name (with complete path), storing influxDB properties in JSON format\n   if this fileName is not given, then data is read from ./config/db_property.json\n      2. fileName [optional] = provide full file name (with complete path), storing Storj configuration in JSON format\n     if this fileName is not given, then data is read from ./config/storj_config.json\n   example = ./storj-influxdb c ./config/db_property.json ./config/storj_config.json\n",
			Action: func(cliContext *cli.Context) error {

				// Default configuration file names.
				var fullFileNameStorj = storjConfigFile
				var fullFileNameInfluxDB = dbConfigFile
				var keyValue string
				var restrict string
				var fileNamesDEBUG []string

				// process arguments - Reading fileName from the command line.
				var foundFirstFileName = false
				var foundSecondFileName = false
				var foundThirdFileName = false

				if len(cliContext.Args().Slice()) > 0 {
					for i := 0; i < len(cliContext.Args().Slice()); i++ {
						// Incase debug is provided as argument.
						if cliContext.Args().Slice()[i] == "debug" {
							setDebug(true)
						} else {
							if !foundFirstFileName {
								fullFileNameInfluxDB = cliContext.Args().Slice()[i]
								foundFirstFileName = true
							} else {
								if !foundSecondFileName {
									fullFileNameStorj = cliContext.Args().Slice()[i]
									foundSecondFileName = true
								} else {
									if !foundThirdFileName {
										keyValue = cliContext.Args().Slice()[i]
										foundThirdFileName = true
									} else {
										restrict = cliContext.Args().Slice()[i]
									}
								}
							}
						}
					}
				}

				// Establish connection with InfluxDB and get io.Reader implementor.
				dbReader, err := influx.ConnectToDB(fullFileNameInfluxDB)

				if err != nil {
					fmt.Printf("Failed to establish connection with InfluxDB:\n")
					return err
				}
				ctx, uplink, proj, bucket, storjConfig, scope, errr := storj.ConnectStorjReadUploadData(fullFileNameStorj, keyValue, restrict)
				// Fetch all tables' documents from InfluxDB instance as dumptsm
				// and simultaneously store them into desired Storj bucket.
				for i := 1; i <= len(dbReader.File)-1; i++ {
					uploadFileName := strings.Split(dbReader.File[i], "\\")
					fileNameStore := uploadFileName[len(uploadFileName)-1]
					uploadFilePath := dbReader.ConfigInfluxDB.Database + "/" + fileNameStore
					fileNamesDEBUG = storj.ConnectUpload(ctx, bucket, dbReader, uploadFilePath, fileNamesDEBUG, storjConfig, errr)
					os.Remove(uploadFilePath)
				}
				os.Remove(dbReader.ConfigInfluxDB.Database)
				for i := 1; i <= len(dbReader.File)-1; i++ {
					uploadFileName := strings.Split(dbReader.File[i], "\\")
					fileNameStore := uploadFileName[len(uploadFileName)-1]
					uploadFilePath := dbReader.ConfigInfluxDB.Database + "/" + fileNameStore
					storj.Debug(bucket, storjConfig, uploadFilePath, dbReader.Copied)
				}
				storj.CloseProject(uplink, proj, bucket)

				fmt.Println(" ")
				if keyValue == "key" {
					if restrict == "restrict" {
						fmt.Println("Restricted Serialized Scope Key: ", scope)
						fmt.Println(" ")
					} else {
						fmt.Println("Serialized Scope Key: ", scope)
						fmt.Println(" ")
					}
				}
				return err
			},
		},
	}
}

func main() {

	setAppInfo()
	setCommands()

	setDebug(false)

	err := app.Run(os.Args)

	if err != nil {
		log.Fatalf("app.Run: %s", err)
	}
}
