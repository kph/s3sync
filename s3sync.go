// Copyright 2019 Kevin Paul Herbert

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var fileMap map[string]int64
var uploader *s3manager.Uploader

func uploadFile(objname string, in io.Reader) (err error) {
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String("allanmh-backup"),
		Key:    aws.String(objname),
		Body:   in,
	})
	if err != nil {
		return fmt.Errorf("Unable to upload %s: %w", objname, err)
	}
	return nil
}

func transferFile(fn string) (err error) {
	fullname := "samsung-T3/" + fn
	if _, found := fileMap[fullname]; found {
		fmt.Printf("Found %s\n", fullname)
		return nil
	}

	file, err := os.Open(fn)
	if err != nil {
		return fmt.Errorf("Error opening %s: %w", fn, err)
	}
	defer file.Close()
	fmt.Printf("Uploading %s\n", fn)

	err = uploadFile(fullname, file)
	if err != nil {
		return fmt.Errorf("Error uploading %s: %w", fn, err)
	}
	return
}

func transferDir(dir string) (err error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("ReadDir of %s failed: %w", dir, err)
	}
	for _, file := range files {
		fn := filepath.Join(dir, file.Name())
		if dir == "." {
			fn = file.Name()
		}
		//		fmt.Printf("Working on %s\n", fn)
		if file.IsDir() {
			err := transferDir(fn)
			if err != nil {
				fmt.Printf("Error reading directory %s: %s",
					fn, err)
			}
			continue
		}
		err := transferFile(fn)
		if err != nil {
			return err
		}
	}
	return nil
}

func getBucketContents(svc *s3.S3, bucket string) (err error) {
	// Get the list of items
	err = svc.ListObjectsV2Pages((&s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	}).SetEncodingType(s3.EncodingTypeUrl),
		func(page *s3.ListObjectsV2Output, lastPage bool) bool {
			for _, item := range page.Contents {
				fileMap[*item.Key] = *item.Size
			}

			return true
		})
	if err != nil {
		return fmt.Errorf("Unable to list items in bucket %s: %w",
			bucket, err)
	}

	fmt.Println("Found", len(fileMap), "items in bucket", bucket)
	fmt.Println("")
	return nil
}

func main() {
	fileMap = make(map[string]int64)

	// Initialize a session that the SDK will use to load credentials
	// from the shared credentials file ~/.aws/credentials.
	sess, err := session.NewSession()
	if err != nil {
		fmt.Printf("Error creating session: %s\n", err)
		os.Exit(1)
	}
	// Create S3 service client
	svc := s3.New(sess)

	// Setup the S3 Upload Manager. Also see the SDK doc for the Upload Manager
	// for more information on configuring part size, and concurrency.
	//
	// http://docs.aws.amazon.com/sdk-for-go/api/service/s3/s3manager/#NewUploader
	uploader = s3manager.NewUploader(sess)

	err = getBucketContents(svc, "allanmh-backup")
	if err != nil {
		fmt.Printf("Error getting bucket contents: %s\n", err)
		os.Exit(1)
	}

	err = transferDir(".")
	if err != nil {
		fmt.Printf("Error from transferDir: %s\n", err)
		os.Exit(1)
	}
}
