package cmd

import (
	"fmt"
	"path"
	"path/filepath"

	"github.com/spf13/cobra"
)

// storeCmd represents the store command.
var storeCmd = &cobra.Command{
	Use:   "store",
	Short: "Command to upload data to storjV3 network.",
	Long:  `Command to connect and transfer ALL tables from a desired InfluxDB instance to given Storj Bucket.`,
	Run:   influxStore,
}

func init() {

	// Setup the store command with its flags.
	rootCmd.AddCommand(storeCmd)
	var defaultInfluxFile string
	var defaultStorjFile string
	storeCmd.Flags().BoolP("accesskey", "a", false, "Connect to storj using access key(default connection method is by using API Key).")
	storeCmd.Flags().BoolP("share", "s", false, "For generating share access of the uploaded backup file.")
	storeCmd.Flags().StringVarP(&defaultInfluxFile, "influx", "i", "././config/db_property.json", "full filepath contaning Influxdb configuration.")
	storeCmd.Flags().StringVarP(&defaultStorjFile, "storj", "u", "././config/storj_config.json", "full filepath contaning Storj V3 configuration.")
}

func influxStore(cmd *cobra.Command, args []string) {

	// Process arguments from the CLI.
	influxConfigfilePath, _ := cmd.Flags().GetString("influx")
	fullFileNameStorj, _ := cmd.Flags().GetString("storj")
	useAccessKey, _ := cmd.Flags().GetBool("accesskey")
	useAccessShare, _ := cmd.Flags().GetBool("share")

	// Read InfluxDB instance's configurations from an external file and create an InfluxDB configuration object.
	configInfluxDB := LoadInfluxProperty(influxConfigfilePath)

	// Read storj network configurations from and external file and create a storj configuration object.
	storjConfig := LoadStorjConfiguration(fullFileNameStorj)

	// Connect to storj network using the specified credentials.
	access, project := ConnectToStorj(fullFileNameStorj, storjConfig, useAccessKey)

	// Create back-up of InfluxDB database and get file names to be uploaded.
	filesToUpload := CreateBackup(configInfluxDB)

	fmt.Printf("\nInitiating back-up.\n")
	// Fetch all backup files from InfluxDB instance and simultaneously store them into desired Storj bucket.
	for i := 1; i <= len(filesToUpload)-1; i++ {
		fileName := filepath.Base(filesToUpload[i])
		uploadFileName := path.Join(configInfluxDB.Database, fileName)
		UploadData(project, storjConfig, uploadFileName, filesToUpload[i])
	}
	fmt.Printf("\nBack-up complete.\n\n")

	// Create restricted shareable serialized access if share is provided as argument.
	if useAccessShare {
		ShareAccess(access, storjConfig)
	}
}
