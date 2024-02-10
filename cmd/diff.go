package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/farzaaaan/nasblobsync/cmd/models"
	"github.com/spf13/cobra"
)

var (
	sourceFile  string
	compareFile string
)

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "find diff",
	Run: func(cmd *cobra.Command, args []string) {
		err := GetDiff(sourceFile, compareFile)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)

	diffCmd.Flags().StringVarP(&sourceFile, "source", "s", "file_details.json", "source file")
	// diffCmd.MarkFlagRequired("source")

	diffCmd.Flags().StringVarP(&compareFile, "compare", "c", "blob_details.json", "compare file")
	// diffCmd.MarkFlagRequired("compare")
}

func GetDiff(sourceFile, compareFile string) error {

	sourceData, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		return err
	}

	compareData, err := ioutil.ReadFile(compareFile)
	if err != nil {
		return err
	}

	var sourceMap map[string]models.FileDetails
	err = json.Unmarshal(sourceData, &sourceMap)
	if err != nil {
		return err
	}
	sourceMap = toLowerKeys(sourceMap)

	var compareMap map[string]models.FileDetails
	err = json.Unmarshal(compareData, &compareMap)
	if err != nil {
		return err
	}
	compareMap = toLowerKeys(compareMap)

	diffMap := make(map[string]models.FileDetails)

	var missingKeys, differentFiles, sizeMismatch, modifiedDateMismatch int

	for key, sourceDetails := range sourceMap {
		compareDetails, exists := compareMap[key]
		if !exists {

			diffMap[key] = sourceDetails
			missingKeys++
		} else {

			if sourceDetails.LastModified != compareDetails.LastModified ||
				sourceDetails.Size != compareDetails.Size {

				diffMap[key] = sourceDetails
				differentFiles++
				if sourceDetails.Size != compareDetails.Size {
					sizeMismatch++
				}
				if !sourceDetails.LastModified.Equal(compareDetails.LastModified) {
					modifiedDateMismatch++
				}
			}
		}
	}

	diffData, err := json.MarshalIndent(diffMap, "", "  ")
	if err != nil {
		return err
	}

	diffFile, err := os.Create("diff.json")
	if err != nil {
		return err
	}
	defer diffFile.Close()

	_, err = diffFile.Write(diffData)
	if err != nil {
		return err
	}

	fmt.Println("Differences saved to diff.json")

	meta := map[string]int{
		"total_source_count":     len(sourceMap),
		"total_compare_count":    len(compareMap),
		"missing_keys_count":     missingKeys,
		"different_files_count":  differentFiles,
		"size_mismatch_count":    sizeMismatch,
		"modified_date_mismatch": modifiedDateMismatch,
	}

	metaData, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}

	metaFile, err := os.Create("diff_meta.json")
	if err != nil {
		return err
	}
	defer metaFile.Close()

	_, err = metaFile.Write(metaData)
	if err != nil {
		return err
	}

	fmt.Println("Diff metadata saved to diff_meta.json")

	return nil
}

func toLowerKeys(m map[string]models.FileDetails) map[string]models.FileDetails {
	lowerMap := make(map[string]models.FileDetails)
	for k, v := range m {
		lowerMap[strings.ToLower(k)] = v
	}
	return lowerMap
}
