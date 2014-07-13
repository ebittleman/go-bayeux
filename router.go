package bayeux

import (
	"container/list"
)

type router struct {
	*list.List
}

type Router interface {
	// HandleFunc(string, BayeuxHandler, map[string]interface{})
	// Exec(string)
	RouteChain
}

type RouteChain interface {
	Back() *list.Element
	Front() *list.Element
	Init() *list.List
	InsertAfter(interface{}, *list.Element) *list.Element
	InsertBefore(interface{}, *list.Element) *list.Element
	Len() int
	MoveToBack(*list.Element)
	MoveToFront(*list.Element)
	PushBack(interface{}) *list.Element
	PushBackList(*list.List)
	PushFront(interface{}) *list.Element
	PushFrontList(*list.List)
	Remove(*list.Element) interface{}
}

func NewRouter() Router {
	return &router{list.New()}
}
