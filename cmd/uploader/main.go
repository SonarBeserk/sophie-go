package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/SonarBeserk/sophie-go/internal/emote"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// Config represents the configuration for the bot
type Config struct {
	Emotes []emote.Emote `toml:"emote"`
	Gifs   []emote.Gif   `toml:"gif"`
}

// Variables used for command line parameters
var (
	emotesFile      string
	bucketName      string
	imagesDirectory string

	urlCleaner *strings.Replacer = strings.NewReplacer("https:", "", "/", "", ".", "_")
)

func init() {
	flag.StringVar(&emotesFile, "emotes", "./emotes.toml", "Path to file containing emotes")
	flag.StringVar(&bucketName, "bucket", "", "The name of the bucket to upload files to")
	flag.StringVar(&imagesDirectory, "images", "images", "Path to download images")
	flag.Parse()
}

func main() {
	// The session the S3 Uploader will use
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create an uploader with the session and default options
	uploader := s3manager.NewUploader(sess)

	conf, err := loadEmoteMaps(emotesFile)
	if err != nil {
		fmt.Printf("Error loading emotes file %s: %v\n", emotesFile, err)
		return
	}

	if _, err := os.Stat(imagesDirectory); os.IsNotExist(err) {
		err = os.Mkdir(imagesDirectory, 0755)
		if err != nil {
			fmt.Printf("Error creating images directory: %v\n", err)
			return
		}
	}

	for _, image := range conf.Gifs {
		_, err := url.ParseRequestURI(image.URL)
		if err != nil {
			continue
		}

		fileName := image.Verb + "-" + image.URL
		fileName = urlCleaner.Replace(fileName)
		extIndex := strings.LastIndexAny(fileName, "_")
		fileName = fileName[:extIndex] + "." + fileName[extIndex+1:]

		if strings.Contains(fileName, bucketName) {
			continue
		}

		fmt.Println("Uploading image: " + image.URL)

		err = downloadFile(image.URL, imagesDirectory+"/"+fileName)
		if err != nil {
			fmt.Printf("Failed to download file: %q, %v", fileName, err)
			return
		}

		f, err := os.Open(imagesDirectory + "/" + fileName)
		if err != nil {
			fmt.Printf("Failed to open file %q, %v", fileName, err)
			return
		}

		// Upload the file to S3.
		result, err := uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(fileName),
			Body:   f,
			ACL:    aws.String("public-read"),
		})
		if err != nil {
			fmt.Printf("Failed to upload file, %v", err)
			return
		}

		fmt.Printf("File uploaded to: %s\n", result.Location)
	}
}

func loadEmoteMaps(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var conf Config
	if _, err := toml.Decode(string(data), &conf); err != nil {
		return nil, err
	}

	return &conf, nil
}

func downloadFile(URL, fileName string) error {
	//Get the response bytes from the url
	response, err := http.Get(URL)
	if err != nil {
	}
	defer response.Body.Close()

	//Write the bytes to the file
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(fileName, data, 0644)
	if err != nil {
		return err
	}

	return nil
}
