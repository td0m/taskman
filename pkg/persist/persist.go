package persist

import (
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
func (j JSON) Save(store task.StoreManager) error {
	bs, err := store.MarshalJSON()
	if err != nil {
		return err
	}
	if err := os.WriteFile(j.file, bs, 0660); err != nil {
		return err
	}
	return nil
}

// Load loads and validates tasks in a json file
func (j JSON) Load() (*task.Store, error) {
	bs, err := os.ReadFile(j.file)
	if err != nil {
		return nil, err
	}
	s := &task.Store{}
	if err := s.UnmarshalJSON(bs); err != nil {
		return nil, err
	}
	if err = s.Check(); err != nil {
		return nil, err
	}
	return s, nil
}
