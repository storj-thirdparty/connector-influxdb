package cmd_test

import (
	"context"
	"fmt"
	"log"

	"testing"

	"github.com/storj-thirdparty/connector-influxdb/cmd"
	"storj.io/uplink"
)

func TestMongoStore(t *testing.T) {

	storjConfig := cmd.LoadStorjConfiguration("../config/storj_config_test.json")
	_, project := cmd.ConnectToStorj(storjConfig, false)

	fmt.Printf("Initiating back-up.\n")
	cmd.UploadData(project, storjConfig, "testdb/testFile.txt", "../testFile.txt")
	fmt.Printf("Back-up complete.\n\n")

	fmt.Printf("\nDeleting the test back-up.\n")
	ctx := context.Background()
	backups := project.ListObjects(ctx, storjConfig.Bucket, &uplink.ListObjectsOptions{Prefix: "testdb/"})
	// Loop to find the latest back-up of all the back-ups.
	for backups.Next() {
		item := backups.Item()
		_, err := project.DeleteObject(ctx, storjConfig.Bucket, item.Key)
		if err != nil {
			log.Fatal(err)
		}
	}
	fmt.Printf("Deleted the test back-up.\n")
}
