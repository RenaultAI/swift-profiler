package main

import (
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.robot.car/cruise/swift-profiler/copier"
)

const defaultGoroutineCount = 16
const defaultDurationInSecond = 5
const defaultInputDirectory = "/tmp/benchmark-test"
const defaultDestinationContainer = "benchmark-test"

func copyFile(srcPath, destPath string) (int64, error) {
	in, err := os.Open(srcPath)
	if err != nil {
		return 0, err
	}
	defer in.Close()

	out, err := os.Create(destPath)
	if err != nil {
		return 0, err
	}
	defer out.Close()

	size, err := io.Copy(out, in)
	if err != nil {
		return 0, err
	}

	return size, nil
}

func main() {
	var goroutineCount int
	var inputDirectory, destinationContainer string

	flag.IntVar(&goroutineCount, "concurrency", defaultGoroutineCount, "Number of goroutines")
	flag.StringVar(&inputDirectory, "input-dir", defaultInputDirectory, "Input directory")
	flag.StringVar(&destinationContainer, "dest-prefix", defaultDestinationContainer, "Destination Swift container name")
	flag.Parse()

	files, err := ioutil.ReadDir(inputDirectory)
	if err != nil {
		log.Fatal(err)
	}

	swiftClient := copier.NewSwiftCopier()
	if err := swiftClient.Setup(); err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	fileChannel := make(chan string)

	log.Printf("Spawning %d goroutines to run\n", goroutineCount)
	for i := 0; i < goroutineCount; i++ {
		wg.Add(1)
		go func() {
			for path := range fileChannel {
				if err := swiftClient.Copy(path, destinationContainer); err != nil {
					log.Printf("Swift copy error: %s\n", err)
				}
				// log.Printf(path)
			}
			wg.Done()
		}()
	}

	log.Printf("Copying %d files concurrently\n", len(files))
	byteCount, fileCount := int64(0), 0
	start := time.Now()
	for _, file := range files {
		fileChannel <- filepath.Join(inputDirectory, file.Name())
		byteCount += file.Size()
		fileCount++
	}
	close(fileChannel)
	wg.Wait()
	duration := time.Since(start)

	log.Printf("  %d files %s written in %s\n", fileCount, humanize.Bytes(uint64(byteCount)), duration)
	log.Printf("  Copy throughput per second: %s\n", humanize.Bytes(uint64(float64(byteCount)/duration.Seconds())))
}
