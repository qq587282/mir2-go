package pet

import (
	"sync"
	"time"

	"github.com/mir2go/mir2/pkg/protocol"
)

type PetType int

const (
	PetTypeNormal PetType = iota
	PetTypeMount
	PetTypeGuard
	PetTypeSummon
)

type PetState int

const (
	PetStateIdle PetState = iota
	PetStateFollow
	PetStateAttack
	PetStateRest
	PetStateDead
)

type TPet struct {
	PetID   int32
	PetType PetType
	Name    string
	OwnerID int32
	Level   int
	Exp     uint64

	HP    int32
	MaxHP int32
	MP    int32
	MaxMP int32

	AC  int32
	MAC int32
	DC  int32
	MC  int32
	SC  int32

	Speed     int
	MoveSpeed int

	Appearance uint16

	State       PetState
	FollowDist  int
	AttackRange int

	SkillList []uint16

	BirthTime    time.Time
	LastFeedTime time.Time
	Loyality     int
	Hungry       int

	Locked   bool
	AutoFeed bool
}

type PetManager struct {
	Pets        map[int32]*TPet
	PetsByOwner map[int32][]*TPet
	Mutex       sync.RWMutex
	NextID      int32
}

var DefaultPetManager *PetManager

func init() {
	DefaultPetManager = NewPetManager()
}

func NewPetManager() *PetManager {
	return &PetManager{
		Pets:        make(map[int32]*TPet),
		PetsByOwner: make(map[int32][]*TPet),
		NextID:      30000,
	}
}

func (pm *PetManager) GetNextID() int32 {
	id := pm.NextID
	pm.NextID++
	return id
}

func (pm *PetManager) CreatePet(ownerID int32, name string, petType PetType) *TPet {
	pm.Mutex.Lock()
	defer pm.Mutex.Unlock()

	pet := &TPet{
		PetID:       pm.GetNextID(),
		PetType:     petType,
		Name:        name,
		OwnerID:     ownerID,
		Level:       1,
		State:       PetStateFollow,
		FollowDist:  3,
		AttackRange: 1,
		BirthTime:   time.Now(),
		Loyality:    100,
		Hungry:      100,
		AutoFeed:    true,
		MaxHP:       100,
		MaxMP:       50,
		Speed:       500,
		MoveSpeed:   350,
	}

	pm.Pets[pet.PetID] = pet

	if _, ok := pm.PetsByOwner[ownerID]; !ok {
		pm.PetsByOwner[ownerID] = make([]*TPet, 0)
	}
	pm.PetsByOwner[ownerID] = append(pm.PetsByOwner[ownerID], pet)

	return pet
}

func (pm *PetManager) DeletePet(petID int32) bool {
	pm.Mutex.Lock()
	defer pm.Mutex.Unlock()

	pet, ok := pm.Pets[petID]
	if !ok {
		return false
	}

	ownerID := pet.OwnerID
	delete(pm.Pets, petID)

	if pets, ok := pm.PetsByOwner[ownerID]; ok {
		for i, p := range pets {
			if p.PetID == petID {
				pm.PetsByOwner[ownerID] = append(pets[:i], pets[i+1:]...)
				break
			}
		}
	}

	return true
}

func (pm *PetManager) GetPet(petID int32) *TPet {
	pm.Mutex.RLock()
	defer pm.Mutex.RUnlock()
	return pm.Pets[petID]
}

func (pm *PetManager) GetOwnerPets(ownerID int32) []*TPet {
	pm.Mutex.RLock()
	defer pm.Mutex.RUnlock()

	if pets, ok := pm.PetsByOwner[ownerID]; ok {
		result := make([]*TPet, len(pets))
		copy(result, pets)
		return result
	}
	return nil
}

func (pm *PetManager) GetPetCount(ownerID int32) int {
	pm.Mutex.RLock()
	defer pm.Mutex.RUnlock()

	if pets, ok := pm.PetsByOwner[ownerID]; ok {
		return len(pets)
	}
	return 0
}

func (pm *PetManager) SetPetState(petID int32, state PetState) bool {
	pm.Mutex.Lock()
	defer pm.Mutex.Unlock()

	pet, ok := pm.Pets[petID]
	if !ok {
		return false
	}

	pet.State = state
	return true
}

func (pm *PetManager) FollowTarget(petID int32, targetX, targetY int) {
	pm.Mutex.RLock()
	pet := pm.Pets[petID]
	pm.Mutex.RUnlock()

	if pet == nil || pet.State != PetStateFollow {
		return
	}
}

func (pm *PetManager) AttackTarget(petID int32, targetID int32) {
	pm.Mutex.RLock()
	pet := pm.Pets[petID]
	pm.Mutex.RUnlock()

	if pet == nil || pet.State != PetStateAttack {
		return
	}
}

func (pm *PetManager) FeedPet(petID int32, foodType int) bool {
	pm.Mutex.Lock()
	defer pm.Mutex.Unlock()

	pet, ok := pm.Pets[petID]
	if !ok {
		return false
	}

	pet.LastFeedTime = time.Now()
	pet.Hungry = 100

	switch foodType {
	case 1:
		pet.Loyality = min(pet.Loyality+5, 100)
	case 2:
		pet.Loyality = min(pet.Loyality+10, 100)
	case 3:
		pet.Loyality = min(pet.Loyality+20, 100)
	}

	return true
}

func (pm *PetManager) RenamePet(petID int32, newName string) bool {
	pm.Mutex.Lock()
	defer pm.Mutex.Unlock()

	pet, ok := pm.Pets[petID]
	if !ok {
		return false
	}

	pet.Name = newName
	return true
}

func (pm *PetManager) LockPet(petID int32) bool {
	pm.Mutex.Lock()
	defer pm.Mutex.Unlock()

	pet, ok := pm.Pets[petID]
	if !ok {
		return false
	}

	pet.Locked = true
	return true
}

func (pm *PetManager) UnlockPet(petID int32) bool {
	pm.Mutex.Lock()
	defer pm.Mutex.Unlock()

	pet, ok := pm.Pets[petID]
	if !ok {
		return false
	}

	pet.Locked = false
	return true
}

func (pm *PetManager) AddExp(petID int32, amount uint64) bool {
	pm.Mutex.Lock()
	defer pm.Mutex.Unlock()

	pet, ok := pm.Pets[petID]
	if !ok {
		return false
	}

	pet.Exp += amount

	needExp := uint64(100 * pet.Level * pet.Level)
	if pet.Exp >= needExp {
		pet.Level++
		pet.Exp = 0
		pet.MaxHP += 20
		pet.MaxMP += 10
		pet.DC += 2
		pet.HP = pet.MaxHP
		pet.MP = pet.MaxMP
		return true
	}
	return false
}

func (pm *PetManager) ProcessAll() {
	pm.Mutex.RLock()
	pets := make([]*TPet, 0, len(pm.Pets))
	for _, p := range pm.Pets {
		pets = append(pets, p)
	}
	pm.Mutex.RUnlock()

	for _, pet := range pets {
		pet.Process()
	}
}

func (pet *TPet) Process() {
	if !pet.Alive() {
		return
	}

	if pet.Hungry > 0 {
		pet.Hungry--
	}

	if pet.Hungry < 30 && pet.AutoFeed {
		pet.State = PetStateRest
	}

	if pet.Loyality > 0 {
		pet.Loyality--
	}
}

func (pet *TPet) Alive() bool {
	return pet.HP > 0 && pet.State != PetStateDead
}

func (pet *TPet) GetFeature() int32 {
	feature := int32(pet.Appearance)
	return feature
}

func (pet *TPet) GetCharDesc() *protocol.TCharDesc {
	return &protocol.TCharDesc{
		Feature:   pet.GetFeature(),
		Status:    0,
		Level:     int32(pet.Level),
		HP:        pet.HP,
		MaxHP:     pet.MaxHP,
		AddStatus: 0,
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func GetPetManager() *PetManager {
	return DefaultPetManager
}

type MountManager struct {
	Mounts map[int32]*TPet
	Mutex  sync.RWMutex
}

var DefaultMountManager *MountManager

func init() {
	DefaultMountManager = NewMountManager()
}

func NewMountManager() *MountManager {
	return &MountManager{
		Mounts: make(map[int32]*TPet),
	}
}

func (mm *MountManager) MountPet(playerID, petID int32) bool {
	pet := DefaultPetManager.GetPet(petID)
	if pet == nil || pet.OwnerID != playerID {
		return false
	}

	if pet.PetType != PetTypeMount {
		return false
	}

	pet.State = PetStateFollow
	return true
}

func (mm *MountManager) UnmountPet(playerID, petID int32) bool {
	pet := DefaultPetManager.GetPet(petID)
	if pet == nil || pet.OwnerID != playerID {
		return false
	}

	pet.State = PetStateIdle
	return true
}

func (mm *MountManager) GetMountedPet(playerID int32) *TPet {
	pets := DefaultPetManager.GetOwnerPets(playerID)
	for _, pet := range pets {
		if pet.PetType == PetTypeMount && pet.State == PetStateFollow {
			return pet
		}
	}
	return nil
}

func GetMountManager() *MountManager {
	return DefaultMountManager
}
