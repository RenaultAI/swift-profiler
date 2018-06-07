# Swift Profiler

```sh
$ go get
$ cp env.sample .env
$ # Update .env with the actual swift credentials
$ go run main.go -concurrency=16 -input-dir=/tmp/benchmark-test -precompute-checksum=false
2018/06/06 18:14:21 Swift container metadata: map[]
2018/06/06 18:14:21 Letting Swift compute checksum...
2018/06/06 18:14:21 Spawning 16 goroutines to run
2018/06/06 18:14:21 Copying 3 files concurrently
2018/06/06 18:14:59   3 files 2.1 GB written in 38.250574358s
2018/06/06 18:14:59   Copy throughput per second: 56 MB
```
