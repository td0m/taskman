package main

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/td0m/taskman/pkg/persist"
	"github.com/td0m/taskman/pkg/task"
)

func main() {
	years := 10
	perDay := 30
	total := 365 * perDay * years
	file := path.Join(os.TempDir(), "tasks.json")
	p, err := persist.InJSON(file)
	check(err)
	s := task.NewStore()
	for i := 0; i < total; i++ {
		id := task.RandomID()
		check(s.Create(id, time.Now()))
		check(s.Rename(id, strings.Repeat("-", 30)))
	}
	writeTime := measureTime(func() {
		err := p.Save(s)
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
