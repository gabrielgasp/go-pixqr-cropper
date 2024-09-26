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
	files, err := os.ReadDir(inputDir)
	if err != nil {
		log.Fatalf("Failed to read input directory: %v", err)
	}

	jobs := make(chan fs.DirEntry, len(files))
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker(jobs, &wg)
	}

	for _, file := range files {
		jobs <- file
	}

	close(jobs)
	wg.Wait()

	fmt.Println("Images cropped successfully")
}

func worker(jobs <-chan fs.DirEntry, wg *sync.WaitGroup) {
	defer wg.Done()

	for file := range jobs {
		cropImage(file)
	}
}

func cropImage(file fs.DirEntry) {
	if !shouldCrop(file) {
		return
	}

	inputFilePath := filepath.Join(inputDir, file.Name())

	img, err := imaging.Open(inputFilePath)
	if err != nil {
		log.Printf("failed to open image '%s': %v", file.Name(), err)
		return
	}

	bounds := img.Bounds()
	outputWidth := bounds.Dx() - leftOffset - rightOffset
	outputHeight := bounds.Dy() - topOffset - bottomOffset
	cropRect := image.Rect(leftOffset, topOffset, leftOffset+outputWidth, topOffset+outputHeight)

	croppedImg := imaging.Crop(img, cropRect)

	outputFilePath := filepath.Join(outputDir, file.Name())

	err = imaging.Save(croppedImg, outputFilePath)
	if err != nil {
		log.Printf("failed to save image '%s': %v", file.Name(), err)
		return
	}

	err = os.Remove(inputFilePath)
	if err != nil {
		log.Printf("failed to delete original image '%s': %v", file.Name(), err)
		return
	}
}

func shouldCrop(file fs.DirEntry) bool {
	if file.IsDir() || strings.HasPrefix(file.Name(), ".") {
		return false
	}

	return true
}
