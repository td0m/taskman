package main

import (
	"fmt"
	"math/rand"
	"os"
	"path"
	"time"

	"github.com/td0m/taskman/pkg/persist"
	"github.com/td0m/taskman/pkg/task"
)

func main() {
	years := 10
	perDay := 30
	total := 365 * perDay * years
	file := path.Join(os.TempDir(), "tasks.json")
	p := persist.InJSON(file)
	tasks := make([]task.Task, total)
	for i := range tasks {
		t := task.Task{
			Info: task.Info{
				ID:          task.ID(randomString(10)),
				DoneHistory: []time.Time{time.Now()},
				ClockIns: []task.ClockIn{
					{Start: time.Now(), End: time.Now()},
				},
			},
		}
		if i > 0 {
			t.Parent = &tasks[0]
		}
		tasks[i] = t
	}
	writeTime := measureTime(func() {
		err := p.Save(tasks)
		check(err)
	})

	readTime := measureTime(func() {
		_, err := p.Load()
		check(err)
	})

	info, err := os.Stat(file)
	check(err)
	fmt.Printf("Tasks: %d years, %d per day (%d total)\n", years, perDay, total)
	fmt.Printf("File size: %dMB\n", info.Size()/1024/1024)
	fmt.Printf("Write time: %dms\n", writeTime.Milliseconds())
	fmt.Printf("Read time: %dms\n", readTime.Milliseconds())
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func measureTime(fn func()) time.Duration {
	start := time.Now()
	fn()
	return time.Since(start)
}
func init() {
	rand.Seed(time.Now().UnixNano())
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomString(l int) string {
	b := make([]byte, l)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
