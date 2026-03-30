package cache

import (
	"slices"

	"github.com/zeroibot/fn/list"
	"github.com/zeroibot/rdb/ze"
)

type child interface {
	GetParentID() ze.ID
}

type codeableChild interface {
	codeable
	child
}

type idcodeableChild interface {
	idcodeable
	child
}

// T is expected to be a reference type
type ChildStore[T codeableChild] struct {
	*Store[T]
}

// T is expected to be a reference type
type ChildIDStore[T idcodeableChild] struct {
	*IDStore[T]
}

// Create new ChildStore
func NewChildStore[T codeableChild]() *ChildStore[T] {
	return &ChildStore[T]{
		Store: NewStore[T](),
	}
}

// Create new ChildIDStore
func NewChildIDStore[T idcodeableChild]() *ChildIDStore[T] {
	return &ChildIDStore[T]{
		IDStore: NewIDStore[T](),
	}
}

// Get items with parent IDs
func (s *ChildStore[T]) FromParentIDs(parentIDs ...ze.ID) []T {
	if !useCache || len(parentIDs) == 0 {
		return nil
	}
	items := list.Filter(s.All(), func(item T) bool {
		return slices.Contains(parentIDs, item.GetParentID())
	})
	if len(items) == 0 {
		return nil
	}
	return items
}

// Get items with parent IDs
func (s *ChildIDStore[T]) FromParentIDs(parentIDs ...ze.ID) []T {
	if !useCache || len(parentIDs) == 0 {
		return nil
	}
	items := list.Filter(s.All(), func(item T) bool {
		return slices.Contains(parentIDs, item.GetParentID())
	})
	if len(items) == 0 {
		return nil
	}
	return items
}

// Group items by parent IDs
func (s *ChildStore[T]) GroupByParentIDs(parentIDs ...ze.ID) map[ze.ID][]T {
	if !useCache || len(parentIDs) == 0 {
		return nil
	}
	groups := make(map[ze.ID][]T)
	for _, item := range s.codeMap.Values() {
		parentID := item.GetParentID()
		if !slices.Contains(parentIDs, parentID) {
			continue
		}
		groups[parentID] = append(groups[parentID], item)
	}
	return groups
}

// Group items by parent IDs
func (s *ChildIDStore[T]) GroupByParentIDs(parentIDs ...ze.ID) map[ze.ID][]T {
	if !useCache || len(parentIDs) == 0 {
		return nil
	}
	groups := make(map[ze.ID][]T)
	for _, item := range s.codeMap.Values() {
		parentID := item.GetParentID()
		if !slices.Contains(parentIDs, parentID) {
			continue
		}
		groups[parentID] = append(groups[parentID], item)
	}
	return groups
}
