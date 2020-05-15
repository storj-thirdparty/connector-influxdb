// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"log"
	"os"
	"github.com/urfave/cli"

	"github.com/utropicmedia/storj-influxdb/influx"
	"github.com/utropicmedia/storj-influxdb/storj"
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
	app.Authors = []cli.Author{{Name: "Shubham Shivam - Utropicmedia", Email: "development@utropicmedia.com"}}
	app.Version = "1.0.5"

}

// helper function to flag debug
func setDebug(debugVal bool) {
	gbDEBUG = debugVal
	influx.DEBUG = debugVal
	storj.DEBUG = debugVal
}

// setCommands sets various command-line options for the app.
func setCommands() {

	app.Commands = []cli.Command{
		{
			Name:    "parse",
			Aliases: []string{"p"},
			Usage:   "Command to read and parse JSON information about InfluxDB instance properties and then fetch ALL its tables. ",
			//\narguments-\n\t  fileName [optional] = provide full file name (with complete path), storing InfluxDB properties; if this fileName is not given, then data is read from ./config/db_connector.json\n\t  example = ./storj-influxdb p ./config/db_property.json\n",
			Action: func(cliContext *cli.Context) error
				var foundArgument = false
				var fullFileName = dbConfigFile
				// process arguments
				if len(cliContext.Args()) > 0 {
					for i := 0; i < len(cliContext.Args()); i++ {
						// Incase, debug is provided as argument.
						if cliContext.Args()[i] == "debug" {
							setDebug(true)
						} else {
							if !foundArgument{
								fullFileName = cliContext.Args()[i]
							} else {
								log.Fatal("Error: Unknown or too many arguments to run parse command.")
							}
						}
					}
				}

				// Establish connection with InfluxDB and get io.Reader implementor.
				dbReader, err := influx.ConnectToDB(fullFileName)
				if err != nil {
					fmt.Printf("Failed to establish connection with InfluxDB:")
					return err
				}
				// Connect to the Database and process data
				data, err := influx.FetchData(dbReader)
				if err != nil {
					fmt.Println("influx.FetchData:", err)
					return err
				}
				fmt.Println("Reading ALL tables from the InfluxDB database...Complete!")
				if gbDEBUG {
					fmt.Println("\nSize of fetched data from database: ", len(data), "bytes")
				}
				return nil
			},
		},
		{
			Name:    "test",
			Aliases: []string{"t"},
			Usage:   "Command to read and parse JSON information about Storj network and upload sample JSON data",
			//\n arguments- 1. fileName [optional] = provide full file name (with complete path), storing Storj configuration information if this fileName is not given, then data is read from ./config/storj_config.json example = ./storj-influxdb t ./config/storj_config.json\n\n\n",
			Action: func(cliContext *cli.Context) error {
				// Default Storj configuration file name.
				var fullFileName = storjConfigFile
				var foundFirstArgument = false
				var keyValue = ""
				var restrict = ""
				// process arguments
				if len(cliContext.Args()) > 0 {
					for i := 0; i < len(cliContext.Args()); i++ {
						// Incase, debug is provided as argument.
						if cliContext.Args()[i] == "debug" {
							setDebug(true)
						} else {
							if cliContext.Args()[i] == "key" {
									keyValue = cliContext.Args()[i]
							} else {
								if cliContext.Args()[i] == "restrict" {
									restrict = cliContext.Args()[i]
								} else {
									if !foundFirstArgument {
										fullFileName = cliContext.Args()[i]
										foundFirstArgument = true
									} else {
										log.Fatal("Error: Unknown or too many agruments to run test command.")
									}
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
				data := bytes.NewReader(DumpTSM)
				var fileNamesDEBUG []string
				// Connect to storj network.
				ctx, _, project, bucket, storjConfig, scope, errr := storj.ConnectStorjReadUploadData(fullFileName, keyValue, restrict)
				// Upload sample data on storj network.
				fileNamesDEBUG, errUpload := storj.ConnectUpload(ctx, bucket, project, data, fileName, fileNamesDEBUG, storjConfig, errr)
				if errUpload != nil {
					if restrict == "restrict" {
						fmt.Println("Restricted Serialized Scope Key: ", scope)
						fmt.Printf("\n")
						}
					fmt.Println("Error occured during sample data upload: ", errUpload)
					return errUpload
				}
				// Close storj project.
				err := project.Close()
				if err != nil {
					log.Fatalf("Could not close project error: %s", err)
				}
				fmt.Println("\nUpload \"testdata\" on Storj: Successful!")
				if keyValue == "key" && restrict == "" {
					fmt.Println("Serialized Scope Key: ", scope)
					fmt.Printf("\n")
				} else {
					if restrict == "restrict" {
						fmt.Println("Restricted Serialized Scope Key: ", scope)
						fmt.Printf("\n")
						}
					}
				return errr
			},
		},
		{
			Name:    "store",
			Aliases: []string{"s"},
			Usage:   "Command to connect and transfer ALL tables from a desired InfluxDB instance to given Storj Bucket as dumptsm",
			//\n    arguments-\n      1. fileName [optional] = provide full file name (with complete path), storing influxDB properties in JSON format\n   if this fileName is not given, then data is read from ./config/db_property.json\n      2. fileName [optional] = provide full file name (with complete path), storing Storj configuration in JSON format\n     if this fileName is not given, then data is read from ./config/storj_config.json\n   example = ./storj-influxdb s ./config/db_property.json ./config/storj_config.json\n",
			Action: func(cliContext *cli.Context) error {
				// Default configuration file names.
				var fullFileNameStorj = storjConfigFile
				var fullFileNameInfluxDB = dbConfigFile
				var keyValue = ""
				var restrict = ""
				var fileNamesDEBUG []string
				// process arguments - Reading fileName from the command line.
				var foundFirstArgument = false
				var foundSecondArgument = false
				if len(cliContext.Args()) > 0 {
					for i := 0; i < len(cliContext.Args()); i++ {
						// Incase debug is provided as argument.
						if cliContext.Args()[i] == "debug" {
							setDebug(true)
						} else {
							if cliContext.Args()[i] == "key" {
								keyValue = cliContext.Args()[i]
							} else {
								if cliContext.Args()[i] == "restrict" {
									restrict = cliContext.Args()[i]
								} else {
									if !foundFirstArgument {
										fullFileNameInfluxDB = cliContext.Args()[i]
										foundFirstArgument = true
									} else {
										if !foundSecondArgument {
											fullFileNameStorj = cliContext.Args()[i]
											foundSecondArgument = true
										} else {
												log.Fatal("Error: Unknown or too many agruments to run store command.")
										}
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
				var slash = ""
				if dbReader.File[0][0] == 'C' {
					slash = "\\"
				} else {
					slash = "/"
				  }
				ctx, _, project, bucket, storjConfig, scope, errr := storj.ConnectStorjReadUploadData(fullFileNameStorj, keyValue, restrict)
				// Fetch all tables' documents from InfluxDB instance as dumptsm
				// and simultaneously store them into desired Storj bucket.
				for i := 1; i <= len(dbReader.File) - 1; i++ {
					uploadFileName := strings.Split(dbReader.File[i], slash)
					fileNameStore := uploadFileName[len(uploadFileName) - 1]
					uploadFilePath := dbReader.ConfigInfluxDB.Database + "/" + fileNameStore
					fileNamesDEBUG,err= storj.ConnectUpload(ctx, bucket, project, dbReader, uploadFilePath, fileNamesDEBUG, storjConfig, errr)
					if err != nil {
						if restrict == "restrict" {
							fmt.Println("Restricted Serialized Scope Key: ", scope)
							fmt.Printf("\n")
							}
						fmt.Println("Error occured during upload: ", err)
						return err
					}
					os.Remove(uploadFilePath)
				}
				os.Remove(dbReader.ConfigInfluxDB.Database)
				if gbDEBUG {
					for i := 1; i <= len(dbReader.File)-1; i++ {
						uploadFileName := strings.Split(dbReader.File[i], slash)
						fileNameStore := uploadFileName[len(uploadFileName)-1]
						uploadFilePath := dbReader.ConfigInfluxDB.Database + "/"+ fileNameStore
						storj.Debug(bucket, project, storjConfig,uploadFilePath, dbReader.Copied)
					}
				}
				err = project.Close()
				if err != nil {
					log.Fatalf("Could not close project error: %s", err)
				}
				fmt.Println("\n")
				if keyValue == "key" && restrict == "" {
					fmt.Println("Serialized Scope Key: ", scope)
					fmt.Printf("\n")
				} else {
					if restrict == "restrict" {
						fmt.Println("Restricted Serialized Scope Key: ", scope)
						fmt.Printf("\n")
						}
					}
				return nil
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
