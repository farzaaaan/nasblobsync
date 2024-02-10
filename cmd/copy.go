package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	sourceDir      string
	destinationDir string
)

var copyCmd = &cobra.Command{
	Use:   "copy",
	Short: "Copy files from diff to a folder",
	Run: func(cmd *cobra.Command, args []string) {
		err := copyFilesToDestination(sourceDir, destinationDir)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(copyCmd)
	copyCmd.Flags().StringVarP(&sourceDir, "source", "s", ".", "Root directory for file")
	copyCmd.Flags().StringVarP(&destinationDir, "destination", "d", "", "Destination directory for file")
	copyCmd.MarkFlagRequired("destination")

}

func copyFilesToDestination(srcDir, destDir string) error {
	fmt.Println(srcDir)
	// Open the diff_flat file for reading
	file, err := os.Open("diff_flat")
	if err != nil {
		return err
	}
	defer file.Close()

	// Create the destination directory if it does not exist
	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		return err
	}

	// Read each file path from the diff_flat file
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// Get the file path
		filePath := scanner.Text()

		// Construct the source file path
		srcFilePath := filepath.Join(srcDir, filePath)
		fmt.Println(srcFilePath)
		// Create the destination file path
		destFilePath := filepath.Join(destDir, strings.ToLower(filePath))

		// Open the source file for reading
		srcFile, err := os.Open(srcFilePath)
		if err != nil {
			return err
		}
		defer srcFile.Close()
		if err := os.MkdirAll(filepath.Dir(destFilePath), os.ModePerm); err != nil {
			return err
		}
		// Create the destination file for writing
		destFile, err := os.Create(destFilePath)
		if err != nil {
			return err
		}
		defer destFile.Close()

		// Copy the content from source to destination
		if _, err := io.Copy(destFile, srcFile); err != nil {
			return err
		}

		// Close the files
		destFile.Close()
		srcFile.Close()
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	fmt.Println("Files copied to destination successfully")
	return nil
}
