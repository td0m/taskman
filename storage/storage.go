package storage

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/td0m/taskman/task"
)

type JSONBackend struct {
	file string
}

func NewJSON(file string) *JSONBackend {
	return &JSONBackend{
		file: file,
	}
}

func (b JSONBackend) Sync(tasks task.Tasks) (task.Tasks, error) {
	f, err := os.OpenFile(b.file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return tasks, err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return tasks, enc.Encode(tasks)
}

func (b JSONBackend) Fetch() (task.Tasks, error) {
	f, err := b.open()
	if errors.Is(err, os.ErrNotExist) {
		f, err := os.Create(b.file)
		if err != nil {
			return task.NewTasks(), err
		}
		defer f.Close()
		tasks := task.NewTasks()
		return tasks, json.NewEncoder(f).Encode(tasks)
	}
	if err != nil {
		return task.Tasks{}, err
	}
	defer f.Close()
	var tasks task.Tasks
	err = json.NewDecoder(f).Decode(&tasks)
	return tasks, err
}

func (b JSONBackend) open() (*os.File, error) {
	w, err := os.OpenFile(b.file, os.O_RDWR, 0600)
	return w, err
}
