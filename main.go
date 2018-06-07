package main

import (
	"crypto/md5"
	_ "expvar"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"sync"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.robot.car/cruise/swift-profiler/copier"
)

const defaultGoroutineCount = 16
const defaultInputDirectory = "/tmp/benchmark-test"
const defaultDestinationContainer = "benchmark-test"
const defaultPrecomputeChecksum = true

func main() {
	var goroutineCount int
	var inputDirectory, destinationContainer string
	var precomputeChecksum bool

	flag.IntVar(&goroutineCount, "concurrency", defaultGoroutineCount, "Number of goroutines")
	flag.StringVar(&inputDirectory, "input-dir", defaultInputDirectory, "Input directory")
	flag.StringVar(&destinationContainer, "dest-prefix", defaultDestinationContainer, "Destination Swift container name")
	flag.BoolVar(&precomputeChecksum, "precompute-checksum", defaultPrecomputeChecksum, "Pre-compute checksum beforehand")
	flag.Parse()

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	files, err := ioutil.ReadDir(inputDirectory)
	if err != nil {
		log.Fatal(err)
	}

	swiftClient := copier.NewSwiftCopier()
	if err := swiftClient.Setup(); err != nil {
		log.Fatal(err)
	}

	// Pre-compute md5 checksums and pass them to the copier.
	checksums := make(map[string]string, len(files))
	if precomputeChecksum {
		log.Printf("Precomputing checksum...\n")
		for _, file := range files {
			path := filepath.Join(inputDirectory, file.Name())
			f, err := os.Open(path)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()

			h := md5.New()
			if _, err := io.Copy(h, f); err != nil {
				log.Fatal(err)
			}

			checksums[path] = fmt.Sprintf("%x", h.Sum(nil))
		}
	} else {
		log.Printf("Letting Swift compute checksum...\n")
	}

	var wg sync.WaitGroup
	fileChannel := make(chan string)

	log.Printf("Spawning %d goroutines to run\n", goroutineCount)
	for i := 0; i < goroutineCount; i++ {
		wg.Add(1)
		go func() {
			for path := range fileChannel {
				var md5 *string
				if precomputeChecksum {
					sum := checksums[path]
					md5 = &sum
				}
				if err := swiftClient.Copy(path, destinationContainer, md5); err != nil {
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
		path := filepath.Join(inputDirectory, file.Name())
		fileChannel <- path
		byteCount += file.Size()
		fileCount++
	}
	close(fileChannel)
	wg.Wait()
	duration := time.Since(start)

	log.Printf("  %d files %s written in %s\n", fileCount, humanize.Bytes(uint64(byteCount)), duration)
	log.Printf("  Copy throughput per second: %s\n", humanize.Bytes(uint64(float64(byteCount)/duration.Seconds())))
}
