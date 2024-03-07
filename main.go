package main

import (
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/PuerkitoBio/goquery"
	"github.com/nfnt/resize"
)

const IMAGE_RESIZE_MULTIPLIER = 2
const DOWNLOAD_DIRECTORY = "download"
const GENESIS_LINK = "https://www.commitstrip.com/en/2012/02/22/interview/"

var re *regexp.Regexp = regexp.MustCompile(`(\d{4})/(\d{2})/(\d{2})`)

type Image struct {
	Source   string
	NextLink string
}

func encodeImage(writer io.Writer, img image.Image) error {
	switch f := writer.(type) {
	case *os.File:
		return jpeg.Encode(f, img, nil)
	default:
		return errors.New("Unsupported output type")
	}
}

func downloadFile(url, fileName string) error {
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	img, _, err := image.Decode(response.Body)
	if err != nil {
		return err
	}

	resizedImg := resize.Resize(uint(img.Bounds().Dx()*IMAGE_RESIZE_MULTIPLIER), uint(img.Bounds().Dy()*IMAGE_RESIZE_MULTIPLIER), img, resize.NearestNeighbor)
	if err := encodeImage(file, resizedImg); err != nil {
		return err
	}

	return nil
}

func fetchImage(url string) (Image, error) {
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var year string
	var month string
	var day string

	matches := re.FindStringSubmatch(url)
	if len(matches) >= 4 {
		year = matches[1]
		month = matches[2]
		day = matches[3]
	}

	image := doc.Find(".entry-content img")
	imageSource, exists := image.Attr("src")
	if exists {
		log.Printf("Downloading image %s\n", imageSource)
		downloadFile(imageSource, fmt.Sprintf("%s/%s-%s-%s.jpg", DOWNLOAD_DIRECTORY, year, month, day))
	}

	nextLink := doc.Find(".nav-single .nav-next a")
	nextLinkHref, exists := nextLink.Attr("href")
	if !exists {
		return Image{}, errors.New("Next link not present")
	}

	return Image{
		Source:   imageSource,
		NextLink: nextLinkHref,
	}, nil
}

func main() {
	var hasNextLink bool = true
	var nextLink string = GENESIS_LINK

	for hasNextLink {
		item, _ := fetchImage(nextLink)
		log.Println(item)

		if len(item.NextLink) == 0 {
			hasNextLink = false
		}

		nextLink = item.NextLink
	}
}
