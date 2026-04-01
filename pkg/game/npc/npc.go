package npc

import (
	"github.com/mir2go/mir2/pkg/game/actor"
	"github.com/mir2go/mir2/pkg/protocol"
	"github.com/mir2go/mir2/pkg/script"
)

type NPCType byte

const (
	NPC_NORMAL NPCType = iota
	NPC_MERCHANT
	NPC_GUARD
)

type TNormNpc struct {
	*actor.BaseObject
	NPCType   NPCType
	Script    *script.NpcScript
	ScriptFile string
	IsCastle  bool
	Castle    interface{}
	
	ScriptList []string
	ScriptCmd []int
	
	DialogText string
	GoodsList  []TGoods
}

type TGoods struct {
	ItemName   string
	Count      int
	RefillTime uint32
	RefillTick uint32
}

type TMerchant struct {
	*TNormNpc
	BuyPrice    []int
	SellPrice   []int
	RepairCost  []int
}

func NewNormNpc(id int32, name, mapName string, x, y int) *TNormNpc {
	return &TNormNpc{
		BaseObject: &actor.BaseObject{
			ID:        id,
			Name:      name,
			MapName:   mapName,
			X:         x,
			Y:         y,
			RaceServer: protocol.RC_NPC,
		},
		NPCType: NPC_NORMAL,
		GoodsList: make([]TGoods, 0),
	}
}

func (n *TNormNpc) Initialize() error {
	return nil
}

func (n *TNormNpc) Run() bool {
	return true
}

func (n *TNormNpc) ProcessCheckTime() {
}

func (n *TNormNpc) GetFeature() int32 {
	feature := int32(0)
	feature |= int32(n.Direction) << 16
	feature |= 0x02000000
	return feature
}

func (n *TNormNpc) UserSelect( player *actor.Player, npc *TNormNpc, data string) bool {
	return false
}

func (n *TNormNpc) ExecuteScript(label string) bool {
	if n.Script == nil {
		return false
	}
	
	scriptEngine := script.NewScriptEngine()
	return scriptEngine.ExecuteScript(n.ScriptFile) == nil
}

func (n *TNormNpc) AddGoods(itemName string, count int, price int) {
	goods := TGoods{
		ItemName:   itemName,
		Count:      count,
		RefillTime: 0,
		RefillTick: 0,
	}
	n.GoodsList = append(n.GoodsList, goods)
}

func (n *TNormNpc) GetGoodsList() []TGoods {
	return n.GoodsList
}

func (n *TNormNpc) GetItemPrice(itemName string) int {
	for i, goods := range n.GoodsList {
		if goods.ItemName == itemName {
			return 100 * (i + 1)
		}
	}
	return 0
}

func NewMerchant(id int32, name, mapName string, x, y int) *TMerchant {
	return &TMerchant{
		TNormNpc: NewNormNpc(id, name, mapName, x, y),
	}
}

func (m *TMerchant) UserBuy(player *actor.Player, itemName string, count int) bool {
	return true
}

func (m *TMerchant) UserSell(player *actor.Player, itemName string, count int) bool {
	return true
}

func (m *TMerchant) UserRepair(player *actor.Player, itemIndex int) bool {
	return true
}

type NPCManager struct {
	NpcList   map[int32]*TNormNpc
	NPCByName map[string][]*TNormNpc
}

func NewNPCManager() *NPCManager {
	return &NPCManager{
		NpcList:   make(map[int32]*TNormNpc),
		NPCByName: make(map[string][]*TNormNpc),
	}
}

func (nm *NPCManager) AddNPC(npc *TNormNpc) {
	nm.NpcList[npc.ID] = npc
	
	name := npc.Name
	if _, ok := nm.NPCByName[name]; !ok {
		nm.NPCByName[name] = make([]*TNormNpc, 0)
	}
	nm.NPCByName[name] = append(nm.NPCByName[name], npc)
}

func (nm *NPCManager) DelNPC(id int32) {
	if npc, ok := nm.NpcList[id]; ok {
		name := npc.Name
		if list, ok := nm.NPCByName[name]; ok {
			for i, n := range list {
				if n.ID == id {
					nm.NPCByName[name] = append(list[:i], list[i+1:]...)
					break
				}
			}
		}
	}
	delete(nm.NpcList, id)
}

func (nm *NPCManager) GetNPC(id int32) *TNormNpc {
	return nm.NpcList[id]
}

func (nm *NPCManager) GetNPCByName(name string) []*TNormNpc {
	return nm.NPCByName[name]
}

func (nm *NPCManager) GetNPCsInMap(mapName string) []*TNormNpc {
	var result []*TNormNpc
	for _, npc := range nm.NpcList {
		if npc.MapName == mapName {
			result = append(result, npc)
		}
	}
	return result
}

func (nm *NPCManager) GetNPCNear(x, y, range_ int) *TNormNpc {
	for _, npc := range nm.NpcList {
		dx := npc.X - x
		dy := npc.Y - y
		if dx < 0 {
			dx = -dx
		}
		if dy < 0 {
			dy = -dy
		}
		if dx <= range_ && dy <= range_ {
			return npc
		}
	}
	return nil
}
