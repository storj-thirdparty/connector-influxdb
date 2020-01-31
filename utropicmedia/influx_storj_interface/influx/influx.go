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
    "time"
    "path/filepath"
    
    "github.com/influxdata/influxdb1-client/v2"
)

// DEBUG allows more detailed working to be exposed through the terminal.
var DEBUG = false

// ConfigInfluxDB stores the InfluxDB configuration parameters.
type ConfigInfluxDB struct {
	HostName   string `json:"hostname"`
	PortNumber string `json:"port"`
	UserName   string `json:"username"`
	Password   string `json:"password"`
	Database   string `json:"database"`
	Influxdb_exeutable_path string `json:"influxdb_exeutable_path"`
//	FileStorePath string `json:"dumpFilePath"`
}

// InfluxReader implements an io.Reader interface.
type InfluxReader struct {
	ConfigInfluxDB ConfigInfluxDB
	lastIndex      int
	File		   []string
	FileIndex 	   int
	Path		   string
	Copied		   int
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
	fmt.Println("Influxd \t",configInfluxDB.Influxdb_exeutable_path);
//	fmt.Println("Dump File Path \t",configInfluxDB.FileStorePath);
	return configInfluxDB, err
}

// ConnectToDB will connect to a InfluxDB instance,
// based on the read property from an external file.
// It returns a reference to an io.Reader with InfluxDB instance information
func ConnectToDB(fullFileName string) (*InfluxReader, error) { // fullFileName for fetching database credentials from given JSON filename.
	// Read InfluxDB instance's properties from an external file.
	configInfluxDB, err := loadInfluxProperty(fullFileName)

	HTTPAddr := fmt.Sprintf("http://%s:%s",configInfluxDB.HostName, configInfluxDB.PortNumber)
	//HTTPAddr := fmt.Sprintf("http://%s:%s/u=%s&p=%s",configInfluxDB.Hostname, configInfluxDB.Portnumber,configInfluxDB.Username,configInfluxDB.Password)
	//
	c, _ := client.NewHTTPClient(client.HTTPConfig{
	    Addr: HTTPAddr,
	    Username: configInfluxDB.UserName,
	    Password: configInfluxDB.Password,
	})

	_,_,err = c.Ping(0)
	if err != nil {
	    log.Fatal("Error creating InfluxDB Client: ", err.Error())
	}

	defer c.Close()

	fileStorePath := os.TempDir() + configInfluxDB.Database
	fmt.Println("FIleStorePath:", fileStorePath)
    // Create command to fetch dumptsm from the database.
	cmd := exec.Command("/usr/local/bin/influxd", "backup", "-portable", "-database", configInfluxDB.Database, "-host", configInfluxDB.HostName+":8088", fileStorePath)
	_, err  = cmd.Output()
	if err != nil {
		log.Fatal("Export failed to execute. Error was:", err.Error())
	}else{
		// Inform about successful connection.
	    fmt.Println("Successfully connected to InfluxDB!")
	}
    var files []string

	err = filepath.Walk(fileStorePath, visit(&files))
    if err != nil {
		fmt.Println("Error after backup:", err)
        panic(err)
	}
	return &InfluxReader{ConfigInfluxDB: configInfluxDB,File: files,Path: fileStorePath}, err
}

// Read reads and copies the dumptsm into the buffer.
func (influxReader *InfluxReader) Read(buf []byte) (int, error) { // buf represents the byte array, where data is
	if len(influxReader.File) == influxReader.FileIndex{
		return len(buf), io.EOF
	}
	if influxReader.File[influxReader.FileIndex] == influxReader.Path{
		influxReader.FileIndex++
	}

	// Read file which created by os package.
	bytes, err := ioutil.ReadFile(influxReader.File[influxReader.FileIndex])
        if err != nil {
            log.Fatal(err)
	}
	if influxReader.lastIndex < len(bytes) {
		if len(bytes[influxReader.lastIndex:]) < cap(buf){
			value:=len(bytes[influxReader.lastIndex:])
			influxReader.Copied = copy(buf[0:value], bytes[influxReader.lastIndex:])
			influxReader.lastIndex = influxReader.lastIndex + influxReader.Copied
			influxReader.lastIndex = 0
			influxReader.FileIndex++
			return len(buf), io.EOF
		}else{
			influxReader.Copied = copy(buf[:], bytes[influxReader.lastIndex:])
			influxReader.lastIndex = influxReader.lastIndex + influxReader.Copied
			return len(buf), io.ErrShortBuffer
		}

	}else{
		if len(influxReader.File) != influxReader.FileIndex{
		influxReader.lastIndex = 0
		influxReader.FileIndex++
		return len(buf), io.EOF
		}
	}
	return len(buf), io.EOF
}

// FetchData reads ALL tables' data, and
// returns them in appended format.
func FetchData(databaseReader io.Reader) ([]byte, error) { // databaseReader is an io.Reader implementation that 'reads' desired data.
	// Create a buffer of feasible size.
	rawDocumentBSON := make([]byte, 0, 32768)

	// Retrieve ALL tables in the database.
	var allCollectionsDataBSON = []byte{}

	var numOfBytesRead int
	var err error

	// Read data using the given io.Reader.
	for err = io.ErrShortBuffer; err == io.ErrShortBuffer; {
		numOfBytesRead, err = databaseReader.Read(rawDocumentBSON)
		//
		if numOfBytesRead > 0 {
			// Append the read data to earlier one.
			allCollectionsDataBSON = append(allCollectionsDataBSON[:], rawDocumentBSON...)
			//
			if DEBUG {
				fmt.Printf("Read %d bytes of data - Error: %s == %s => %t\n", numOfBytesRead, err, io.ErrShortBuffer, err == io.ErrShortBuffer)
			}
		}
	}
	//
	if DEBUG {
		// complete read data from ALL tables.
		t := time.Now()
		time := t.Format("2006-01-02_15:04:05")
		var filename = "dumptsm_" + time + ".sql"
		err = ioutil.WriteFile(filename, allCollectionsDataBSON, 0644)
	}

	fmt.Println("FetchData", err)
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