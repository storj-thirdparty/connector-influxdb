package cmd

import (
	"fmt"
	"path"
	"path/filepath"

	"github.com/spf13/cobra"
)

// storeCmd represents the store command
var storeCmd = &cobra.Command{
	Use:   "store",
	Short: "Command to upload data to storjV3 network.",
	Long:  `Command to connect and transfer ALL tables from a desired InfluxDB instance to given Storj Bucket.`,
	Run:   influxStore,
}

func init() {
	rootCmd.AddCommand(storeCmd)
	var defaultInfluxFile string
	var defaultStorjFile string
	storeCmd.Flags().BoolP("accesskey", "a", false, "Connect to storj using access key(default connection method is by using API Key).")
	storeCmd.Flags().BoolP("share", "s", false, "For generating share access of the uploaded backup file.")
	storeCmd.Flags().BoolP("debug", "d", false, "For debugging purpose only.")
	storeCmd.Flags().StringVarP(&defaultInfluxFile, "influx", "i", "././config/db_property.json", "full filepath contaning Influxdb configuration.")
	storeCmd.Flags().StringVarP(&defaultStorjFile, "storj", "u", "././config/storj_config.json", "full filepath contaning storj V3 configuration.")
}

func influxStore(cmd *cobra.Command, args []string) {
	fmt.Println("store called")
	influxConfigfilePath, _ := cmd.Flags().GetString("influx")
	fmt.Println(influxConfigfilePath)
	fullFileNameStorj, _ := cmd.Flags().GetString("storj")
	fmt.Println(fullFileNameStorj)
	useAccessKey, _ := cmd.Flags().GetBool("accesskey")
	fmt.Println(useAccessKey)
	useAccessShare, _ := cmd.Flags().GetBool("share")
	fmt.Println(useAccessShare)
	useDebug, _ := cmd.Flags().GetBool("debug")
	fmt.Println(useDebug)

	// Read InfluxDB instance's properties from an external file.
	configInfluxDB := LoadInfluxProperty(influxConfigfilePath)

	// Establish connection with InfluxDB and get file names to be uploaded.
	filesToUpload := ConnectToDB(configInfluxDB)

	// Create storj configuration object.
	storjConfig := LoadStorjConfiguration(fullFileNameStorj)

	// Connect to storj network using the specified credentials.
	access, project := ConnectToStorj(fullFileNameStorj, storjConfig, useAccessKey)

	// Fetch all backup files from InfluxDB instance and simultaneously store them into desired Storj bucket.
	for i := 1; i <= len(filesToUpload)-1; i++ {
		fileName := filepath.Base(filesToUpload[i])
		uploadFileName := path.Join(configInfluxDB.Database, fileName)
		UploadData(project, storjConfig, uploadFileName, filesToUpload[i])
	}

	// Download the uploaded data if debug is provided as argument.
	if useDebug {
		for i := 1; i <= len(filesToUpload)-1; i++ {
			fileName := filepath.Base(filesToUpload[i])
			downloadFileName := path.Join(configInfluxDB.Database, fileName)
			DownloadData(project, storjConfig, downloadFileName)
		}
	}

	// Create restricted shareable serialized access if share is provided as argument.
	if useAccessShare {
		ShareAccess(access, storjConfig)
	}
}
