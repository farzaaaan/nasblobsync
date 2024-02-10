package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	model "github.com/farzaaaan/nasblobsync/cmd/models"
	"github.com/spf13/cobra"
)

var localDir string

var nasCmd = &cobra.Command{
	Use:   "local",
	Short: "Traverse Local Nas",
	Long:  "Traverse Local Nas and output finding as map[string]{fileSize, lasModified} to file_details.json",
	Run: func(cmd *cobra.Command, args []string) {

		err := GetLocal(localDir)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	},
}

func init() {

	rootCmd.AddCommand(nasCmd)

	nasCmd.Flags().StringVarP(&localDir, "local-dir", "d", ".", "Root directory for file information")
	nasCmd.MarkFlagRequired("local-dir")
}

func GetLocal(rootDir string) error {

	fileMap := make(map[string]model.FileDetails)

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		fileMap[path] = model.FileDetails{
			LastModified: info.ModTime(),
			Size:         info.Size(),
		}
		return nil
	})

	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	var keys []string
	for key := range fileMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	sortedFileMap := make(map[string]model.FileDetails)
	for _, key := range keys {
		sortedFileMap[key] = fileMap[key]
	}

	jsonData, err := json.MarshalIndent(sortedFileMap, "", "  ")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	outputFile, err := os.Create("file_details.json")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer outputFile.Close()

	_, err = outputFile.Write(jsonData)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	fmt.Println("File details saved to file_details.json")
	return nil
}
