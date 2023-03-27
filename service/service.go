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
    if s.isAllDone(t) == completeFilter {
      filtered = append(filtered, t)
    }
  }
  return filtered
}

func (s *Service) isAllDone(item repo.Todo) bool {
  if len(item.Children) == 0 {
    return item.Done
  }

  for _, child := range item.Children {
    if !s.isAllDone(child) {
      return false
    }
  }

  return true
}

func (s *Service) AddTodo(afterItem *repo.Todo, name string) {
  t := repo.Todo{
    Id: uuid.New().String(),
    Name: name,
  }

  if afterItem == nil {
    s.repo.Todos = append([]repo.Todo{t}, s.repo.Todos...)
  } else {
    children := s.repo.Todos
    parent := s.findParent(afterItem.Id, nil)
    if parent != nil {
      children = parent.Children
    }

    index := -1 
    for i, item := range children {
      if item.Id == afterItem.Id {
        index = i + 1
      }
    }
    if index == -1 {
      if parent == nil {
        s.repo.Todos = append([]repo.Todo{t}, s.repo.Todos...)
      } else {
        parent.Children = append([]repo.Todo{t}, parent.Children...)
      }
    } else if index >= len(children) {
      if parent == nil {
        s.repo.Todos = append(s.repo.Todos, t)
      } else {
        parent.Children = append(parent.Children, t)
      }
    } else {
      if parent == nil {
        s.repo.Todos = append(s.repo.Todos[:index+1], s.repo.Todos[index:]...)
        s.repo.Todos[index] = t
      } else {
        parent.Children = append(parent.Children[:index+1], parent.Children[index:]...)
        parent.Children[index] = t
      }
    }
  }

  s.repo.Persist()
}

func (s *Service) findParent(itemId string, currentParent *repo.Todo) *repo.Todo {
  children := s.repo.Todos
  if currentParent != nil {
    children = currentParent.Children
  }

  for i, child := range children {
    if child.Id == itemId {
      return currentParent
    }
    parent := s.findParent(itemId, &children[i])
    if parent != nil {
      return parent
    }
  }
  return nil
}

func (s *Service) ToggleTodo(item repo.Todo) {
  s.toggleTodoFromSlice(item, s.repo.Todos)
}

func (s *Service) toggleTodoFromSlice(item repo.Todo, scope []repo.Todo) (bool) {
  for i, t := range scope {
    if t.Id == item.Id {
      scope[i].Done = !t.Done
      s.repo.Persist()
      return true
    } else {
      done := s.toggleTodoFromSlice(item, t.Children)
      if done {
        return done
      }
    }
  }
  return false
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
