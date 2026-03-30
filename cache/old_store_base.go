package cache

import (
	"slices"

	"github.com/zeroibot/fn/dict"
	"github.com/zeroibot/fn/list"
)

type codeable interface {
	GetCode() string
	SetIsActive(bool)
}

// T is expected to be a reference type
type Store[T codeable] struct {
	codeMap *dict.SyncMap[string, T]
}

// Create new Store
func NewStore[T codeable]() *Store[T] {
	return &Store[T]{
		codeMap: dict.NewSyncMap[string, T](),
	}
}

// Gets all stored objects
func (s *Store[T]) All() []T {
	if !useCache {
		return nil
	}
	items := s.codeMap.Values()
	if len(items) == 0 {
		return nil
	}
	return items
}

// Get item by code
func (s *Store[T]) GetByCode(code string) (T, bool) {
	if !useCache {
		var t T
		return t, false
	}
	return s.codeMap.Get(code)
}

// Get items by codes, no guarantee on order
func (s *Store[T]) GetByCodes(codes ...string) []T {
	if !useCache {
		return nil
	}
	allItems := s.All()
	return list.Filter(allItems, func(item T) bool {
		return slices.Contains(codes, item.GetCode())
	})
}

// Add items to store
func (s *Store[T]) AddItems(items []T) {
	if !useCache {
		return
	}
	for _, item := range items {
		s.Add(item)
	}
}

// Add item to store
func (s *Store[T]) Add(item T) {
	if !useCache {
		return
	}
	s.codeMap.Set(item.GetCode(), item)
}

// Update item in store
func (s *Store[T]) Update(item T) {
	if !useCache {
		return
	}
	s.codeMap.Set(item.GetCode(), item)
}

// Toggle item in store by code
func (s *Store[T]) ToggleByCode(code string, isActive bool) {
	if !useCache {
		return
	}
	item, ok := s.GetByCode(code)
	if !ok {
		return
	}
	item.SetIsActive(isActive)
	s.Update(item)
}

// Delete item in store by code
func (s *Store[T]) DeleteByCode(code string) {
	if !useCache {
		return
	}
	s.codeMap.Delete(code)
}
