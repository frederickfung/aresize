package main

import (
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"sync"

	"github.com/gabriel-vasile/mimetype"
	"golang.org/x/image/draw"
)

func main() {
	var pattern string
	var prefix string
	var longestSide int
	var quality int
	var concurrentWorkers int
	var useCatmullRom bool

	flag.StringVar(&pattern, "p", "", "File Glob pattern to search for image files")
	flag.StringVar(&prefix, "pre", "resized_", "Prefix added to the resized image files' name")
	flag.IntVar(&longestSide, "long", 2560, "The length in pixel of the long side to resize the image")
	flag.IntVar(&quality, "q", 100, "JPEG Quality of the output image file. Only used for JPEG")
	flag.IntVar(&concurrentWorkers, "c", 4, "Concurrent conversions allowed. This is usually memory bound.")
	flag.BoolVar(&useCatmullRom, "betterResize", false, "Use CatmullRom instead of BiLinear for resize")

	flag.Parse()

	if pattern == "" {
		log.Fatal("[Main][Error] No File Glob pattern provided")
	}

	var wg sync.WaitGroup

	fmt.Printf("===\nFile Glob Pattern = %s\nResized Filename Prefix = %s\nLong Side in Pixel = %d\nJPEG Quality = %d\nConcurrency = %d\nUse CatmullRom Resize = %v\n===\n",
		pattern, prefix, longestSide, quality, concurrentWorkers, useCatmullRom)
	files, err := filepath.Glob(pattern)
	if err != nil {
		log.Fatalf("[Main][Error] Error opening path/pattern - %s", pattern)
	}

	totalFiles := len(files)
	if totalFiles <= 0 {
		log.Fatalf("[Main][Error] No files fond with pattern - %s", pattern)
	} else {
		// Naive way for "controlled" concurrency
		currentCount := 0
		fileCount := 1
		for _, filePath := range files {
			dstFilePath := filepath.Join(filepath.Dir(filePath), fmt.Sprintf("%s%s", prefix, filepath.Base(filePath)))
			fmt.Printf("[Main] (%d/%d) %s -> %s\n", fileCount, totalFiles, filePath, dstFilePath)
			currentCount++
			wg.Add(1)
			go copyOrResizeImageFile(filePath, dstFilePath, longestSide, quality, useCatmullRom, &wg)
			fileCount++
			if currentCount >= concurrentWorkers {
				wg.Wait()
				currentCount = 0
			}
		}
	}

	wg.Wait()
}

func copyOrResizeImageFile(srcFilePath string, dstFilePath string, longSideLength int, quality int, useCatmullRom bool, wg *sync.WaitGroup) {
	file, err := os.Open(srcFilePath)
	if err != nil {
		fmt.Println(fmt.Sprintf("[Resize][Error] Error opening file %s", srcFilePath), err)

		wg.Done()
		return
	}
	defer file.Close()

	mtype, _ := mimetype.DetectReader(file)
	file.Seek(0, io.SeekStart)

	var srcImage image.Image
	if mtype.String() == "image/jpeg" {
		srcImage, err = jpeg.Decode(file)
		if err != nil {
			fmt.Println(fmt.Sprintf("[Resize][Error] Error decoding file %s as JPG", srcFilePath), err)

			wg.Done()
			return
		}
	} else if mtype.String() == "image/png" {
		srcImage, err = png.Decode(file)
		if err != nil {
			fmt.Println(fmt.Sprintf("[Resize][Error] Error decoding file %s as PNG", srcFilePath), err)

			wg.Done()
			return
		}
	} else {
		fmt.Printf("[Resize][Error] File %s is of unsupported MIME type %s\n", srcFilePath, mtype.String())

		wg.Done()
		return
	}
	// Force collection of src file
	file = nil

	fmt.Printf("[Resize] Resizing %s...\n", srcFilePath)
	needResize, resizedImg := resizeImage(srcImage, longSideLength, useCatmullRom)
	// Force collection of srcImg
	srcImage = nil

	if !needResize {
		// Just copy the file
		fmt.Printf("[Copy] No resize required. Copying %s -> %s\n", srcFilePath, dstFilePath)
		copy(srcFilePath, dstFilePath)
	} else {
		dstFile, err := os.Create(dstFilePath)
		if err != nil {
			fmt.Println(fmt.Sprintf("[Resize][Error] Failed to create destination file %s", dstFilePath), err)

			wg.Done()
			return
		}
		defer dstFile.Close()

		fmt.Printf("[Resize] Writing to %s...\n", dstFilePath)
		if mtype.String() == "image/jpeg" {
			jpeg.Encode(dstFile, resizedImg, &jpeg.Options{Quality: quality})
		} else if mtype.String() == "image/png" {
			png.Encode(dstFile, resizedImg)
		}
		// Force collection of resizedImg
		resizedImg = nil
	}
	wg.Done()
}

func resizeImage(src image.Image, longSideLength int, useCatmullRom bool) (bool, image.Image) {
	if src.Bounds().Max.X <= longSideLength && src.Bounds().Max.Y <= longSideLength {
		// No need to process, image is already smaller than required
		return false, nil
	}

	newX := longSideLength
	newY := longSideLength
	if src.Bounds().Max.X >= src.Bounds().Max.Y {
		newY = int(math.Round(float64(longSideLength) / float64(src.Bounds().Max.X) * float64(src.Bounds().Max.Y)))
	} else {
		newX = int(math.Round(float64(longSideLength) / float64(src.Bounds().Max.Y) * float64(src.Bounds().Max.X)))
	}
	dst := image.NewRGBA(image.Rect(0, 0, newX, newY))

	// Resize
	if useCatmullRom {
		draw.CatmullRom.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)
	} else {
		draw.BiLinear.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)
	}
	return true, dst
}

func copy(srcPath string, dstPath string) {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		fmt.Println(fmt.Sprintf("[Copy][Error] Unable to open file %s for copy", srcPath), err)
		return
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dstPath)
	if err != nil {
		fmt.Println(fmt.Sprintf("[Copy][Error] Failed to create file %s for copy", dstPath), err)
		return
	}
	defer dstFile.Close()

	bytesCopied, err := io.Copy(dstFile, srcFile)
	if err != nil {
		fmt.Println(fmt.Sprintf("[Copy][Error] Failed to copy file from %s to %s", srcPath, dstPath), err)
		return
	}
	fmt.Printf("[Copy] Copied %d bytes from %s to %s\n", bytesCopied, srcPath, dstPath)
}
