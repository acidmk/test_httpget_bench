Test benchmark
=======

Yandex search result based http get benchmarking service

### How to deploy by yourself

Set environment in `docker-compose` and run `docker-compose up -d` command to build and run service

### Supported ENV parameters

- `PORT:` The port service will be available on
- `CONCURRENCY:` Number of concurrent gorouitine workers that will be used for benchmarking
- `REQUESTS_PER_HOST:` Number of requests that will run against host
- `REQUEST_TIMEOUT:` Timeout
- `USE_COMPRESSION:` Use http gzip compression

### Service API

Currently, service support only one method `/sites?search=foobar`
