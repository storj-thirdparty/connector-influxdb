// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package storj

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"storj.io/uplink"
)

// DEBUG allows more detailed working to be exposed through the terminal.
var DEBUG = false

// ConfigStorj depicts keys to search for within the stroj_config.json file.
type ConfigStorj struct {
	APIKey					string `json:"apikey"`
	Satellite				string `json:"satellite"`
	Bucket					string `json:"bucket"`
	UploadPath				string `json:"uploadPath"`
	EncryptionPassphrase	string `json:"encryptionpassphrase"`
	SerializedScope			string `json:"serializedScope"`
	AllowDownload			string `json:"allowDownload"`
	AllowUpload				string `json:"allowUpload"`
	AllowList				string `json:"allowList"`
	AllowDelete				string `json:"allowDelete"`
	NotBefore				string `json:"notBefore"`
	NotAfter				string `json:"notAfter"`
}

// LoadStorjConfiguration reads and parses the JSON file that contain Storj configuration information.
func LoadStorjConfiguration(fullFileName string) (ConfigStorj, error) { // fullFileName for fetching storj V3 credentials from  given JSON filename.
	var configStorj ConfigStorj
	fileHandle, err := os.Open(fullFileName)
	if err != nil {
		return configStorj, err
	}
	defer fileHandle.Close()

	jsonParser := json.NewDecoder(fileHandle)
	jsonParser.Decode(&configStorj)
	// Display storj configuration read from file.
	fmt.Println("\nRead Storj configuration from the ", fullFileName, " file")
	fmt.Println("\nAPI Key\t\t: ", configStorj.APIKey)
	fmt.Println("Satellite	: ", configStorj.Satellite)
	fmt.Println("Bucket		: ", configStorj.Bucket)
	fmt.Println("Upload Path\t: ", configStorj.UploadPath)
	fmt.Println("Serialized Scope Key\t: ", configStorj.SerializedScope, "\n")
	return configStorj, nil
}

// ConnectStorjReadUploadData reads Storj configuration from given file,
// connects to the desired Storj network.
/// It then reads data property from an external file.
func ConnectStorjReadUploadData(fullFileName string, keyValue string, restrict string) (context.Context, *uplink.Access, *uplink.Project, *uplink.Bucket, ConfigStorj, string, error) { // fullFileName for fetching storj V3 credentials from  given JSON filename
	// Read Storj bucket's configuration from an external file.
	configStorj, err := LoadStorjConfiguration(fullFileName)
	if err != nil {
		log.Fatal("Could not load storj configurations: %v", err)
	}
	var scope string
	var access *uplink.Access
	var cfg uplink.Config
	// Configure the UserAgent
	cfg.UserAgent = "InfluxDB"
	ctx := context.Background()

	if keyValue == "key" {
		access, err = cfg.RequestAccessWithPassphrase(ctx, configStorj.Satellite, configStorj.APIKey, configStorj.EncryptionPassphrase) //Connect to storj using the specified configuration to retieve access
		if err != nil {
			log.Fatal("Could not request access grant: ", err)
		}
		scope,err = access.Serialize() // Serialize the access
		if err != nil {
			log.Fatal("Could not serialize access: ", err)
		}
	} else {
		if keyValue == ""{
			//Parse the serialized access
			access, err = uplink.ParseAccess(configStorj.SerializedScope) // Connect to storj using the serialized scope
			if err != nil {
				log.Fatal("Could not parse access: ", err)
			}
			scope,err = access.Serialize() //Serialize the parsed access
			if err != nil {
				log.Fatal("Could not serialize access: ", err)
			}
		}
	}
	if restrict == "restrict" {
		allowDownload, _ := strconv.ParseBool(configStorj.AllowDownload)
		allowUpload, _ := strconv.ParseBool(configStorj.AllowUpload)
		allowList, _ := strconv.ParseBool(configStorj.AllowList)
		allowDelete, _ := strconv.ParseBool(configStorj.AllowDelete)
		notBefore, _ := time.Parse("2006-01-02_15:04:05", configStorj.NotBefore)
		notAfter, _ := time.Parse("2006-01-02_15:04:05", configStorj.NotAfter)
		permission := uplink.Permission{
		AllowDownload:	allowDownload,
		AllowUpload:  allowUpload,
		AllowList:	allowList,
		AllowDelete:	allowDelete,
		NotBefore:	notBefore,
		NotAfter:	notAfter,
		}
		sharedAccess, err := access.Share(permission)
		if err != nil {
			log.Fatal("Could not generate shared access: ", err)
		}
		scope,err = sharedAccess.Serialize() // Generate restricted serialized scope
		if err != nil {
			log.Fatal("Could not serialize shared access: ", err)
		}
	}
	project, err := cfg.OpenProject(ctx, access) // Open a new porject
	if err != nil {
		log.Fatal("Could not open project:", err)
	}
	fmt.Println("Ensuring Bucket: ", configStorj.Bucket)
	// Ensure the desired Bucket within the Project is created.
	bucket, err := project.EnsureBucket(ctx, configStorj.Bucket)
	if err != nil {
		log.Fatal(err)
	}
	err = project.Close()
	if err != nil {
		log.Fatalf("Could not close project error: %s", err)
	}
	return ctx, access, project, bucket, configStorj, scope, err
}

// ConnectUpload uploads the data to storj network.
func ConnectUpload(ctx context.Context, bucket *uplink.Bucket, project *uplink.Project, databaseReader io.Reader, databaseName string, fileNamesDEBUG []string, configStorj ConfigStorj, err error) ([]string, error) {
	// databaseReader is an io.Reader implementation that 'reads' desired data,
	// which is to be uploaded to storj network.
	// databaseName for adding dataBase name in storj filename.
	// Read data using bytes and upload it to Storj.
	var file []string
	file = fileNamesDEBUG
	var filename string
	i := 0
	for err = io.ErrShortBuffer; err == io.ErrShortBuffer; {
		slitString := strings.Split(databaseName, ".")
		if len(slitString) > 2 {
			filename = slitString[0] + slitString[1] + slitString[2] + "/" + strconv.Itoa(i) + "." + slitString[1] + "." + slitString[2] + "." + slitString[3]
		} else {
			filename = slitString[0] + slitString[1] + "/" + strconv.Itoa(i) + "." + slitString[1]
		}
		checkSlash := configStorj.UploadPath[len(configStorj.UploadPath)-1:]
		if checkSlash != "/" {
			configStorj.UploadPath = configStorj.UploadPath + "/"
		}
		fmt.Println("\nUpload Object Path: ", configStorj.UploadPath + filename)
		upload, err1 := project.UploadObject(ctx, bucket.Name, configStorj.UploadPath + filename, nil)
		if err1 != nil {
			fmt.Println("Could not upload data: ", err)
			return nil, err1
		}
		fmt.Println("\nUploading of the object to the Storj bucket: Initiated...")
		// Copy the data to storj object through the database reader
		_, err = io.Copy(upload, databaseReader)
		// Commit the upload
		err2 := upload.Commit()
			if err2 != nil {
				fmt.Println("could not commit uploaded object...")
				return nil, err2
			}
		i = i + 1
	}
	if err != nil {
		fmt.Printf("Could not upload: %s", err)
		return nil, err
	}
	fmt.Println("Uploading object to Storj bucket: Completed!")
	return file, nil
}

// Debug function downloads the data from storj bucket after upload to verify data is uploaded successfully.
func Debug(bucket *uplink.Bucket, project *uplink.Project, configStorj ConfigStorj, lastFileName string, copied int) {
	checkSlash := configStorj.UploadPath[len(configStorj.UploadPath)-1:]
	if checkSlash != "/" {
		configStorj.UploadPath = configStorj.UploadPath + "/"
	}
	ctx := context.Background()
	var cfg uplink.ListObjectsOptions
	slitString := strings.Split(lastFileName, "/")
	slitFileString := strings.Split(slitString[1], ".")
	var lastextension string
	// Configure the Prefix
	if len(slitFileString) > 2 {
		cfg.Prefix = configStorj.UploadPath + slitString[0] + "/" + slitFileString[0] + slitFileString[1] + slitFileString[2] + "/"
		lastextension = slitFileString[1] + "." + slitFileString[2] + "." + slitFileString[3]
	} else {
		cfg.Prefix = configStorj.UploadPath + slitString[0] + "/" + slitFileString[0] + slitFileString[1] + "/"
		lastextension = slitFileString[1]
	}
	list := project.ListObjects(ctx, bucket.Name, &cfg)
	var downloadFileDisk *os.File
	i := 0
	for list.Next() {
		// Retrieve contents of the object uploaded on storj
		receivedContents, err := downloadObject(ctx, bucket, project, cfg.Prefix+strconv.Itoa(i) + "." + lastextension )
		if err != nil {
			log.Fatalf("could not read object:", err)
		}
		if _, err := os.Stat("./debug"); os.IsNotExist(err) {
			err1 := os.Mkdir("./debug", 0755)
			if err1 != nil {
				log.Fatal("Invalid Download Path: ", err1)
			}
		}
		if _, err := os.Stat("./debug" + "/" + slitString[0]); os.IsNotExist(err) {
			err1 := os.Mkdir("./debug"+"/"+slitString[0], 0755)
			if err1 != nil {
				log.Fatal("Invalid Download Path: ", err1)
			}
		}
		// Create/open file in append mode
		downloadFileDisk, err = os.OpenFile("./debug"+"/"+slitString[0] + "/" + slitString[1], os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
		if err != nil {
			log.Fatal(err)
		}
		// Write the contents retrieved from downloaded object to file on local disk
		_, err = downloadFileDisk.Write(receivedContents)
		if err != nil {
			log.Fatal(err)
		}
		defer downloadFileDisk.Close()
		downloadFileDisk.Close()
		fmt.Printf("\nDownloaded %d bytes of Object from bucket!\n", len(receivedContents))
		i = i + 1
	}
	fmt.Printf("Debug file \"%s\" downloaded to \"%s\"\n", lastFileName, "debug/")
}

// downloadObject reads the the object from storj bucket and
// returns the byte array of the read content.
func downloadObject(ctx context.Context, bucket *uplink.Bucket, project *uplink.Project, filename string) ([]byte, error) {
	download, err := project.DownloadObject(ctx, bucket.Name, filename, nil)
	if err != nil {
		return nil, fmt.Errorf("Could not download object at %v", err)
	}
	defer download.Close()
	// Read everything from the stream.
	receivedContents, err := ioutil.ReadAll(download)
	if err != nil {
		return nil, fmt.Errorf("Could not read object: %v", err)
	}
	return receivedContents, err
}
