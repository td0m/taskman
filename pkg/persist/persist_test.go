package persist

import (
	"os"
	"path"
	"testing"

	"github.com/matryer/is"
	"github.com/td0m/taskman/pkg/task"
)

func TestJSON_SaveLoad(t *testing.T) {
	is := is.New(t)

	deadline := task.Info{}
	goal1 := task.Info{}
	goal2 := task.Info{}
	goals := task.Info{}
	root := task.Info{}
	store := &task.Store{}
	store.Nodes = map[task.ID]task.Info{
		"root":     root,
		"goals":    goals,
		"goal1":    goal1,
		"goal2":    goal2,
		"deadline": deadline,
	}
	store.Parent = map[task.ID]task.ID{
		"goal1":    "goals",
		"goal2":    "goals",
		"goals":    "root",
		"deadline": "root",
	}
	store.Children = map[task.ID][]task.ID{
		"root":  {"goals", "deadline"},
		"goals": {"goal1", "goal2"},
	}
	tasks := store.Root()

	json := InJSON(path.Join(os.TempDir(), "tasks.json"))
	is.NoErr(json.Save(store))

	store2, err := json.Load()
	tasks2 := store2.Root()
	is.NoErr(err)
	is.Equal(len(tasks2.Children), len(tasks.Children))

	root2 := store.Root()
	is.Equal(len(dfs(root2)), 5)
	is.Equal(dfs(store.Root()), dfs(root2))
}

// dfs is a depth-first-search traversal utility
// it is used to compare trees
func dfs(t *task.Task) []task.Task {
	out := []task.Task{*t}
	for _, child := range t.Children {
		out = append(out, dfs(child)...)
	}
	return out
}
