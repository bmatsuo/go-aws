// Does one PUT and one GET against a bucket
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	//"time"

	"github.com/bmatsuo/go-aws"
	"github.com/bmatsuo/go-aws/s3"
)

func main() {
	args := os.Args[1:]
	if len(args) < 2 {
		fmt.Println("usage:", os.Args[0], "BUCKET", "OBJECT", "[FILEPATH]")
		os.Exit(1)
	}
	bucket := args[0]
	object := args[1]
	path := "-"
	if len(args) >= 3 {
		path = args[2]
	}
	input := os.Stdin
	if path != "-" {
		var err error
		input, err = os.Open(path)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	defer input.Close()
	data, err := ioutil.ReadAll(input)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	creds := aws.Getenv()
	region := s3.USStandard
	client := s3.NewClient(creds, region)
	put, err := client.
		PutObject(bucket, object).
		Content(data).
		ContentType("text/plain").
		CacheControl("max-age=300").
		Expires(5 * time.Minute).
		Acl("public-read").
		Exec()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(put.Status())
	fmt.Println(put.Header.Get("Etag"))
	get, err := client.
		GetObject(bucket, object).
		Exec()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer get.Body.Close()
	p, err := ioutil.ReadAll(get.Body)
	fmt.Println(string(p), err)
}
