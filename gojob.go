package gojob

import (
	"bufio"
	"log"
	"log/slog"
	"os"
	"sync"
)

// Fanin takes a slice of channels and returns a single channel that
func Fanin[T interface{}](cs []chan T) chan T {
	var wg sync.WaitGroup
	out := make(chan T)
	output := func(c chan T) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

// Fanout takes a channel and returns a slice of channels
// the item in the input channel will be distributed to the output channels
func Fanout[T interface{}](in chan *T, n int) []chan *T {
	cs := make([]chan *T, n)
	for i := 0; i < n; i++ {
		cs[i] = make(chan *T)
		go func(c chan *T) {
			for n := range in {
				c <- n
			}
			close(c)
		}(cs[i])
	}
	return cs
}

// Head takes a channel and returns a channel with the first n items
func Head[T interface{}](in chan T, max int) chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		i := 0
		for line := range in {
			if i >= max {
				break
			}
			out <- line
			i++
		}
	}()
	return out
}

// Tail takes a channel and returns a channel with the last n items
func Tail[T interface{}](in chan T, max int) chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		var lines []T
		for line := range in {
			lines = append(lines, line)
			if len(lines) > max {
				lines = lines[1:]
			}
		}
		for _, line := range lines {
			out <- line
		}
	}()
	return out
}

// Cat takes a file path and returns a channel with the lines of the file
func Cat(filePath string) <-chan string {
	out := make(chan string)

	go func() {
		defer close(out) // Ensure the channel is closed when the goroutine finishes

		// Open the file
		file, err := os.Open(filePath)
		if err != nil {
			log.Printf("Error opening file: %v", err)
			return // Close the channel and exit the goroutine
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			out <- scanner.Text() // Send the line to the channel
		}

		// Check for errors during Scan, excluding EOF
		if err := scanner.Err(); err != nil {
			log.Printf("Error reading file: %v", err)
		}
	}()

	return out
}

// Filter takes a channel and returns a channel with the items that pass the filter
func Filter[T interface{}](in chan T, f func(T) bool) chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		for line := range in {
			if f(line) {
				out <- line
			}
		}
	}()
	return out
}

// Map takes a channel and returns a channel with the items that pass the filter
func Map[T interface{}, U interface{}](in chan T, f func(T) U) chan U {
	out := make(chan U)
	go func() {
		defer close(out)
		for line := range in {
			out <- f(line)
		}
	}()
	return out
}

// Reduce takes a channel and returns a channel with the items that pass the filter
func Reduce[T interface{}](in chan T, f func(T, T) T) T {
	var result T
	for line := range in {
		result = f(result, line)
	}
	return result
}

// Task is an interface that defines a task
type Task interface {
	// Do starts the task
	Do()
	// Bytes serializes a task to a byte array, returns an error if the task is invalid
	// For example, a task can be serialized to a line of a file
	// You can store the result of a task in the task itself, when the task is serialized, the bytes of the result will be written to the log file
	Bytes() ([]byte, error)
}

// Scheduler is a task scheduler
type Scheduler struct {
	NumWorkers     int
	OutputFilePath string
	TaskChan       chan Task
}

// NewScheduler creates a new scheduler
func NewScheduler(numWorkers int, outputFilePath string) *Scheduler {
	return &Scheduler{
		NumWorkers:     numWorkers,
		TaskChan:       make(chan Task, numWorkers),
		OutputFilePath: outputFilePath,
	}
}

// Add adds a task to the scheduler
func (s *Scheduler) Add(task Task) {
	s.TaskChan <- task
}

// Start starts the scheduler
func (s *Scheduler) Start() {
	results := []chan string{}
	for i := 0; i < s.NumWorkers; i++ {
		results = append(results, s.Worker())
	}
	s.Writer(Fanin(results), s.OutputFilePath)
}

func (s *Scheduler) Worker() chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for task := range s.TaskChan {
			task.Do()
			data, err := task.Bytes()
			if err != nil {
				slog.Error("error occured while serializing task", slog.String("error", err.Error()))
			}
			out <- string(data)
		}
	}()
	return out
}

func (s *Scheduler) Writer(lines chan string, outputFilePath string) {
	var fd *os.File
	var err error
	if outputFilePath == "-" {
		fd = os.Stdout
	} else {
		fd, err = os.OpenFile(outputFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			slog.Error("error occured while opening file", slog.String("path", outputFilePath), slog.String("error", err.Error()))
			return
		}
	}
	for result := range lines {
		fd.WriteString(result + "\n")
	}
}
