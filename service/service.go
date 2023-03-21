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

func (s *Service) AddTodo(index int, name string) {
  t := repo.Todo{
    Id: uuid.New().String(),
    Name: name,
  }

  incomplete := s.Todos(false)
  previous := incomplete[index]

  indexInFullList := -1 
  for i, item := range s.repo.Todos {
    if item.Id == previous.Id {
      indexInFullList = i + 1
    }
  }

  if indexInFullList == -1 {
    s.repo.Todos = append([]repo.Todo{t}, s.repo.Todos...)
  } else if indexInFullList >= len(s.repo.Todos) {
    s.repo.Todos = append(s.repo.Todos, t)
  } else {
    s.repo.Todos = append(s.repo.Todos[:indexInFullList+1], s.repo.Todos[indexInFullList:]...)
    s.repo.Todos[indexInFullList] = t
  }

  s.repo.Persist()
}

//func AddTodoGroup(name string) model.TodoGroup {
//  group := model.TodoGroup{Name: name, Id: uuid.New().String()}
//  model.Inst.TodoGroups = append(model.Inst.TodoGroups, group)
//  model.Persist()
//  return group
//}

//func SetIsAddingGroup(val bool) {
//  model.Inst.IsAdding = val
//}

//func IncompleteTodoGroups() []model.TodoGroup {
//  //TODO filter out the complete ones
//  return model.Inst.TodoGroups
//}

//func SetActiveGroupId(id string) {
//  model.Inst.ActiveGroupId = id
//}

//func ActiveGroupId() string {
//  return model.Inst.ActiveGroupId
//}

//func TodoGroupAfterId(id string) *model.TodoGroup {
//  var index int
//  todoGroups := IncompleteTodoGroups()
//  for i, g := range IncompleteTodoGroups() {
//    if g.Id == id {
//      index = i
//    }
//  }
//  if len(todoGroups)-1 < index {
//    return nil
//  } else {
//    return &todoGroups[index]
//  }
//}
