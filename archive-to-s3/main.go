package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/spf13/viper"
)

func init() {
	if len(os.Args) == 1 {
		log.Fatalln("First argument has to be a config file.")
	}

	if _, err := os.Stat(os.Args[1]); os.IsNotExist(err) {
		log.Fatalln("First argument has to be a readable config file.")
	}

	flag.Usage = func() {
		u := fmt.Sprint("Archives the directories defined in the config file and then uploades the archive to AWS S3 compatible storage.",)
		fmt.Fprint(os.Stderr, u)
	}
	flag.Parse()

	viper.SetConfigName(strings.TrimRight(filepath.Base(os.Args[1]),
		filepath.Ext(os.Args[1])))
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("%s\n", err.Error())
	}
}

func main() {
	t := time.Now()
	formattedTime := fmt.Sprintf("%04d%02d%02d_%02d%02d%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())

	target := path.Join(viper.GetString("tempFilepath"),
		fmt.Sprintf(viper.GetString("archiveFilenameTpl"), formattedTime))
	// make sure the temporary local archive gets deleted
	defer os.Remove(target)

	sources := viper.GetStringSlice("sourceDirs")
	zipFile, err := os.Create(target)
	if err != nil {
		log.Fatalf("Failed to create %s: %s\n", target, err.Error())
	}
	archive := zip.NewWriter(zipFile)

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
			log.Fatalf("Error during archiving: %s\n", err.Error())
		}
	}

	fmt.Println("Temporary archive: " + target)

	archive.Close()
	zipFile.Close()

	fmt.Print("Uploading...")
	err = putOnS3(target)
	if err != nil {
		log.Printf("%s\n", err.Error())
	}
	fmt.Println("done.")

}

func putOnS3(filePath string) error {
	sess, err := session.NewSession(&aws.Config{
		Endpoint: aws.String(viper.GetString("s3Endpoint")),
		Region:   aws.String(viper.GetString("s3Region")),
		Credentials: credentials.NewStaticCredentials(
			viper.GetString("s3Key"),
			viper.GetString("s3Secret"),
			""),
	})
	if err != nil {
		return err
	}

	// upload
	uploader := s3manager.NewUploader(sess)

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(viper.GetString("s3Bucket")),
		Key: aws.String(path.Join(viper.GetString("s3Dir"), path.Base(filePath))),
		Body: file,
		// https://docs.aws.amazon.com/AmazonS3/latest/dev/acl-overview.html#canned-acl
		ACL:             aws.String("private"),
		ContentType:     aws.String("application/zip"),
		ContentEncoding: aws.String("utf-8"),
	})

	return err
}
