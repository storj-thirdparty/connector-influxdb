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

	"storj.io/storj/lib/uplink"
	"storj.io/storj/pkg/macaroon"
)

// DEBUG allows more detailed working to be exposed through the terminal.
var DEBUG = false

// ConfigStorj depicts keys to search for within the stroj_config.json file.
type ConfigStorj struct {
	APIKey               string `json:"apikey"`
	Satellite            string `json:"satellite"`
	Bucket               string `json:"bucket"`
	UploadPath           string `json:"uploadPath"`
	EncryptionPassphrase string `json:"encryptionpassphrase"`
	SerializedScope      string `json:"serializedScope"`
	DisallowReads        string `json:"disallowReads"`
	DisallowWrites       string `json:"disallowWrites"`
	DisallowDeletes      string `json:"disallowDeletes"`
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

	// Display read information.
	fmt.Println("\nRead Storj configuration from the ", fullFileName, " file")
	fmt.Println("\nAPI Key\t\t: ", configStorj.APIKey)
	fmt.Println("Satellite	: ", configStorj.Satellite)
	fmt.Println("Bucket		: ", configStorj.Bucket)
	fmt.Println("Upload Path\t: ", configStorj.UploadPath)
	fmt.Println("Serialized Scope Key\t: ", configStorj.SerializedScope)

	return configStorj, nil
}

// ConnectStorjReadUploadData reads Storj configuration from given file,
// connects to the desired Storj network.
/// It then reads data property from an external file.
func ConnectStorjReadUploadData(fullFileName string, keyValue string, restrict string) (context.Context, *uplink.Uplink, *uplink.Project, *uplink.Bucket, ConfigStorj, string, error) { // fullFileName for fetching storj V3 credentials from  given JSON filename
	// databaseReader is an io.Reader implementation that 'reads' desired data,
	// which is to be uploaded to storj V3 network.
	// databaseName for adding dataBase name in storj V3 filename.
	// Read Storj bucket's configuration from an external file.
	var scope string = ""
	configStorj, err := LoadStorjConfiguration(fullFileName)
	if err != nil {
		log.Fatal("loadStorjConfiguration:", err)
	}

	fmt.Println("\nCreating New Uplink...")

	var cfg uplink.Config
	// Configure the partner id
	cfg.Volatile.PartnerID = "a1ba07a4-e095-4a43-914c-1d56c9ff5afd"

	ctx := context.Background()

	uplinkstorj, err := uplink.NewUplink(ctx, &cfg)
	if err != nil {
		uplinkstorj.Close()
		log.Fatal("Could not create new Uplink object:", err)
	}
	var serializedScope string
	if keyValue == "key" {

		fmt.Println("Parsing the API key...")
		key, err := uplink.ParseAPIKey(configStorj.APIKey)
		if err != nil {
			uplinkstorj.Close()
			log.Fatal("Could not parse API key:", err)
		}

		if DEBUG {
			fmt.Println("API key \t   :", configStorj.APIKey)
			fmt.Println("Serialized API key :", key.Serialize())
		}

		fmt.Println("Opening Project...")
		proj, err := uplinkstorj.OpenProject(ctx, configStorj.Satellite, key)

		if err != nil {
			CloseProject(uplinkstorj, proj, nil)
			log.Fatal("Could not open project:", err)
		}

		// Creating an encryption key from encryption passphrase.
		if DEBUG {
			fmt.Println("\nGetting encryption key from pass phrase...")
		}

		encryptionKey, err := proj.SaltedKeyFromPassphrase(ctx, configStorj.EncryptionPassphrase)
		if err != nil {
			CloseProject(uplinkstorj, proj, nil)
			log.Fatal("Could not create encryption key:", err)
		}

		// Creating an encryption context.
		access := uplink.NewEncryptionAccessWithDefaultKey(*encryptionKey)

		if DEBUG {
			fmt.Println("Encryption access \t:", configStorj.EncryptionPassphrase)
		}

		// Serializing the parsed access, so as to compare with the original key.
		serializedAccess, err := access.Serialize()
		if err != nil {
			CloseProject(uplinkstorj, proj, nil)
			log.Fatal("Error Serialized key : ", err)
		}

		if DEBUG {
			fmt.Println("Serialized access key\t:", serializedAccess)
		}

		// Load the existing encryption access context
		accessParse, err := uplink.ParseEncryptionAccess(serializedAccess)
		if err != nil {
			log.Fatal(err)
		}

		if restrict == "restrict" {
			disallowRead, _ := strconv.ParseBool(configStorj.DisallowReads)
			disallowWrite, _ := strconv.ParseBool(configStorj.DisallowWrites)
			disallowDelete, _ := strconv.ParseBool(configStorj.DisallowDeletes)
			userAPIKey, err := key.Restrict(macaroon.Caveat{
				DisallowReads:   disallowRead,
				DisallowWrites:  disallowWrite,
				DisallowDeletes: disallowDelete,
			})
			if err != nil {
				log.Fatal(err)
			}
			userAPIKey, userAccess, err := accessParse.Restrict(userAPIKey,
				uplink.EncryptionRestriction{
					Bucket:     configStorj.Bucket,
					PathPrefix: configStorj.UploadPath,
				},
			)
			if err != nil {
				log.Fatal(err)
			}
			userRestrictScope := &uplink.Scope{
				SatelliteAddr:    configStorj.Satellite,
				APIKey:           userAPIKey,
				EncryptionAccess: userAccess,
			}
			serializedRestrictScope, err := userRestrictScope.Serialize()
			if err != nil {
				log.Fatal(err)
			}
			scope = serializedRestrictScope
			//fmt.Println("Restricted serialized user scope", serializedRestrictScope)
		}
		userScope := &uplink.Scope{
			SatelliteAddr:    configStorj.Satellite,
			APIKey:           key,
			EncryptionAccess: access,
		}
		serializedScope, err = userScope.Serialize()
		if err != nil {
			log.Fatal(err)
		}
		if restrict == "" {
			scope = serializedScope
		}

		proj.Close()
		uplinkstorj.Close()
	} else {
		serializedScope = configStorj.SerializedScope

	}
	parsedScope, err := uplink.ParseScope(serializedScope)
	if err != nil {
		log.Fatal(err)
	}

	uplinkstorj, err = uplink.NewUplink(ctx, &cfg)
	if err != nil {
		log.Fatal("Could not create new Uplink object:", err)
	}
	proj, err := uplinkstorj.OpenProject(ctx, parsedScope.SatelliteAddr, parsedScope.APIKey)
	if err != nil {
		CloseProject(uplinkstorj, proj, nil)
		log.Fatal("Could not open project:", err)
	}
	fmt.Println("Opening Bucket\t: ", configStorj.Bucket)

	// Open up the desired Bucket within the Project.
	bucket, err := proj.OpenBucket(ctx, configStorj.Bucket, parsedScope.EncryptionAccess)
	//
	if err != nil {
		fmt.Println("Could not open bucket", configStorj.Bucket, ":", err)
		fmt.Println("Trying to create new bucket....")
		_, err1 := proj.CreateBucket(ctx, configStorj.Bucket, nil)
		if err1 != nil {
			CloseProject(uplinkstorj, proj, bucket)
			fmt.Printf("Could not create bucket %q:", configStorj.Bucket)
			log.Fatal(err1)
		} else {
			fmt.Println("Created Bucket", configStorj.Bucket)
		}
		fmt.Println("Opening created Bucket: ", configStorj.Bucket)
		bucket, err = proj.OpenBucket(ctx, configStorj.Bucket, parsedScope.EncryptionAccess)
		if err != nil {
			fmt.Printf("Could not open bucket %q: %s", configStorj.Bucket, err)
		}
	}

	return ctx, uplinkstorj, proj, bucket, configStorj, scope, err
}

// ConnectUpload uploads the data to storj network.
func ConnectUpload(ctx context.Context, bucket *uplink.Bucket, data io.Reader, databaseName string, fileNamesDEBUG []string, configStorj ConfigStorj, err error) []string {
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
		fmt.Println("\nUpload Object Path: ", configStorj.UploadPath+filename)
		fmt.Println("\nUploading of the object to the Storj bucket: Initiated...")
		err = bucket.UploadObject(ctx, configStorj.UploadPath+filename, data, nil)
		if err != nil {
			i++
		}
		if DEBUG {
			file = append(file, filename)
		}
	}

	if err != nil {
		fmt.Printf("Could not upload: %s", err)
		return nil
	}

	fmt.Println("Uploading object to Storj bucket: Completed!")
	return file
}

// Debug function downloads the data from storj bucket after upload to verify data is uploaded successfully.
func Debug(bucket *uplink.Bucket, configStorj ConfigStorj, lastFileName string, copied int) {

	if DEBUG {
		checkSlash := configStorj.UploadPath[len(configStorj.UploadPath)-1:]
		if checkSlash != "/" {
			configStorj.UploadPath = configStorj.UploadPath + "/"
		}
		ctx := context.Background()
		var cfg uplink.ListOptions
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

		cfg.Direction = 2
		list, err := bucket.ListObjects(ctx, &cfg)
		if err != nil {
			fmt.Println(err)
		}
		var downloadFileDisk *os.File
		for i := 0; i < len(list.Items); i++ {
			readBack, err := bucket.OpenObject(ctx, cfg.Prefix+strconv.Itoa(i)+"."+lastextension)
			if err != nil {
				fmt.Printf("Could not open object at %q: %v", cfg.Prefix+list.Items[i].Path, err)
				log.Fatal(err)
			}
			defer readBack.Close()

			fmt.Println("\nDownloading file uploaded on storj...")
			// We want the whole thing, so range from 0 to -1.
			strm, err := readBack.DownloadRange(ctx, 0, -1)
			if err != nil {
				fmt.Printf("Could not initiate download: %v", err)
			}
			defer strm.Close()
			fmt.Printf("Downloading Object %s from bucket : Initiated...\n", "debug/"+lastFileName)
			// Read everything from the stream.
			receivedContents, err := ioutil.ReadAll(strm)

			if err != nil {
				log.Fatal("could not read object:", err)
			}

			if _, err := os.Stat("./debug"); os.IsNotExist(err) {
				err1 := os.Mkdir("./debug", os.ModeDir)
				if err1 != nil {
					log.Fatal("Invalid Download Path")
				}
			}

			if _, err := os.Stat("./debug" + "/" + slitString[0]); os.IsNotExist(err) {
				err1 := os.Mkdir("./debug"+"/"+slitString[0], os.ModeDir)
				if err1 != nil {
					log.Fatal("Invalid Download Path")
				}
			}

			downloadFileDisk, _ = os.OpenFile("debug/"+lastFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			_, err = downloadFileDisk.Write(receivedContents)
			if err != nil {
				log.Fatal(err)
			}

			defer downloadFileDisk.Close()
			readBack.Close()
			strm.Close()
			downloadFileDisk.Close()
			fmt.Printf("Downloaded %d bytes of Object from bucket!\n", len(receivedContents))
		}

		fmt.Printf("File downloading: Complete!\n")
		fmt.Printf("\nDebug file \"%s\" downloaded to \"%s\"\n", lastFileName, "debug/")
	}

}

// CloseProject closes bucket, project and uplink.
func CloseProject(uplink *uplink.Uplink, proj *uplink.Project, bucket *uplink.Bucket) {
	if bucket != nil {
		bucket.Close()
	}

	if proj != nil {
		proj.Close()
	}

	if uplink != nil {
		uplink.Close()
	}
}
