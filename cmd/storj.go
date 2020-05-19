// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"storj.io/uplink"
)

// DEBUG allows more detailed working to be exposed through the terminal.
var DEBUG = false

// MAXRETRY defines number of times to try upload data to storj before throwing error
var MAXRETRY = 5

// ConfigStorj depicts keys to search for within the stroj_config.json file.
type ConfigStorj struct {
	APIKey               string `json:"apikey"`
	Satellite            string `json:"satellite"`
	Bucket               string `json:"bucket"`
	UploadPath           string `json:"uploadPath"`
	EncryptionPassphrase string `json:"encryptionpassphrase"`
	SerializedAccess     string `json:"serializedAccess"`
	AllowDownload        string `json:"allowDownload"`
	AllowUpload          string `json:"allowUpload"`
	AllowList            string `json:"allowList"`
	AllowDelete          string `json:"allowDelete"`
	NotBefore            string `json:"notBefore"`
	NotAfter             string `json:"notAfter"`
}

// LoadStorjConfiguration reads and parses the JSON file that contain Storj configuration information.
func LoadStorjConfiguration(fullFileName string) ConfigStorj {

	var configStorj ConfigStorj
	fileHandle, err := os.Open(fullFileName)
	if err != nil {
		log.Fatal("Could not load storj config file: ", err)
	}
	defer fileHandle.Close()

	jsonParser := json.NewDecoder(fileHandle)
	if err = jsonParser.Decode(&configStorj); err != nil {
		log.Fatal(err)
	}
	// Display storj configuration read from file.
	fmt.Println("\nRead Storj configuration from the ", fullFileName, " file")
	fmt.Println("\nAPI Key\t\t: ", configStorj.APIKey)
	fmt.Println("Satellite	: ", configStorj.Satellite)
	fmt.Println("Bucket		: ", configStorj.Bucket)

	// Convert the upload path to standard form.
	checkSlash := configStorj.UploadPath[len(configStorj.UploadPath)-1:]
	if checkSlash != "/" {
		configStorj.UploadPath = configStorj.UploadPath + "/"
	}

	fmt.Println("Upload Path\t: ", configStorj.UploadPath)
	fmt.Println("Serialized Access Key\t: ", configStorj.SerializedAccess)
	return configStorj
}

// ShareAccess generates and prints the shareable serialized access
// as per the restrictions provided by the user.
func ShareAccess(access *uplink.Access, configStorj ConfigStorj) {

	allowDownload, _ := strconv.ParseBool(configStorj.AllowDownload)
	allowUpload, _ := strconv.ParseBool(configStorj.AllowUpload)
	allowList, _ := strconv.ParseBool(configStorj.AllowList)
	allowDelete, _ := strconv.ParseBool(configStorj.AllowDelete)
	notBefore, _ := time.Parse("2006-01-02_15:04:05", configStorj.NotBefore)
	notAfter, _ := time.Parse("2006-01-02_15:04:05", configStorj.NotAfter)

	permission := uplink.Permission{
		AllowDownload: allowDownload,
		AllowUpload:   allowUpload,
		AllowList:     allowList,
		AllowDelete:   allowDelete,
		NotBefore:     notBefore,
		NotAfter:      notAfter,
	}

	// Create shared access.
	sharedAccess, err := access.Share(permission)
	if err != nil {
		log.Fatal("Could not generate shared access: ", err)
	}

	// Generate restricted serialized access.
	serializedAccess, err := sharedAccess.Serialize()
	if err != nil {
		log.Fatal("Could not serialize shared access: ", err)
	}
	fmt.Println("Shareable sererialized access: ", serializedAccess)
}

// ConnectToStorj reads Storj configuration from given file
// and connects to the desired Storj network.
// It then reads data property from an external file.
func ConnectToStorj(fullFileName string, configStorj ConfigStorj, accesskey bool) (*uplink.Access, *uplink.Project) {

	var access *uplink.Access
	var cfg uplink.Config

	// Configure the UserAgent
	cfg.UserAgent = "InfluxDB"
	ctx := context.Background()
	var err error

	if accesskey {
		fmt.Println("\nRequesting access to storj network using Serialized access.")
		// Generate access handle using serialized access.
		access, err = uplink.ParseAccess(configStorj.SerializedAccess)
		if err != nil {
			log.Fatal("Could not parse access: ", err)
		}
	} else {
		fmt.Println("\nRequesting access to storj network using Api key.")
		// Generate access handle using API key, satellite url and encryption passphrase.
		access, err = cfg.RequestAccessWithPassphrase(ctx, configStorj.Satellite, configStorj.APIKey, configStorj.EncryptionPassphrase)
		if err != nil {
			log.Fatal("Could not request access grant: ", err)
		}
	}
	fmt.Println("Access to storj, granted.")

	fmt.Println("\nOpening new project.")
	// Open a new porject.
	project, err := cfg.OpenProject(ctx, access)
	if err != nil {
		log.Fatal("Could not open project:", err)
	}
	defer project.Close()

	fmt.Println("\nOpening Bucket: ", configStorj.Bucket)
	// Ensure the desired Bucket within the Project
	_, err = project.EnsureBucket(ctx, configStorj.Bucket)
	if err != nil {
		log.Fatal(err)
	}

	return access, project
}

// UploadData uploads the backup file to storj network.
func UploadData(project *uplink.Project, configStorj ConfigStorj, uploadFileName string, filePath string) {

	ctx := context.Background()

	fmt.Println("\nUpload Object Path: ", configStorj.UploadPath+uploadFileName)

	// Create an upload handle.
	upload, err := project.UploadObject(ctx, configStorj.Bucket, configStorj.UploadPath+uploadFileName, nil)
	if err != nil {
		log.Fatal("Could not initiate upload : ", err)
	}
	fmt.Println("\nUploading of the object to the Storj bucket: Initiated...")

	var lastIndex int64
	var numOfBytesRead int
	lastIndex = 0
	var buf = make([]byte, 32768)

	fileReader, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}

	var err1 error
	// Loop to append to backup file contents to the upload object.
	for err1 != io.EOF {
		sectionReader := io.NewSectionReader(fileReader, lastIndex, int64(cap(buf)))
		numOfBytesRead, err1 = sectionReader.ReadAt(buf, 0)
		if numOfBytesRead > 0 {
			reader := bytes.NewBuffer(buf[0:numOfBytesRead])
			// Try to upload data on storj n number of times
			retry := 0
			for retry < MAXRETRY {
				_, err = io.Copy(upload, reader)
				if err != nil {
					retry++
				} else {
					break
				}
			}
			if retry == MAXRETRY {
				log.Fatal("Could not upload data to storj: ", err)
			}

		}

		lastIndex = lastIndex + int64(numOfBytesRead)
	}
	fmt.Println("\nCommiting Object on storj V3 Network")
	// Commit the upload after copying the complete content of the backup file to upload object.
	err = upload.Commit()
	if err != nil {
		log.Fatal("Could not commit object upload : ", err)
	}

	fileReader.Close()

	// Delete the temporary file ftom local disk after uploading.
	if err = os.Remove(filePath); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Uploading object to Storj bucket: Completed!")
}

// DownloadData function downloads the data from storj bucket after upload to verify data is uploaded successfully.
func DownloadData(project *uplink.Project, configStorj ConfigStorj, downloadFileName string) {

	ctx := context.Background()

	directory, file := filepath.Split(downloadFileName)

	var receivedContents = []byte{}
	var lastIndex int64
	var buf = make([]byte, 32768)
	lastIndex = 0

	object, err := project.StatObject(ctx, configStorj.Bucket, configStorj.UploadPath+downloadFileName)
	if err != nil {
		log.Fatal(err)
	}

	// Loop to read the object in chunks and store the read data in a byte array.
	for int64(len(receivedContents)) < object.System.ContentLength {
		var download *uplink.Download
		retry := 0
		// Try to download data from storj n number of times
		for retry < MAXRETRY {
			download, err = project.DownloadObject(ctx, configStorj.Bucket, configStorj.UploadPath+downloadFileName, &uplink.DownloadOptions{Offset: lastIndex, Length: int64(cap(buf))})
			if err != nil {
				retry++
			} else {
				break
			}
		}
		if retry == MAXRETRY {
			log.Fatal("Could not download data form storj: ", err)
		}

		data, err2 := ioutil.ReadAll(download)
		if err2 != nil {
			break
		}

		receivedContents = append(receivedContents, data...)
		lastIndex = lastIndex + int64(len(data))
	}

	if _, err := os.Stat("./debug"); os.IsNotExist(err) {
		err1 := os.Mkdir("./debug", 0755)
		if err1 != nil {
			log.Fatal("Invalid Download Path: ", err1)
		}
	}
	if _, err := os.Stat("./debug" + "/" + directory); os.IsNotExist(err) {
		err1 := os.Mkdir("./debug"+"/"+directory, 0755)
		if err1 != nil {
			log.Fatal("Invalid Download Path: ", err1)
		}
	}

	// Create/open file in append mode.
	downloadFileDisk, err := os.OpenFile(filepath.Join("./debug", directory, file), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer downloadFileDisk.Close()

	// Write the contents retrieved from downloaded object to file on local disk.
	_, err = downloadFileDisk.Write(receivedContents)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Downloaded %d bytes of data.\n", len(receivedContents))
	fmt.Printf("Debug file \"%s\" downloaded to \"%s\"\n\n", downloadFileName, "debug/")
}
