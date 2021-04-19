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

	deadline := task.Task{Info: task.Info{ID: "deadline"}}
	goal1 := task.Task{Info: task.Info{ID: "goal1"}}
	goal2 := task.Task{Info: task.Info{ID: "goal2"}}
	goals := task.Task{Info: task.Info{ID: "goals"}, Children: []*task.Task{&goal1, &goal2}}
	root := task.Task{Info: task.Info{ID: "root"}, Children: []*task.Task{&goals, &deadline}}
	goal1.Parent = &goals
	goal2.Parent = &goals
	goals.Parent = &root
	deadline.Parent = &root
	tasks := []task.Task{root, goals, goal1, goal2, deadline}

	json := InJSON(path.Join(os.TempDir(), "tasks.json"))
	is.NoErr(json.Save(tasks))

	tasks2, err := json.Load()
	is.NoErr(err)
	is.Equal(len(tasks2), len(tasks))

	root2 := &tasks2[0]
	for root2.Parent != nil {
		root2 = root2.Parent
	}
	is.Equal(dfs(root), dfs(*root2))
}

// dfs is a depth-first-search traversal utility
// it is used to compare trees
func dfs(t task.Task) []task.ID {
	out := []task.ID{t.ID}
	for _, child := range t.Children {
		out = append(out, dfs(*child)...)
	}
	return out
}
