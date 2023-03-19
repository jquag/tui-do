package service

import "github.com/jquag/tui-do/repo"

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
