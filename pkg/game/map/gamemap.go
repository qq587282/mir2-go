package gamemap

import (
	"encoding/binary"
	"fmt"
	"os"
	"sync"
)

type TMapHeader struct {
	Width       uint16
	Height      uint16
	Version     byte
	Title       [14]byte
	UpdateDate  int64
	Reserved    [25]byte
}

type TMapUnitInfo struct {
	BkImg      uint16
	MidImg     uint16
	FrImg      uint16
	DoorIndex  byte
	DoorOffset byte
	AniFrame   byte
	AniTick    byte
	Area       byte
	Light      byte
}

type TMapCellInfo struct {
	Flag    byte
	ObjList []*MapObject
}

type MapObject interface {
	GetID() int32
	GetX() int
	GetY() int
	GetName() string
}

type GameMap struct {
	Header        TMapHeader
	MapName       string
	FileName      string
	Width         int
	Height        int
	Cells         [][]*TMapCellInfo
	Objects       sync.RWMutex
	Players       map[int32]MapObject
	NPCs          map[int32]MapObject
	Monsters      map[int32]MapObject
	Events        map[int32]MapObject
	
	SAFE          bool
	DARK          bool
	FIGHT         bool
	FIGHT3        bool
	NORECONNECT   bool
	NEEDHOLE      bool
	NORECALL      bool
	NORANDOMMOVE  bool
	NODRUG        bool
	MINE          bool
	NOPOSITIONMOVE bool
}

func NewGameMap(name, filename string) *GameMap {
	return &GameMap{
		MapName: name,
		FileName: filename,
		Players:  make(map[int32]MapObject),
		NPCs:    make(map[int32]MapObject),
		Monsters: make(map[int32]MapObject),
		Events:  make(map[int32]MapObject),
	}
}

func (m *GameMap) LoadFromFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open map file: %w", err)
	}
	defer file.Close()
	
	header := TMapHeader{}
	if err := binary.Read(file, binary.LittleEndian, &header); err != nil {
		return fmt.Errorf("failed to read map header: %w", err)
	}
	
	m.Header = header
	m.Width = int(header.Width)
	m.Height = int(header.Height)
	
	m.Cells = make([][]*TMapCellInfo, m.Width)
	for i := 0; i < m.Width; i++ {
		m.Cells[i] = make([]*TMapCellInfo, m.Height)
		for j := 0; j < m.Height; j++ {
			info := TMapUnitInfo{}
			if err := binary.Read(file, binary.LittleEndian, &info); err != nil {
				return fmt.Errorf("failed to read cell [%d,%d]: %w", i, j, err)
			}
			m.Cells[i][j] = &TMapCellInfo{
				Flag:    0,
				ObjList: make([]*MapObject, 0, 4),
			}
		}
	}
	
	return nil
}

func readStruct(file interface{}, v interface{}) error {
	return nil
}

func (m *GameMap) CanWalk(x, y int) bool {
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return false
	}
	cell := m.Cells[x][y]
	return cell != nil && (cell.Flag&0x80) == 0
}

func (m *GameMap) CanFly(x, y int) bool {
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return false
	}
	return true
}

func (m *GameMap) CanMove(x, y int) bool {
	if !m.CanWalk(x, y) {
		return false
	}
	
	if m.NODRUG {
		return false
	}
	
	return true
}

func (m *GameMap) GetCell(x, y int) *TMapCellInfo {
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return nil
	}
	return m.Cells[x][y]
}

func (m *GameMap) AddPlayer(obj MapObject) {
	m.Objects.Lock()
	defer m.Objects.Unlock()
	m.Players[obj.GetID()] = obj
}

func (m *GameMap) DelPlayer(id int32) {
	m.Objects.Lock()
	defer m.Objects.Unlock()
	delete(m.Players, id)
}

func (m *GameMap) AddNPC(obj MapObject) {
	m.Objects.Lock()
	defer m.Objects.Unlock()
	m.NPCs[obj.GetID()] = obj
}

func (m *GameMap) DelNPC(id int32) {
	m.Objects.Lock()
	defer m.Objects.Unlock()
	delete(m.NPCs, id)
}

func (m *GameMap) AddMonster(obj MapObject) {
	m.Objects.Lock()
	defer m.Objects.Unlock()
	m.Monsters[obj.GetID()] = obj
}

func (m *GameMap) DelMonster(id int32) {
	m.Objects.Lock()
	defer m.Objects.Unlock()
	delete(m.Monsters, id)
}

func (m *GameMap) GetPlayersInView(x, y, range_ int) []MapObject {
	m.Objects.RLock()
	defer m.Objects.RUnlock()
	
	var result []MapObject
	for _, p := range m.Players {
		dx := p.GetX() - x
		dy := p.GetY() - y
		if dx < 0 {
			dx = -dx
		}
		if dy < 0 {
			dy = -dy
		}
		if dx <= range_ && dy <= range_ {
			result = append(result, p)
		}
	}
	return result
}

func (m *GameMap) GetObjectsInView(x, y, range_ int) []MapObject {
	m.Objects.RLock()
	defer m.Objects.RUnlock()
	
	var result []MapObject
	for _, npc := range m.NPCs {
		dx := npc.GetX() - x
		dy := npc.GetY() - y
		if dx < 0 {
			dx = -dx
		}
		if dy < 0 {
			dy = -dy
		}
		if dx <= range_ && dy <= range_ {
			result = append(result, npc)
		}
	}
	
	for _, mon := range m.Monsters {
		dx := mon.GetX() - x
		dy := mon.GetY() - y
		if dx < 0 {
			dx = -dx
		}
		if dy < 0 {
			dy = -dy
		}
		if dx <= range_ && dy <= range_ {
			result = append(result, mon)
		}
	}
	
	return result
}

func (m *GameMap) GetX() int { return m.Width }
func (m *GameMap) GetY() int { return m.Height }
func (m *GameMap) GetName() string { return m.MapName }

type MapManager struct {
	Maps       map[string]*GameMap
	Mutex      sync.RWMutex
	PathFinder *PathFinder
}

func NewMapManager() *MapManager {
	return &MapManager{
		Maps:       make(map[string]*GameMap),
		PathFinder: NewPathFinder(),
	}
}

func (mm *MapManager) LoadMap(mapName, filename string) (*GameMap, error) {
	mm.Mutex.Lock()
	defer mm.Mutex.Unlock()
	
	if m, ok := mm.Maps[mapName]; ok {
		return m, nil
	}
	
	m := NewGameMap(mapName, filename)
	if err := m.LoadFromFile(filename); err != nil {
		return nil, err
	}
	
	mm.Maps[mapName] = m
	return m, nil
}

func (mm *MapManager) GetMap(mapName string) *GameMap {
	mm.Mutex.RLock()
	defer mm.Mutex.RUnlock()
	return mm.Maps[mapName]
}

func (mm *MapManager) FindPath(mapName string, startX, startY, endX, endY int) []*PathNode {
	mm.Mutex.RLock()
	m := mm.Maps[mapName]
	mm.Mutex.RUnlock()
	
	if m == nil {
		return nil
	}
	
	return mm.PathFinder.FindPath(m, startX, startY, endX, endY)
}

func (mm *MapManager) GetAllMaps() []*GameMap {
	mm.Mutex.RLock()
	defer mm.Mutex.RUnlock()
	
	result := make([]*GameMap, 0, len(mm.Maps))
	for _, m := range mm.Maps {
		result = append(result, m)
	}
	return result
}

var defaultMapManager *MapManager

func init() {
	defaultMapManager = NewMapManager()
}

func GetMapManager() *MapManager {
	return defaultMapManager
}
