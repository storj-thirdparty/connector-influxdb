// Module to connect to a Influx DB instance
// and fetch its dumptsm.txt.

package influx

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	client "github.com/influxdata/influxdb1-client/v2"
)

// DEBUG allows more detailed working to be exposed through the terminal.
var DEBUG = false

// ConfigInfluxDB stores the InfluxDB configuration parameters.
type ConfigInfluxDB struct {
	HostName             string `json:"hostname"`
	PortNumber           string `json:"port"`
	UserName             string `json:"username"`
	Password             string `json:"password"`
	Database             string `json:"database"`
	InfluxdExeutablePath string `json:"influxdExeutablePath"`
}

// InfluxReader implements an io.Reader interface.
type InfluxReader struct {
	ConfigInfluxDB ConfigInfluxDB
	lastIndex      int
	File           []string
	FileIndex      int
	Path           string
	Copied         int
}

// loadInfluxProperty reads and parses the JSON file
// that contain a InfluxDB instance's property
// and returns all the properties as an object.
func loadInfluxProperty(fullFileName string) (ConfigInfluxDB, error) { // fullFileName for fetching database credentials from  given JSON filename.
	var configInfluxDB ConfigInfluxDB
	// Open and read the file.
	fileHandle, err := os.Open(fullFileName)
	if err != nil {
		return configInfluxDB, err
	}
	defer fileHandle.Close()

	// Decode and parse the JSON properties.
	jsonParser := json.NewDecoder(fileHandle)
	jsonParser.Decode(&configInfluxDB)
	// Display the read InfluxDB configuration properties.
	fmt.Println("Read InfluxDB configuration from the ", fullFileName, " file")
	fmt.Println("HostName\t", configInfluxDB.HostName)
	fmt.Println("PortNumber\t", configInfluxDB.PortNumber)
	fmt.Println("UserName \t", configInfluxDB.UserName)
	fmt.Println("Password \t", configInfluxDB.Password)
	fmt.Println("Database \t", configInfluxDB.Database)
	fmt.Println("InfluxdExecutablePath \t", configInfluxDB.InfluxdExeutablePath)
	return configInfluxDB, err
}

// ConnectToDB will connect to a InfluxDB instance,
// based on the read property from an external file.
// It returns a reference to an io.Reader with InfluxDB instance information
func ConnectToDB(fullFileName string) (*InfluxReader, error) { // fullFileName for fetching database credentials from given JSON filename.
	// Read InfluxDB instance's properties from an external file.
	configInfluxDB, err := loadInfluxProperty(fullFileName)

	// TODO:  What is this used for ??
	HTTPAddr := fmt.Sprintf("http://%s:%s", configInfluxDB.HostName, configInfluxDB.PortNumber)
	c, _ := client.NewHTTPClient(client.HTTPConfig{
		Addr:     HTTPAddr,
		Username: configInfluxDB.UserName,
		Password: configInfluxDB.Password,
	})
	_, _, err = c.Ping(0)
	if err != nil {
		log.Fatal("Error creating InfluxDB Client: ", err.Error())
	}
	defer c.Close()

	var slash = ""
	if string(os.TempDir()[0]) == "C" {
		slash = "\\"
	} else {
		slash = "/"
	}
	fileStorePath := os.TempDir() + slash + configInfluxDB.Database
	if _, err := os.Stat(fileStorePath); os.IsNotExist(err) {
		// Create command to fetch dumptsm from the database.
		cmd := exec.Command(configInfluxDB.InfluxdExeutablePath, "backup", "-portable", "-database", configInfluxDB.Database, "-host", configInfluxDB.HostName+":8088", fileStorePath)
		_, err = cmd.Output()
		if err != nil {
			log.Fatal("Export failed to execute. Error was:", err.Error())
		}
		fmt.Println("Successfully connected to InfluxDB!")
	}
	var files []string
	err = filepath.Walk(fileStorePath, visit(&files))
	if err != nil {
		fmt.Println("Could not store file names:", err)
		return nil, err
	}
	return &InfluxReader{ConfigInfluxDB: configInfluxDB, File: files, Path: fileStorePath}, err
}

// Read reads and copies the dumptsm into the buffer.
func (influxReader *InfluxReader) Read(buf []byte) (int, error) { // buf represents the byte array, where data is
	if len(influxReader.File) == influxReader.FileIndex {
		return len(buf), nil
	}
	if influxReader.File[influxReader.FileIndex] == influxReader.Path {
		influxReader.FileIndex++
	}
	fmt.Println("Reading the backup file: ", influxReader.File[influxReader.FileIndex])
	// Read file which created by os package.
	bytes, err := ioutil.ReadFile(influxReader.File[influxReader.FileIndex])
	if err != nil {
		log.Fatal(err)
	}
	if influxReader.lastIndex < len(bytes) {
		if len(bytes[influxReader.lastIndex:]) < cap(buf) {
			value := len(bytes[influxReader.lastIndex:])
			influxReader.Copied = copy(buf[0:value], bytes[influxReader.lastIndex:])
			influxReader.lastIndex = influxReader.lastIndex + influxReader.Copied
			influxReader.lastIndex = 0
			influxReader.FileIndex++
			return value, io.EOF
		}
		influxReader.Copied = copy(buf[:], bytes[influxReader.lastIndex:])
		influxReader.lastIndex = influxReader.lastIndex + influxReader.Copied
		return len(buf), io.ErrShortBuffer
	}
	if len(influxReader.File) != influxReader.FileIndex {
		influxReader.lastIndex = 0
		influxReader.FileIndex++
		return 0, io.ErrShortBuffer
	}
	return len(buf), io.EOF
}

// FetchData reads ALL tables' data, and
// returns them in appended format.
func FetchData(databaseReader io.Reader) ([]byte, error) { // databaseReader is an io.Reader implementation that 'reads' desired data.
	// Create a buffer of feasible size.
	rawDocumentBSON := make([]byte, 32768)
	// Retrieve ALL tables in the database.
	var allCollectionsDataBSON = []byte{}
	var numOfBytesRead int
	var err error
	fmt.Println("Reading from the database...Initiated.")
	// Read data using the given io.Reader.
	for err = io.ErrShortBuffer; err != nil; {
		numOfBytesRead, err = databaseReader.Read(rawDocumentBSON)
		if numOfBytesRead > 0 {
			// Append the read data to earlier one.
			tempCollectionBson := make([]byte, numOfBytesRead)
			copy(tempCollectionBson, rawDocumentBSON[0:numOfBytesRead])
			allCollectionsDataBSON = append(allCollectionsDataBSON[:], tempCollectionBson...)
		}
	}
	return allCollectionsDataBSON, err
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
