package actor

import (
	"sync"

	"github.com/mir2go/mir2/pkg/protocol"
)

type StorageItem struct {
	Item      *TItem
	Index     int
	StoredTime int64
}

type TStorage struct {
	OwnerID    int32
	Items      []*StorageItem
	Gold       int32
	MaxItems   int
	Mutex      sync.RWMutex
}

func NewStorage(ownerID int32) *TStorage {
	return &TStorage{
		OwnerID:  ownerID,
		Items:    make([]*StorageItem, 0, protocol.MAXSTORAGEITEM),
		Gold:     0,
		MaxItems: protocol.MAXSTORAGEITEM,
	}
}

func (s *TStorage) AddItem(item *TItem) bool {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	
	if len(s.Items) >= s.MaxItems {
		return false
	}
	
	for i := 0; i < s.MaxItems; i++ {
		if !s.hasItemAtIndex(i) {
			storageItem := &StorageItem{
				Item: item,
				Index: i,
			}
			s.Items = append(s.Items, storageItem)
			return true
		}
	}
	
	return false
}

func (s *TStorage) hasItemAtIndex(index int) bool {
	for _, item := range s.Items {
		if item.Index == index {
			return true
		}
	}
	return false
}

func (s *TStorage) RemoveItem(makeIndex int32) *TItem {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	
	for i, item := range s.Items {
		if item.Item.MakeIndex == makeIndex {
			s.Items = append(s.Items[:i], s.Items[i+1:]...)
			return item.Item
		}
	}
	return nil
}

func (s *TStorage) GetItem(makeIndex int32) *TItem {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	
	for _, item := range s.Items {
		if item.Item.MakeIndex == makeIndex {
			return item.Item
		}
	}
	return nil
}

func (s *TStorage) GetAllItems() []*TItem {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	
	items := make([]*TItem, len(s.Items))
	for i, si := range s.Items {
		items[i] = si.Item
	}
	return items
}

func (s *TStorage) GetItemCount() int {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	return len(s.Items)
}

func (s *TStorage) IsFull() bool {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	return len(s.Items) >= s.MaxItems
}

func (s *TStorage) AddGold(gold int32) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	s.Gold += gold
}

func (s *TStorage) RemoveGold(gold int32) bool {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	
	if s.Gold < gold {
		return false
	}
	
	s.Gold -= gold
	return true
}

func (s *TStorage) GetGold() int32 {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	return s.Gold
}

func (s *TStorage) GetFreeSlotCount() int {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	return s.MaxItems - len(s.Items)
}

type StorageManager struct {
	Storages  map[int32]*TStorage
	Mutex     sync.RWMutex
}

var DefaultStorageManager *StorageManager

func init() {
	DefaultStorageManager = NewStorageManager()
}

func NewStorageManager() *StorageManager {
	return &StorageManager{
		Storages: make(map[int32]*TStorage),
	}
}

func (sm *StorageManager) GetStorage(ownerID int32) *TStorage {
	sm.Mutex.RLock()
	defer sm.Mutex.RUnlock()
	return sm.Storages[ownerID]
}

func (sm *StorageManager) CreateStorage(ownerID int32) *TStorage {
	sm.Mutex.Lock()
	defer sm.Mutex.Unlock()
	
	if storage, ok := sm.Storages[ownerID]; ok {
		return storage
	}
	
	storage := NewStorage(ownerID)
	sm.Storages[ownerID] = storage
	return storage
}

func (sm *StorageManager) DeleteStorage(ownerID int32) {
	sm.Mutex.Lock()
	defer sm.Mutex.Unlock()
	delete(sm.Storages, ownerID)
}

func (sm *StorageManager) SaveStorage(storage *TStorage) error {
	return nil
}

func (sm *StorageManager) LoadStorage(ownerID int32) *TStorage {
	sm.Mutex.RLock()
	if storage, ok := sm.Storages[ownerID]; ok {
		sm.Mutex.RUnlock()
		return storage
	}
	sm.Mutex.RUnlock()
	
	storage := sm.CreateStorage(ownerID)
	return storage
}

func (p *Player) OpenStorage() *TStorage {
	return DefaultStorageManager.GetStorage(p.CharID)
}

func (p *Player) SaveToStorage(item *TItem) bool {
	storage := DefaultStorageManager.GetStorage(p.CharID)
	if storage == nil {
		storage = DefaultStorageManager.CreateStorage(p.CharID)
	}
	
	if !storage.AddItem(item) {
		return false
	}
	
	return true
}

func (p *Player) TakeFromStorage(makeIndex int32) *TItem {
	storage := DefaultStorageManager.GetStorage(p.CharID)
	if storage == nil {
		return nil
	}
	
	item := storage.RemoveItem(makeIndex)
	return item
}

func (p *Player) GetStorageGold() int32 {
	storage := DefaultStorageManager.GetStorage(p.CharID)
	if storage == nil {
		return 0
	}
	return storage.GetGold()
}

func (p *Player) AddStorageGold(gold int32) {
	storage := DefaultStorageManager.GetStorage(p.CharID)
	if storage == nil {
		storage = DefaultStorageManager.CreateStorage(p.CharID)
	}
	storage.AddGold(gold)
}

func (p *Player) RemoveStorageGold(gold int32) bool {
	storage := DefaultStorageManager.GetStorage(p.CharID)
	if storage == nil {
		return false
	}
	return storage.RemoveGold(gold)
}

type BigStorage struct {
	OwnerID    int32
	Items      []*TItem
	MaxItems   int
	Mutex      sync.RWMutex
}

func NewBigStorage(ownerID int32) *BigStorage {
	return &BigStorage{
		OwnerID:  ownerID,
		Items:    make([]*TItem, 0),
		MaxItems: 100,
	}
}

func (bs *BigStorage) AddItem(item *TItem) bool {
	bs.Mutex.Lock()
	defer bs.Mutex.Unlock()
	
	if bs.MaxItems > 0 && len(bs.Items) >= bs.MaxItems {
		return false
	}
	
	bs.Items = append(bs.Items, item)
	return true
}

func (bs *BigStorage) RemoveItem(makeIndex int32) *TItem {
	bs.Mutex.Lock()
	defer bs.Mutex.Unlock()
	
	for i, item := range bs.Items {
		if item.MakeIndex == makeIndex {
			bs.Items = append(bs.Items[:i], bs.Items[i+1:]...)
			return item
		}
	}
	return nil
}

func (bs *BigStorage) GetAllItems() []*TItem {
	bs.Mutex.RLock()
	defer bs.Mutex.RUnlock()
	
	result := make([]*TItem, len(bs.Items))
	copy(result, bs.Items)
	return result
}

func (bs *BigStorage) GetItemCount() int {
	bs.Mutex.RLock()
	defer bs.Mutex.RUnlock()
	return len(bs.Items)
}
