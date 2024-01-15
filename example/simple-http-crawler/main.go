package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/WangYihang/gojob"
)

type MyTask struct {
	Url        string `json:"url"`
	StartedAt  int64  `json:"started_at"`
	FinishedAt int64  `json:"finished_at"`
	StatusCode int    `json:"status_code"`
	Error      string `json:"error"`
}

func NewTask(line []byte) *MyTask {
	t := &MyTask{}
	t.Parse(line)
	return t
}

func (t *MyTask) Parse(data []byte) (err error) {
	t.Url = string(bytes.TrimSpace(data))
	return
}

func (t *MyTask) Do() {
	t.StartedAt = time.Now().UnixMilli()
	defer func() {
		t.FinishedAt = time.Now().UnixMilli()
	}()
	response, err := http.Get(t.Url)
	if err != nil {
		t.Error = err.Error()
		return
	}
	t.StatusCode = response.StatusCode
	defer response.Body.Close()
}

func (t *MyTask) Bytes() ([]byte, error) {
	return json.Marshal(t)
}

func main() {
	scheduler := gojob.NewScheduler(16, "output.txt")
	go func() {
		for line := range gojob.Cat("input.txt") {
			scheduler.Add(NewTask(line))
		}
	}()
	scheduler.Start()
}