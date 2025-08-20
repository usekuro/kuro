package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func exportCollection(source, output string) error {
	outFile, err := os.Create(output)
	if err != nil {
		return err
	}
	defer outFile.Close()

	zipWriter := zip.NewWriter(outFile)
	defer zipWriter.Close()

	err = filepath.Walk(source, func(path string, info os.FileInfo, _ error) error {
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}

		fw, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		fs, err := os.Open(path)
		if err != nil {
			return err
		}
		defer fs.Close()

		_, err = io.Copy(fw, fs)
		return err
	})

	if err != nil {
		return err
	}
	fmt.Println("🎉 Export completed to:", output)
	return nil
}
