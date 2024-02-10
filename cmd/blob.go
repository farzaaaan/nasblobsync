package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/farzaaaan/nasblobsync/cmd/models"
	"github.com/farzaaaan/nasblobsync/cmd/utils"
	"github.com/spf13/cobra"
)

var (
	storageAccount    string
	container         string
	storageAccountKey string
	initialPrefix     string
)
var blobCmd = &cobra.Command{
	Use:   "blob",
	Short: "Traverse blob container",
	Long:  "Traverse blob details and output finding as map[string]{fileSize, lasModified} to blob_details.json",
	Run: func(cmd *cobra.Command, args []string) {
		err := GetBlob(storageAccount, container, storageAccountKey, initialPrefix)
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
	blobCmd.Flags().StringVar(&container, "prefix", "", "prefix")
	blobCmd.Flags().StringVar(&storageAccountKey, "storage-account-key", "", "Azure Storage Account Connection String or Key")
	blobCmd.MarkFlagRequired("storage-account")
	blobCmd.MarkFlagRequired("container")
	blobCmd.MarkFlagRequired("storage-account-key")
}

func GetBlob(storageAccount, container, accountKey, initialPrefix string) error {
	credential, err := azblob.NewSharedKeyCredential(storageAccount, accountKey)
	if err != nil {
		return err
	}

	p := azblob.NewPipeline(credential, azblob.PipelineOptions{})

	URL, _ := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net/%s", storageAccount, container))
	containerURL := azblob.NewContainerURL(*URL, p)

	ctx := context.Background()

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		blobMap = make(map[string]models.FileDetails)
	)

	var processPrefix func(ctx context.Context, prefix string)
	processPrefix = func(ctx context.Context, prefix string) {
		defer wg.Done()

		blobList, err := containerURL.ListBlobsHierarchySegment(ctx, azblob.Marker{}, "/", azblob.ListBlobsSegmentOptions{})
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		for _, blob := range blobList.Segment.BlobItems {
			blobName := strings.TrimPrefix(blob.Name, "/"+container+"/")
			if utils.ShouldIgnoreFile(blobName) {
				continue
			}

			props, err := containerURL.NewBlobURL(blob.Name).GetProperties(ctx, azblob.BlobAccessConditions{}, azblob.ClientProvidedKeyOptions{})
			if err != nil {
				fmt.Println("Error:", err)
				return
			}

			mu.Lock()
			blobMap[blob.Name] = models.FileDetails{
				Size: props.ContentLength(),
			}
			mu.Unlock()
		}

		for _, dir := range blobList.Segment.BlobPrefixes {
			dirName := strings.TrimSuffix(strings.TrimPrefix(dir.Name, "/"+container+"/"), "/")
			newPrefix := prefix + "/" + dirName
			wg.Add(1)
			go processPrefix(ctx, newPrefix)
		}
	}

	wg.Add(1)
	go processPrefix(ctx, initialPrefix)

	wg.Wait()

	// Marshal and write to blob_details.json file
	blobDetailsFile, err := os.Create("blob_details.json")
	if err != nil {
		return err
	}
	defer blobDetailsFile.Close()

	jsonData, err := json.MarshalIndent(blobMap, "", "  ")
	if err != nil {
		return err
	}
	_, err = blobDetailsFile.Write(jsonData)
	if err != nil {
		return err
	}

	fmt.Println("Blob details saved to blob_details.json")
	return nil
}
