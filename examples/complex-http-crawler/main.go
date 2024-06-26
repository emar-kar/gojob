package main

import (
	"os"

	"github.com/WangYihang/gojob"
	"github.com/WangYihang/gojob/examples/complex-http-crawler/pkg/model"
	"github.com/WangYihang/gojob/pkg/utils"
	"github.com/WangYihang/gojob/pkg/version"
	"github.com/jessevdk/go-flags"
)

type Options struct {
	InputFilePath            string `short:"i" long:"input" description:"input file path" required:"true"`
	OutputFilePath           string `short:"o" long:"output" description:"output file path" required:"true"`
	MaxRetries               int    `short:"r" long:"max-retries" description:"max retries" default:"3"`
	MaxRuntimePerTaskSeconds int    `short:"t" long:"max-runtime-per-task-seconds" description:"max runtime per task seconds" default:"60"`
	NumWorkers               int    `short:"n" long:"num-workers" description:"number of workers" default:"32"`
	NumShards                int    `short:"s" long:"num-shards" description:"number of shards" default:"1"`
	Shard                    int    `short:"d" long:"shard" description:"shard" default:"0"`
	Version                  func() `long:"version" description:"print version and exit" json:"-"`
}

var opts Options

func init() {
	opts.Version = version.PrintVersion
	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}
}

func main() {
	total := utils.Count(utils.Cat(opts.InputFilePath))
	scheduler := gojob.New(
		gojob.WithNumWorkers(opts.NumWorkers),
		gojob.WithMaxRetries(opts.MaxRetries),
		gojob.WithMaxRuntimePerTaskSeconds(opts.MaxRuntimePerTaskSeconds),
		gojob.WithNumShards(int64(opts.NumShards)),
		gojob.WithShard(int64(opts.Shard)),
		gojob.WithResultFilePath(opts.OutputFilePath),
		gojob.WithTotalTasks(total),
		gojob.WithStatusFilePath("status.json"),
		gojob.WithResultFilePath("result.json"),
		gojob.WithMetadataFilePath("metadata.json"),
		gojob.WithPrometheusPushGateway("http://localhost:9091", "gojob"),
	).
		Start()
	for line := range utils.Cat(opts.InputFilePath) {
		scheduler.Submit(model.New(string(line)))
	}
	scheduler.Wait()
}
