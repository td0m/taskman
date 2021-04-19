package persist

import (
	"encoding/json"
	"os"

	"github.com/td0m/taskman/pkg/task"
)

type Persistor interface {
	Save([]task.Task) error
	Load() ([]task.Task, error)
}

type JSON struct {
	file string
}

func InJSON(file string) *JSON {
	return &JSON{file}
}

// Save saves a list of tasks to a json file
func (j JSON) Save(ts []task.Task) error {
	data, err := newSavable(ts)
	if err != nil {
		return err
	}
	bs, err := json.Marshal(data)
	if err != nil {
		return err
	}
	if err := os.WriteFile(j.file, bs, 0660); err != nil {
		return err
	}
	return nil
}

// Load loads and validates tasks in a json file
func (j JSON) Load() ([]task.Task, error) {
	bs, err := os.ReadFile(j.file)
	if err != nil {
		return nil, err
	}
	s, err := newSavable([]task.Task{})
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bs, &s); err != nil {
		return nil, err
	}
	tasks, err := s.Load()
	if err != nil {
		return nil, err
	}
	return tasks, nil
}
