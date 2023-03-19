package repo

import (
	"encoding/json"
	"log"
	"os"
)

type Todo struct {
  Id string
  Name string
  Done bool
  Children []Todo
}

type Repo struct {
  filename string
  Todos []Todo
}

func NewRepo(filename string) *Repo {
  return &Repo{
    filename: filename,
    Todos: loadFromFile(filename),
  }
}

func loadFromFile(filename string) []Todo {
  var payload []Todo
  content, err := os.ReadFile(filename)
  if os.IsNotExist(err) {
    content = []byte("[]")
    os.WriteFile(filename, content, 0644)
  } else if err != nil {
    log.Fatal("Error when opening file: ", err)
  }

  err = json.Unmarshal(content, &payload)
  if err != nil {
    log.Fatal("Error during Unmarshal(): ", err)
  }

  return payload
}

func (r *Repo) Persist() {
  content, _ := json.MarshalIndent(r.Todos, "", "  ")
  os.WriteFile(r.filename, content, 0644)
}
