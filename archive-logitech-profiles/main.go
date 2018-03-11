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

var (
	fileBaseName string
)

func init() {
	fileBaseName = strings.TrimRight(filepath.Base(os.Args[0]),
		filepath.Ext(os.Args[0]))

	flag.Usage = func() {
		u := fmt.Sprintf(`Archives the Logitech profile XMLs and then uploades the archive to DigitalOcean Spaces.

Create a config file named %s.json by filling in what is defined in its sample file.
`, fileBaseName)
		fmt.Fprint(os.Stderr, u)
	}
	flag.Parse()

	viper.SetConfigName(fileBaseName)
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
	target := fmt.Sprintf(viper.GetString("tempFilenameTpl"), formattedTime)
	// make sure the temporary local archive gets deleted
	defer os.Remove(target)

	sources := []string{
		filepath.Join(viper.GetString("sourceBaseDir"), "profiles"),
	}

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
			log.Fatal(err)
		}
	}

	fmt.Println("Temporary archive: " + target)

	archive.Close()
	zipFile.Close()

	fmt.Print("Uploading to DO-Spaces...")
	err = putOnS3(target)
	if err != nil {
		log.Printf("%s\n", err.Error())
	}
	fmt.Println("done.")

}

func putOnS3(filePath string) error {
	sess, err := session.NewSession(&aws.Config{
		Endpoint: aws.String(viper.GetString("doSpacesEndpoint")),
		Region:   aws.String(viper.GetString("doSpacesRegion")),
		Credentials: credentials.NewStaticCredentials(
			viper.GetString("doSpacesKey"),
			viper.GetString("doSpacesSecret"),
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
		Bucket: aws.String(viper.GetString("doSpacesBucket")),
		Key: aws.String(fmt.Sprintf("/backups/%s",
			filepath.Base(filePath))),
		Body: file,
		// https://docs.aws.amazon.com/AmazonS3/latest/dev/acl-overview.html#canned-acl
		ACL:             aws.String("private"),
		ContentType:     aws.String("application/atom+xml"),
		ContentEncoding: aws.String("utf-8"),
	})

	return err
}
