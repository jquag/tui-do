package service

import (

	"github.com/google/uuid"
	"github.com/jquag/tui-do/repo"
)

type Service struct {
  repo *repo.Repo
}

func NewService(r *repo.Repo) *Service {
  return &Service{repo: r}
}

func (s *Service) Todos(completeFilter bool) []repo.Todo {
  var filtered []repo.Todo

  for _, t := range s.repo.Todos {
    if t.Done == completeFilter {
      filtered = append(filtered, t)
    }
  }
  return filtered
}

func (s *Service) AddTodo(afterItem *repo.Todo, name string) {
  t := repo.Todo{
    Id: uuid.New().String(),
    Name: name,
  }

  if afterItem == nil {
    s.repo.Todos = append([]repo.Todo{t}, s.repo.Todos...)
  } else {
    index := -1 
    for i, item := range s.repo.Todos {
      if item.Id == afterItem.Id {
        index = i + 1
      }
    }
    if index == -1 {
      s.repo.Todos = append([]repo.Todo{t}, s.repo.Todos...)
    } else if index >= len(s.repo.Todos) {
      s.repo.Todos = append(s.repo.Todos, t)
    } else {
      s.repo.Todos = append(s.repo.Todos[:index+1], s.repo.Todos[index:]...)
      s.repo.Todos[index] = t
    }
  }

  s.repo.Persist()
}

func (s *Service) ToggleTodo(item repo.Todo) {
  for i, t := range s.repo.Todos {
    if t.Id == item.Id {
      s.repo.Todos[i].Done = !t.Done
      s.repo.Persist()
      break
    }
  }
}

func (s *Service) ChangeTodo(item repo.Todo, name string) {
  for i, t := range s.repo.Todos {
    if t.Id == item.Id {
      s.repo.Todos[i].Name = name
      s.repo.Persist()
      break
    }
  }
}

func (s *Service) DeleteTodo(item repo.Todo) {
  indexToDelete := -1
  for i, t := range s.repo.Todos {
    if t.Id == item.Id {
      indexToDelete = i
      break
    }
  }
  s.repo.Todos = append(s.repo.Todos[:indexToDelete], s.repo.Todos[indexToDelete+1:]...)
  s.repo.Persist()
}
