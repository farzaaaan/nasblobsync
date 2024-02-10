package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"

	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/farzaaaan/nasblobsync/cmd/models"
	"github.com/farzaaaan/nasblobsync/cmd/utils"
	"github.com/spf13/cobra"
)

var (
	storageAccount          string
	container               string
	storageAccountKeyOrConn string
)
var blobCmd = &cobra.Command{
	Use:   "blob",
	Short: "Traverse blob container",
	Long:  "Traverse blob details and output finding as map[string]{fileSize, lasModified} to blob_details.json",
	Run: func(cmd *cobra.Command, args []string) {
		err := GetBlob(storageAccount, container, storageAccountKeyOrConn)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(blobCmd)

	blobCmd.Flags().StringVar(&storageAccount, "storage-account", "", "Azure Storage Account Name")
	blobCmd.Flags().StringVar(&container, "container", "", "Azure Storage Container Name")
	blobCmd.Flags().StringVar(&storageAccountKeyOrConn, "storage-account-connection-string", "", "Azure Storage Account Connection String or Key")
	blobCmd.MarkFlagRequired("storage-account")
	blobCmd.MarkFlagRequired("container")
	blobCmd.MarkFlagRequired("storage-account-connection-string")
}

func GetBlob(storageAccount, container, storageAccountKeyOrConn string) error {

	credential, err := azblob.NewSharedKeyCredential(storageAccount, storageAccountKeyOrConn)
	if err != nil {
		return err
	}

	p := azblob.NewPipeline(credential, azblob.PipelineOptions{})

	URL, _ := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net/%s", storageAccount, container))
	containerURL := azblob.NewContainerURL(*URL, p)

	blobMap := make(map[string]models.FileDetails)

	ctx := context.Background()
	blobList, err := containerURL.ListBlobsFlatSegment(ctx, azblob.Marker{}, azblob.ListBlobsSegmentOptions{})
	if err != nil {
		return err
	}

	for _, blob := range blobList.Segment.BlobItems {

		blobName := strings.TrimPrefix(blob.Name, "/"+container+"/")
		blobURL := containerURL.NewBlobURL(blobName)
		props, err := blobURL.GetProperties(ctx, azblob.BlobAccessConditions{}, azblob.ClientProvidedKeyOptions{})
		if err != nil {
			return err
		}

		if utils.ShouldIgnoreFile(blobName) {
			return nil
		}

		blobMap[blobName] = models.FileDetails{
			LastModified: props.LastModified(),
			Size:         props.ContentLength(),
		}
	}

	var keys []string
	for key := range blobMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	sortedBlobMap := make(map[string]models.FileDetails)
	for _, key := range keys {
		sortedBlobMap[key] = blobMap[key]
	}

	jsonData, err := json.MarshalIndent(sortedBlobMap, "", "  ")
	if err != nil {
		return err
	}

	outputFile, err := os.Create("blob_details.json")
	if err != nil {
		return err
	}
	defer outputFile.Close()

	_, err = outputFile.Write(jsonData)
	if err != nil {
		return err
	}

	fmt.Println("Blob details saved to blob_details.json")
	return nil
}
