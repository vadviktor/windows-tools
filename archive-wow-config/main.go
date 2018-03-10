package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"
)

const (
	sourceBaseDir          = `E:\World of Warcraft`
	destinationFilenameTpl = `D:\gdrive\backups\wow-%s.zip`
)

func main() {
	t := time.Now()
	formattedTime := fmt.Sprintf("%04d%02d%02d_%02d%02d%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
	target := fmt.Sprintf(destinationFilenameTpl, formattedTime)

	sources := []string{
		filepath.Join(sourceBaseDir, "Interface"),
		filepath.Join(sourceBaseDir, "WTF"),
	}

	zipFile, err := os.Create(target)
	if err != nil {
		log.Fatalf("Failed to create %s: %s\n", target, err.Error())
	}
	defer zipFile.Close()
	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	for _, source := range sources {
		basePath := filepath.Dir(source)

		err = filepath.Walk(source, func(filePath string, fileInfo os.FileInfo, err error) error {
			if err != nil || fileInfo.IsDir() {
				return err
			}

			relativeFilePath, err := filepath.Rel(basePath, filePath)
			if err != nil {
				return err
			}

			archivePath := path.Join(filepath.SplitList(relativeFilePath)...)
			fmt.Printf("Archiving: %s\n", archivePath)

			file, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer file.Close()

			zipFileWriter, err := archive.Create(archivePath)
			if err != nil {
				return err
			}

			_, err = io.Copy(zipFileWriter, file)
			return err
		})

		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("Zipped File: " + target)
}