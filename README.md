# Go(od) Job

[![Go Reference](https://pkg.go.dev/badge/github.com/WangYihang/gojob.svg)](https://pkg.go.dev/github.com/WangYihang/gojob)
[![Go Report Card](https://goreportcard.com/badge/github.com/WangYihang/gojob)](https://goreportcard.com/report/github.com/WangYihang/gojob)
[![codecov](https://codecov.io/gh/WangYihang/gojob/graph/badge.svg?token=FG1HT7FCKG)](https://codecov.io/gh/WangYihang/gojob)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FWangYihang%2Fgojob.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2FWangYihang%2Fgojob?ref=badge_shield)

gojob is a simple job scheduler.

## Install

```
go get github.com/WangYihang/gojob
```

## Usage

Create a job scheduler with a worker pool of size 32. To do this, you need to implement the `Task` interface.

```go
// Task is an interface that defines a task
type Task interface {
	// Do starts the task, returns error if failed
	// If an error is returned, the task will be retried until MaxRetries
	// You can set MaxRetries by calling SetMaxRetries on the scheduler
	Do() error
}
```

The whole [code](./examples/simple-http-crawler/main.go) looks like this (try it [online](https://go.dev/play/p/UiYextGte4v)).

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/WangYihang/gojob"
)

type MyTask struct {
	Url        string `json:"url"`
	StatusCode int    `json:"status_code"`
}

func New(url string) *MyTask {
	return &MyTask{
		Url: url,
	}
}

func (t *MyTask) Do() error {
	response, err := http.Get(t.Url)
	if err != nil {
		return err
	}
	t.StatusCode = response.StatusCode
	defer response.Body.Close()
	return nil
}

func main() {
	var numTotalTasks int64 = 256
	scheduler := gojob.New(
		gojob.WithNumWorkers(8),
		gojob.WithMaxRetries(4),
		gojob.WithMaxRuntimePerTaskSeconds(16),
		gojob.WithNumShards(4),
		gojob.WithShard(0),
		gojob.WithTotalTasks(numTotalTasks),
		gojob.WithStatusFilePath("status.json"),
		gojob.WithResultFilePath("result.json"),
		gojob.WithMetadataFilePath("metadata.json"),
	).
		Start()
	for i := range numTotalTasks {
		scheduler.Submit(New(fmt.Sprintf("https://httpbin.org/task/%d", i)))
	}
	scheduler.Wait()
}
```

## Use Case

### http-crawler

Let's say you have a bunch of URLs that you want to crawl and save the HTTP response to a file. You can use gojob to do that.
Check [it](./examples/complex-http-crawler/) out for details.

Try it out using the following command.

```bash
$ go run github.com/WangYihang/gojob/examples/complex-http-crawler@latest --help
Usage:
  main [OPTIONS]

Application Options:
  -i, --input=                        input file path
  -o, --output=                       output file path
  -r, --max-retries=                  max retries (default: 3)
  -t, --max-runtime-per-task-seconds= max runtime per task seconds (default: 60)
  -n, --num-workers=                  number of workers (default: 32)

Help Options:
  -h, --help                          Show this help message
```

```bash
$ cat urls.txt
https://www.google.com/
https://www.facebook.com/
https://www.youtube.com/
```

```
$ go run github.com/WangYihang/gojob/examples/complex-http-crawler@latest -i input.txt -o output.txt -n 4
```

```json
$ tail -n 1 output.txt
{
    "started_at": 1708934911909748,
    "finished_at": 1708934913160935,
    "num_tries": 1,
    "task": {
        "url": "https://www.google.com/",
        "http": {
            "request": {
                "method": "HEAD",
                "url": "https://www.google.com/",
                "host": "www.google.com",
            	// details omitted for simplicity
            },
            "response": {
                "status": "200 OK",
                "proto": "HTTP/2.0",
                "header": {
                    "Alt-Svc": [
                        "h3=\":443\"; ma=2592000,h3-29=\":443\"; ma=2592000"
                    ],
                },
            	// details omitted for simplicity
                "body": "",
            }
        }
    },
    "error": ""
}
```

## Integration with Prometheus

gojob provides metrics (`num_total`, `num_finshed`, `num_succeed`, `num_finished`) for Prometheus. You can use the following code to expose the metrics.

```go
package main

func main() {
	scheduler := gojob.New(
		// All you need to do is just adding the following option to the scheduler constructor
		gojob.WithPrometheusPushGateway("http://localhost:9091", "gojob"),
	).Start()
}
```

## Coverage

[![codecov-graph](https://codecov.io/gh/WangYihang/gojob/graphs/tree.svg?token=FG1HT7FCKG)](https://codecov.io/gh/WangYihang/gojob)


## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FWangYihang%2Fgojob.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2FWangYihang%2Fgojob?ref=badge_large)