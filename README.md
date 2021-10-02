Test benchmark
=======

Yandex search result based http get benchmarking service

### How to deploy by yourself

Set environment in `docker-compose` and run `docker-compose up -d` command to build and run service

### Supported ENV parameters

- `PORT:` The port service will be available on
- `CONCURRENCY:` The number of concurrent gorouitine workers that will be used for benchmarking
- `REQUESTS_PER_HOST:` The number of requests that will run against host
- `REQUEST_TIMEOUT:` The time before request is cancelled
- `USE_COMPRESSION:` Request the use of http gzip compression
- `USE_HTTP_GET:` Use http get. Raw requests otherwise
- `CHUNK_SIZE:` The size of processing chunk

### Service API

Currently, service support only one method `/sites?search=foobar`
