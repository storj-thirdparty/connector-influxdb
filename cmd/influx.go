// Module to connect to a Influx DB instance
// and fetch its database backup files.

package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// ConfigInfluxDB stores the InfluxDB configuration parameters.
type ConfigInfluxDB struct {
	HostName             string `json:"hostname"`
	PortNumber           string `json:"port"`
	UserName             string `json:"username"`
	Password             string `json:"password"`
	Database             string `json:"database"`
	InfluxdExeutablePath string `json:"influxdExeutablePath"`
}

// LoadInfluxProperty reads and parses the configuration JSON file
// that contains an InfluxDB instance's credentials
// and returns all the properties embedded in a configuration object.
func LoadInfluxProperty(fullFileName string) ConfigInfluxDB {

	var configInfluxDB ConfigInfluxDB

	// Open the file and generate file handle.
	fileHandle, err := os.Open(filepath.Clean(fullFileName))
	if err != nil {
		log.Fatal("Could not load influx cofig file: ", err)
	}

	// Decode and parse the JSON properties.
	jsonParser := json.NewDecoder(fileHandle)
	if err = jsonParser.Decode(&configInfluxDB); err != nil {
		log.Fatal(err)
	}

	// Close the file handle after reading from it.
	if err = fileHandle.Close(); err != nil {
		log.Fatal(err)
	}

	// Display the read InfluxDB configuration properties.
	fmt.Println("\nRead InfluxDB configuration from the ", fullFileName, " file")
	fmt.Println("HostName\t", configInfluxDB.HostName)
	fmt.Println("PortNumber\t", configInfluxDB.PortNumber)
	fmt.Println("UserName \t", configInfluxDB.UserName)
	fmt.Println("Password \t", configInfluxDB.Password)
	fmt.Println("Database \t", configInfluxDB.Database)
	fmt.Println("InfluxdExecutablePath \t", configInfluxDB.InfluxdExeutablePath)

	return configInfluxDB
}

// CreateBackup creates the backup of the specified database in the configuration file
// and stores the backup file names in a slice inside the reader object.
// It returns a slice of string containing back-up file names.
func CreateBackup(configInfluxDB ConfigInfluxDB) []string {

	// Path to a temporary directory to store the backup files into.
	backupPath := filepath.Join(os.TempDir(), configInfluxDB.Database)
	if _, err := os.Stat(backupPath); !os.IsNotExist(err) {
		if err := os.RemoveAll(backupPath); err != nil {
			log.Fatal(err)
		}
	}

	// Create command to create backup files from the database in the temportary directory if the backup file don't exists beforehand.
	cmd := exec.Command(configInfluxDB.InfluxdExeutablePath, "backup", "-portable", "-database", configInfluxDB.Database, "-host", configInfluxDB.HostName+":8088", backupPath)
	_, err := cmd.Output()
	if err != nil {
		log.Fatalf("Export failed to execute. Error was: %s", err)
	}
	fmt.Println("\nBackup created successfully")

	// Store the back-up file names in a slice.
	var files []string
	err = filepath.Walk(backupPath, visit(&files))
	if err != nil {
		fmt.Println("Could not store file names:", err)
		log.Fatalf("Export failed to execute. Error was: %s", err)
	}

	return files
}

func visit(files *[]string) filepath.WalkFunc {

	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
		}

		*files = append(*files, path)
		return nil
	}
}
