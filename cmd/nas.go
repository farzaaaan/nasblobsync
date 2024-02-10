package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	model "github.com/farzaaaan/nasblobsync/cmd/models"
	"github.com/farzaaaan/nasblobsync/cmd/utils"
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

type ConcurrentFileDetails struct {
	sync.Mutex
	FileDetails model.FileDetails
}

type Progress struct {
	sync.Mutex
	Processed int
	Total     int
}

func GetLocal(rootDir string) error {
	fileMap := make(map[string]*ConcurrentFileDetails)
	var wg sync.WaitGroup
	var mutex sync.Mutex
	progress := Progress{}
	var workerFunc func(string)
	workerFunc = func(path string) {
		defer wg.Done()

		info, err := os.Stat(path)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		if info.IsDir() {
			err := filepath.Walk(path, func(subPath string, subInfo os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if subInfo.IsDir() {
					return nil
				}
				if utils.ShouldIgnoreFile(subPath) {
					return nil
				}

				relPath, err := filepath.Rel(rootDir, subPath)
				if err != nil {
					return err
				}

				mutex.Lock()
				defer mutex.Unlock()

				if _, ok := fileMap[relPath]; !ok {
					fileMap[relPath] = &ConcurrentFileDetails{}
				}

				fileMap[relPath].Lock()
				defer fileMap[relPath].Unlock()

				fileMap[relPath].FileDetails = model.FileDetails{
					LastModified: subInfo.ModTime(),
					Size:         subInfo.Size(),
				}

				progress.Lock()
				progress.Processed++
				fmt.Printf("\rProgress: %d/%d", progress.Processed, progress.Total)
				progress.Unlock()

				return nil
			})
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
		}
	}

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			totalFiles, err := countFilesInDir(path)
			if err != nil {
				fmt.Println("Error:", err)
				return nil
			}

			progress.Lock()
			progress.Total += totalFiles
			progress.Unlock()

			wg.Add(1)
			go workerFunc(path)
		}
		return nil
	})

	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	wg.Wait()

	finalFileMap := make(map[string]model.FileDetails)
	for key, value := range fileMap {
		value.Lock()
		finalFileMap[key] = value.FileDetails
		value.Unlock()
	}

	writeToFile(finalFileMap)

	return nil
}
func countFilesInDir(dir string) (int, error) {
	count := 0
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && !utils.ShouldIgnoreFile(path) {
			count++
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return count, nil
}

func writeToFile(fileMap map[string]model.FileDetails) {
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

	fmt.Print("\nFile details saved to file_details.json\n")
}
