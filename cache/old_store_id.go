package cache

import (
	"slices"

	"github.com/zeroibot/fn/dict"
	"github.com/zeroibot/fn/list"
	"github.com/zeroibot/rdb/ze"
)

type idcodeable interface {
	codeable
	GetID() ze.ID
}

// T is expected to be a reference type
type IDStore[T idcodeable] struct {
	*Store[T]
	idMap  *dict.SyncMap[ze.ID, T]
	idCode *dict.SyncMap[ze.ID, string]
	codeID *dict.SyncMap[string, ze.ID]
}

// Create new IDStore
func NewIDStore[T idcodeable]() *IDStore[T] {
	return &IDStore[T]{
		Store:  NewStore[T](),
		idMap:  dict.NewSyncMap[ze.ID, T](),
		idCode: dict.NewSyncMap[ze.ID, string](),
		codeID: dict.NewSyncMap[string, ze.ID](),
	}
}

// Get item by ID
func (s *IDStore[T]) GetByID(id ze.ID) (T, bool) {
	if !useCache {
		var t T
		return t, false
	}
	return s.idMap.Get(id)
}

// Get items by IDs, no guarantee on order
func (s *IDStore[T]) GetByIDs(ids ...ze.ID) []T {
	if !useCache {
		return nil
	}
	allItems := s.All()
	return list.Filter(allItems, func(item T) bool {
		return slices.Contains(ids, item.GetID())
	})
}

// Add items to IDStore
func (s *IDStore[T]) AddItems(items []T) {
	if !useCache {
		return
	}
	for _, item := range items {
		s.Add(item)
	}
}

// Add item to IDStore
func (s *IDStore[T]) Add(item T) {
	if !useCache {
		return
	}
	s.Store.Add(item) // Add in codeMap
	id := item.GetID()
	code := item.GetCode()
	s.idMap.Set(id, item)
	s.idCode.Set(id, code)
	s.codeID.Set(code, id)
}

// Update item in IDStore
func (s *IDStore[T]) Update(item T) {
	if !useCache {
		return
	}
	s.Store.Update(item) // Add in codeMap
	s.idMap.Set(item.GetID(), item)
}

// Toggle item in IDStore by code
func (s *IDStore[T]) ToggleByCode(code string, isActive bool) {
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

// Toggle item in IDStore by ID
func (s *IDStore[T]) ToggleByID(id ze.ID, isActive bool) {
	if !useCache {
		return
	}
	item, ok := s.GetByID(id)
	if !ok {
		return
	}
	item.SetIsActive(isActive)
	s.Update(item)
}

// Delete item in IDStore by code
func (s *IDStore[T]) DeleteByCode(code string) {
	if !useCache {
		return
	}
	s.Store.DeleteByCode(code)
	if id, ok := s.codeID.Get(code); ok {
		s.idMap.Delete(id)
	}
}

// Delete item in IDStore by ID
func (s *IDStore[T]) DeleteByID(id ze.ID) {
	if !useCache {
		return
	}
	s.idMap.Delete(id)
	if code, ok := s.idCode.Get(id); ok {
		s.Store.DeleteByCode(code)
	}
}

// Return ID => Code lookup
func (s *IDStore[T]) IDCodeLookup() map[ze.ID]string {
	if !useCache {
		return nil
	}
	lookup := s.idCode.Map()
	if len(lookup) == 0 {
		return nil
	}
	return lookup
}
