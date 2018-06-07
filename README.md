# Swift Profiler

## Setup
```sh
$ go get
$ cp env.sample .env
$ # Update .env with the actual swift credentials
```

## Run Options

### Let Swift compute checksum
Worst performance.

```sh
go run main.go -concurrency=16 -input-dir=/tmp/benchmark-test -precompute-checksum=false
2018/06/07 11:40:32 Listening on 0.0.0.0:6060
2018/06/07 11:40:32 Swift container metadata: map[]
2018/06/07 11:40:32 Letting Swift compute checksum...
2018/06/07 11:40:32 Spawning 16 goroutines to run
2018/06/07 11:40:32 Copying 3 files concurrently
2018/06/07 11:41:08   3 files 2.1 GB written in 35.752444541s
2018/06/07 11:41:08   Copy throughput per second: 60 MB
```

### Pre-compute checksum
Decent performance.

```sh
go run main.go -concurrency=16 -input-dir=/tmp/benchmark-test -precompute-checksum=true
2018/06/07 13:02:09 Listening on 0.0.0.0:6060
2018/06/07 13:02:09 Swift container metadata: map[]
2018/06/07 13:02:09 Precomputing checksum...
2018/06/07 13:02:12 Spawning 16 goroutines to run
2018/06/07 13:02:12 Copying 3 files concurrently
2018/06/07 13:02:38   3 files 2.1 GB written in 25.792156232s
2018/06/07 13:02:38   Copy throughput per second: 83 MB
```

### Ignore checksums completely
Speed up the test.

```sh
go run main.go -concurrency=16 -input-dir=/tmp/benchmark-test -verify-checksum=false
2018/06/07 13:03:04 Listening on 0.0.0.0:6060
2018/06/07 13:03:04 Swift container metadata: map[]
2018/06/07 13:03:04 Ignore checksums completely...
2018/06/07 13:03:04 Spawning 16 goroutines to run
2018/06/07 13:03:04 Copying 3 files concurrently
2018/06/07 13:03:28   3 files 2.1 GB written in 24.621356122s
2018/06/07 13:03:28   Copy throughput per second: 87 MB
```

## Performance Test Results
120 files each 2GB large.

### 16 goroutines - 1.1GB/s
```
infra@r201-u33:~/ryoung-test$ ./swift-profiler -concurrency=16 -input-dir=/home/infra/testfiles -verify-checksum=false
2018/06/07 20:10:12 Swift container metadata: map[]
2018/06/07 20:10:12 Ignore checksums completely...
2018/06/07 20:10:12 Spawning 16 goroutines to run
2018/06/07 20:10:12 Copying 120 files concurrently
2018/06/07 20:14:14   120 files 258 GB written in 4m1.692408635s
2018/06/07 20:14:14   Copy throughput per second: 1.1 GB
```

### 32 goroutines - 2.2GB/s
```
infra@r201-u33:~/ryoung-test$ ./swift-profiler -concurrency=32 -input-dir=/home/infra/testfiles -verify-checksum=false
2018/06/07 20:22:45 Listening on 0.0.0.0:6060
2018/06/07 20:22:45 Swift container metadata: map[]
2018/06/07 20:22:45 Ignore checksums completely...
2018/06/07 20:22:45 Spawning 32 goroutines to run
2018/06/07 20:22:45 Copying 120 files concurrently
2018/06/07 20:24:42   120 files 258 GB written in 1m57.356396652s
2018/06/07 20:24:42   Copy throughput per second: 2.2 GB
```
