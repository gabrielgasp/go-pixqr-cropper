package main

import (
	"fmt"
	"image"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/disintegration/imaging"
)

var workerCount = runtime.NumCPU()

const (
	topOffset    = 650
	bottomOffset = 230
	leftOffset   = 150
	rightOffset  = 150
)

var inputDir = "input"
var outputDir = "output"

func main() {
	dirItems, err := os.ReadDir(inputDir)
	if err != nil {
		log.Fatalf("Failed to read input directory: %v", err)
	}

	jobs := make(chan fs.DirEntry, len(dirItems))
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker(jobs, &wg)
	}

	for _, dirItem := range dirItems {
		jobs <- dirItem
	}

	close(jobs)
	wg.Wait()

	fmt.Println("Images cropped successfully")
}

func worker(jobs <-chan fs.DirEntry, wg *sync.WaitGroup) {
	defer wg.Done()

	for dirItem := range jobs {
		if dirItem.IsDir() || strings.HasPrefix(dirItem.Name(), ".") {
			continue
		}

		cropImage(dirItem.Name())
	}
}

func cropImage(imageName string) {
	inputFilePath := filepath.Join(inputDir, imageName)

	img, err := imaging.Open(inputFilePath)
	if err != nil {
		log.Printf("failed to open image '%s': %v", imageName, err)
		return
	}

	bounds := img.Bounds()
	outputWidth := bounds.Dx() - leftOffset - rightOffset
	outputHeight := bounds.Dy() - topOffset - bottomOffset
	cropRect := image.Rect(leftOffset, topOffset, leftOffset+outputWidth, topOffset+outputHeight)

	croppedImg := imaging.Crop(img, cropRect)

	outputFilePath := filepath.Join(outputDir, imageName)

	err = imaging.Save(croppedImg, outputFilePath)
	if err != nil {
		log.Printf("failed to save image '%s': %v", imageName, err)
		return
	}

	err = os.Remove(inputFilePath)
	if err != nil {
		log.Printf("failed to delete original image '%s': %v", imageName, err)
		return
	}
}
