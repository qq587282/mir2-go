package item

import (
	"sync"

	"github.com/mir2go/mir2/pkg/protocol"
)

type Job byte
type Gender byte

const (
	JobWarr   Job = 0
	JobWizard Job = 1
	JobTaos   Job = 2
	JobRogue  Job = 3
)

const (
	GenderMan   Gender = 0
	GenderWoman Gender = 1
)

type TAbility struct {
	Level         int32
	AC, MAC       int32
	DC, MC, SC    int32
	CC            int32
	HP, MP        int32
	MaxHP, MaxMP  int32
	Exp, MaxExp   uint32
	Weight        uint16
	MaxWeight     uint16
	WearWeight    uint16
	MaxWearWeight uint16
	HandWeight    uint16
	MaxHandWeight uint16
}

type TAddAbility struct {
	WHP, WMP      int32
	HitPoint      uint16
	SpeedPoint    uint16
	AC, MAC       int32
	DC, MC, SC    int32
	CC            int32
	AntiPoison    uint16
	PoisonRecover uint16
	HealthRecover uint16
	SpellRecover  uint16
	AntiMagic     uint16
	Luck, UnLuck  int8
	HitSpeed      int32
}

type ItemType byte

const (
	ItemTypeWeapon     ItemType = 5
	ItemTypeArmor      ItemType = 10
	ItemTypeHelmet     ItemType = 15
	ItemTypeNecklace   ItemType = 19
	ItemTypeRing       ItemType = 22
	ItemTypeBracelet   ItemType = 24
	ItemTypeBelt       ItemType = 23
	ItemTypeConsumable ItemType = 2
	ItemTypeScorll     ItemType = 1
	ItemTypePotion     ItemType = 2
)

type ItemQuality byte

const (
	ItemQualityCommon    ItemQuality = 0
	ItemQualityRare      ItemQuality = 1
	ItemQualityEpic      ItemQuality = 2
	ItemQualityLegendary ItemQuality = 3
	ItemQualityMythical  ItemQuality = 4
)

type ItemFlag byte

const (
	ItemFlagNone           ItemFlag = 0
	ItemFlagBindOnPick     ItemFlag = 0x01
	ItemFlagBindOnEquip    ItemFlag = 0x02
	ItemFlagBindOnUse      ItemFlag = 0x04
	ItemFlagCannotTrade    ItemFlag = 0x08
	ItemFlagCannotDrop     ItemFlag = 0x10
	ItemFlagCannotRepair   ItemFlag = 0x20
	ItemFlagCannotUpgrade  ItemFlag = 0x40
	ItemFlagIndestructible ItemFlag = 0x80
)

type ItemAttr struct {
	Name    string
	Index   int
	StdMode byte
	Shape   byte
	Weight  byte
	Looks   uint16
	DuraMax uint16

	AC  int32
	MAC int32
	DC  int32
	MC  int32
	SC  int32
	CC  int32

	Need      int32
	NeedLevel int32
	Price     int32
	BuyPrice  int32
	SellPrice int32

	Job    Job
	Gender Gender

	HP    int32
	MP    int32
	Hit   int32
	Speed int32

	Desc string

	ItemType ItemType
	Quality  ItemQuality
	Flags    ItemFlag

	AniCount int
	Overlap  int
	Effect   int
}

type TUserItem struct {
	MakeIndex int32
	Name      string
	Index     uint16
	Dura      uint16
	DuraMax   uint16
	Value     [14]byte
	AddValue  [14]byte
	AddPoint  [14]byte
	MaxDate   int64

	BindMode   int
	ExpireDate int64
	UserItemID int64
	UseDate    int64
	ValidDate  int64

	Name2 [14]byte
	Desc  [16]byte

	AC  [2]int16
	MAC [2]int16
	DC  [2]int16
	MC  [2]int16
	SC  [2]int16

	HP int16
	MP int16

	Hit   int16
	Speed int16

	Strong int8

	Level        int8
	UpgradeLevel int8

	Count int32

	Price int32

	CaseParams [8]int32

	RandomProps [5]int32
}

func (iu *TUserItem) GetItemAttr() *ItemAttr {
	if int(iu.Index) >= len(StdItemList) {
		return nil
	}
	return &StdItemList[iu.Index]
}

func (iu *TUserItem) IsExpire() bool {
	if iu.ValidDate > 0 {
		return iu.ValidDate < 86400
	}
	return false
}

func (iu *TUserItem) CalcMaxDura() uint16 {
	attr := iu.GetItemAttr()
	if attr == nil {
		return 0
	}
	base := attr.DuraMax
	upgrade := uint16(iu.UpgradeLevel) * 200
	return base + upgrade
}

func (iu *TUserItem) Upgrade() bool {
	if iu.UpgradeLevel >= 3 {
		return false
	}
	iu.UpgradeLevel++
	iu.DuraMax = iu.CalcMaxDura()
	return true
}

var (
	StdItemList   []ItemAttr
	itemIndexMap  map[string]int
	looksIndexMap map[uint16]int
	itemMutex     sync.RWMutex
)

func init() {
	StdItemList = loadStdItems()
	itemIndexMap = make(map[string]int)
	looksIndexMap = make(map[uint16]int)

	for i, item := range StdItemList {
		itemIndexMap[item.Name] = i
		looksIndexMap[item.Looks] = i
	}
}

func loadStdItems() []ItemAttr {
	return []ItemAttr{
		{Index: 0, Name: "木剑", StdMode: 5, Shape: 1, Weight: 6, Looks: 13000, DuraMax: 20, DC: 1, NeedLevel: 0, Price: 100, ItemType: ItemTypeWeapon},
		{Index: 1, Name: "铁剑", StdMode: 5, Shape: 2, Weight: 8, Looks: 13001, DuraMax: 30, DC: 2, NeedLevel: 7, Price: 300, ItemType: ItemTypeWeapon},
		{Index: 2, Name: "青铜剑", StdMode: 5, Shape: 3, Weight: 10, Looks: 13002, DuraMax: 40, DC: 3, NeedLevel: 12, Price: 800, ItemType: ItemTypeWeapon},
		{Index: 3, Name: "炼狱", StdMode: 5, Shape: 4, Weight: 14, Looks: 13003, DuraMax: 50, DC: 5, NeedLevel: 25, Price: 5000, ItemType: ItemTypeWeapon},
		{Index: 4, Name: "魔杖", StdMode: 5, Shape: 10, Weight: 6, Looks: 13004, DuraMax: 20, MC: 2, NeedLevel: 7, Price: 400, ItemType: ItemTypeWeapon},
		{Index: 5, Name: "骨玉", StdMode: 5, Shape: 11, Weight: 10, Looks: 13005, DuraMax: 35, MC: 3, NeedLevel: 20, Price: 3000, ItemType: ItemTypeWeapon},
		{Index: 6, Name: "血饮", StdMode: 5, Shape: 12, Weight: 12, Looks: 13006, DuraMax: 40, MC: 4, NeedLevel: 30, Price: 8000, ItemType: ItemTypeWeapon},

		{Index: 10, Name: "布衣(男)", StdMode: 10, Shape: 0, Weight: 3, Looks: 10000, DuraMax: 20, AC: 1, NeedLevel: 0, Price: 100, ItemType: ItemTypeArmor},
		{Index: 11, Name: "布衣(女)", StdMode: 10, Shape: 1, Weight: 3, Looks: 10001, DuraMax: 20, AC: 1, NeedLevel: 0, Price: 100, ItemType: ItemTypeArmor},
		{Index: 12, Name: "轻型盔甲(男)", StdMode: 10, Shape: 10, Weight: 8, Looks: 10002, DuraMax: 35, AC: 3, NeedLevel: 7, Price: 500, ItemType: ItemTypeArmor},
		{Index: 13, Name: "轻型盔甲(女)", StdMode: 10, Shape: 11, Weight: 8, Looks: 10003, DuraMax: 35, AC: 3, NeedLevel: 7, Price: 500, ItemType: ItemTypeArmor},
		{Index: 14, Name: "中型盔甲(男)", StdMode: 10, Shape: 20, Weight: 15, Looks: 10004, DuraMax: 50, AC: 5, NeedLevel: 15, Price: 2000, ItemType: ItemTypeArmor},
		{Index: 15, Name: "中型盔甲(女)", StdMode: 10, Shape: 21, Weight: 15, Looks: 10005, DuraMax: 50, AC: 5, NeedLevel: 15, Price: 2000, ItemType: ItemTypeArmor},
		{Index: 16, Name: "重型盔甲(男)", StdMode: 10, Shape: 30, Weight: 25, Looks: 10006, DuraMax: 70, AC: 8, NeedLevel: 22, Price: 6000, ItemType: ItemTypeArmor},
		{Index: 17, Name: "重型盔甲(女)", StdMode: 10, Shape: 31, Weight: 25, Looks: 10007, DuraMax: 70, AC: 8, NeedLevel: 22, Price: 6000, ItemType: ItemTypeArmor},

		{Index: 18, Name: "魔法袍(男)", StdMode: 10, Shape: 40, Weight: 6, Looks: 10008, DuraMax: 30, AC: 2, MC: 1, NeedLevel: 7, Price: 800, ItemType: ItemTypeArmor},
		{Index: 19, Name: "魔法袍(女)", StdMode: 10, Shape: 41, Weight: 6, Looks: 10009, DuraMax: 30, AC: 2, MC: 1, NeedLevel: 7, Price: 800, ItemType: ItemTypeArmor},

		{Index: 20, Name: "骷髅头盔", StdMode: 15, Shape: 0, Weight: 3, Looks: 15000, DuraMax: 25, AC: 1, NeedLevel: 10, Price: 1000, ItemType: ItemTypeHelmet},
		{Index: 21, Name: "道士头盔", StdMode: 15, Shape: 1, Weight: 4, Looks: 15001, DuraMax: 30, AC: 1, MAC: 1, NeedLevel: 12, Price: 1500, ItemType: ItemTypeHelmet},

		{Index: 22, Name: "大手镯", StdMode: 24, Shape: 0, Weight: 1, Looks: 20000, DuraMax: 10, NeedLevel: 0, Price: 50, ItemType: ItemTypeBracelet},
		{Index: 23, Name: "小手镯", StdMode: 24, Shape: 1, Weight: 1, Looks: 20001, DuraMax: 10, NeedLevel: 0, Price: 50, ItemType: ItemTypeBracelet},
		{Index: 24, Name: "坚固手套", StdMode: 26, Shape: 0, Weight: 1, Looks: 20002, DuraMax: 15, AC: 1, NeedLevel: 5, Price: 300, ItemType: ItemTypeBracelet},
		{Index: 25, Name: "死神手套", StdMode: 26, Shape: 1, Weight: 1, Looks: 20003, DuraMax: 20, DC: 1, NeedLevel: 15, Price: 2000, ItemType: ItemTypeBracelet},

		{Index: 26, Name: "古铜戒指", StdMode: 22, Shape: 0, Weight: 1, Looks: 21000, DuraMax: 10, NeedLevel: 0, Price: 100, ItemType: ItemTypeRing},
		{Index: 27, Name: "魔法戒指", StdMode: 22, Shape: 1, Weight: 1, Looks: 21001, DuraMax: 10, MC: 1, NeedLevel: 7, Price: 500, ItemType: ItemTypeRing},
		{Index: 28, Name: "力量戒指", StdMode: 22, Shape: 2, Weight: 1, Looks: 21002, DuraMax: 10, DC: 1, NeedLevel: 10, Price: 800, ItemType: ItemTypeRing},
		{Index: 29, Name: "护身符", StdMode: 22, Shape: 3, Weight: 1, Looks: 21003, DuraMax: 15, NeedLevel: 12, Price: 1500, ItemType: ItemTypeRing},

		{Index: 30, Name: "项链", StdMode: 19, Shape: 0, Weight: 1, Looks: 22000, DuraMax: 10, NeedLevel: 0, Price: 100, ItemType: ItemTypeNecklace},
		{Index: 31, Name: "Talisman", StdMode: 19, Shape: 1, Weight: 1, Looks: 22001, DuraMax: 10, NeedLevel: 5, Price: 300, ItemType: ItemTypeNecklace},

		{Index: 40, Name: "金创药(小)", StdMode: 2, Shape: 0, Weight: 1, Looks: 23000, DuraMax: 1, HP: 30, NeedLevel: 0, Price: 50, ItemType: ItemTypePotion},
		{Index: 41, Name: "金创药(中)", StdMode: 2, Shape: 1, Weight: 1, Looks: 23001, DuraMax: 1, HP: 60, NeedLevel: 5, Price: 100, ItemType: ItemTypePotion},
		{Index: 42, Name: "金创药(大)", StdMode: 2, Shape: 2, Weight: 1, Looks: 23002, DuraMax: 1, HP: 100, NeedLevel: 10, Price: 200, ItemType: ItemTypePotion},
		{Index: 43, Name: "魔法药(小)", StdMode: 2, Shape: 10, Weight: 1, Looks: 23003, DuraMax: 1, MP: 20, NeedLevel: 0, Price: 50, ItemType: ItemTypePotion},
		{Index: 44, Name: "魔法药(中)", StdMode: 2, Shape: 11, Weight: 1, Looks: 23004, DuraMax: 1, MP: 40, NeedLevel: 5, Price: 100, ItemType: ItemTypePotion},
		{Index: 45, Name: "魔法药(大)", StdMode: 2, Shape: 12, Weight: 1, Looks: 23005, DuraMax: 1, MP: 80, NeedLevel: 10, Price: 200, ItemType: ItemTypePotion},

		{Index: 50, Name: "回城卷", StdMode: 1, Shape: 0, Weight: 1, Looks: 23010, DuraMax: 1, NeedLevel: 0, Price: 30, ItemType: ItemTypeScorll},
		{Index: 51, Name: "随机卷", StdMode: 1, Shape: 1, Weight: 1, Looks: 23011, DuraMax: 1, NeedLevel: 0, Price: 50, ItemType: ItemTypeScorll},
		{Index: 52, Name: "地牢逃脱卷", StdMode: 1, Shape: 2, Weight: 1, Looks: 23012, DuraMax: 1, NeedLevel: 0, Price: 80, ItemType: ItemTypeScorll},

		{Index: 60, Name: "祝福油", StdMode: 1, Shape: 100, Weight: 1, Looks: 23050, DuraMax: 1, NeedLevel: 0, Price: 500, ItemType: ItemTypeConsumable},
		{Index: 61, Name: "战神油", StdMode: 1, Shape: 101, Weight: 1, Looks: 23051, DuraMax: 1, NeedLevel: 0, Price: 1000, ItemType: ItemTypeConsumable},
		{Index: 62, Name: "修复油", StdMode: 1, Shape: 102, Weight: 1, Looks: 23052, DuraMax: 1, NeedLevel: 0, Price: 300, ItemType: ItemTypeConsumable},

		{Index: 70, Name: "随机传送石", StdMode: 1, Shape: 200, Weight: 1, Looks: 23060, DuraMax: 1, NeedLevel: 0, Price: 100, ItemType: ItemTypeScorll},
		{Index: 71, Name: "定位传送石", StdMode: 1, Shape: 201, Weight: 1, Looks: 23061, DuraMax: 1, NeedLevel: 0, Price: 150, ItemType: ItemTypeScorll},

		{Index: 100, Name: "强效金创药", StdMode: 2, Shape: 3, Weight: 1, Looks: 23006, DuraMax: 1, HP: 150, NeedLevel: 15, Price: 300, ItemType: ItemTypePotion},
		{Index: 101, Name: "强效魔法药", StdMode: 2, Shape: 13, Weight: 1, Looks: 23007, DuraMax: 1, MP: 120, NeedLevel: 15, Price: 300, ItemType: ItemTypePotion},

		{Index: 110, Name: "疗伤药", StdMode: 2, Shape: 4, Weight: 1, Looks: 23008, DuraMax: 1, HP: 200, NeedLevel: 20, Price: 500, ItemType: ItemTypePotion},
		{Index: 111, Name: "魔力药", StdMode: 2, Shape: 14, Weight: 1, Looks: 23009, DuraMax: 1, MP: 180, NeedLevel: 20, Price: 500, ItemType: ItemTypePotion},

		{Index: 120, Name: "魔力药水", StdMode: 2, Shape: 20, Weight: 1, Looks: 23020, DuraMax: 1, MP: 50, NeedLevel: 0, Price: 50, ItemType: ItemTypePotion},

		{Index: 130, Name: "召唤令", StdMode: 1, Shape: 30, Weight: 1, Looks: 23030, DuraMax: 1, NeedLevel: 0, Price: 100, ItemType: ItemTypeConsumable},

		{Index: 140, Name: "行会召唤令", StdMode: 1, Shape: 31, Weight: 1, Looks: 23031, DuraMax: 1, NeedLevel: 0, Price: 200, ItemType: ItemTypeConsumable},
	}
}

func FindItem(name string) *ItemAttr {
	itemMutex.RLock()
	defer itemMutex.RUnlock()

	if idx, ok := itemIndexMap[name]; ok {
		return &StdItemList[idx]
	}
	return nil
}

func GetItemByIndex(idx int) *ItemAttr {
	itemMutex.RLock()
	defer itemMutex.RUnlock()

	if idx >= 0 && idx < len(StdItemList) {
		return &StdItemList[idx]
	}
	return nil
}

func GetItemByLooks(looks uint16) *ItemAttr {
	itemMutex.RLock()
	defer itemMutex.RUnlock()

	if idx, ok := looksIndexMap[looks]; ok {
		return &StdItemList[idx]
	}
	return nil
}

func GetItemByLooksEx(looks uint16) (int, *ItemAttr) {
	itemMutex.RLock()
	defer itemMutex.RUnlock()

	if idx, ok := looksIndexMap[looks]; ok {
		return idx, &StdItemList[idx]
	}
	return -1, nil
}

func AddItem(item ItemAttr) {
	itemMutex.Lock()
	defer itemMutex.Unlock()

	item.Index = len(StdItemList)
	StdItemList = append(StdItemList, item)
	itemIndexMap[item.Name] = len(StdItemList) - 1
	looksIndexMap[item.Looks] = len(StdItemList) - 1
}

func GetStdItemCount() int {
	itemMutex.RLock()
	defer itemMutex.RUnlock()
	return len(StdItemList)
}

func GetStdItemList() []ItemAttr {
	itemMutex.RLock()
	defer itemMutex.RUnlock()

	result := make([]ItemAttr, len(StdItemList))
	copy(result, StdItemList)
	return result
}

func NewUserItem(itemIndex int) *TUserItem {
	attr := GetItemByIndex(itemIndex)
	if attr == nil {
		return nil
	}

	return &TUserItem{
		MakeIndex:  0,
		Name:       attr.Name,
		Index:      uint16(itemIndex),
		Dura:       attr.DuraMax,
		DuraMax:    attr.DuraMax,
		UserItemID: 0,
		Count:      1,
	}
}

func GetUserItemAttr(item *TUserItem) *ItemAttr {
	if item == nil {
		return nil
	}
	return GetItemByIndex(int(item.Index))
}

func CalcPlayerAbility2(base *TAbility, addBase *TAddAbility, job Job, level int32, useItems [protocol.U_CHARM + 1]TUserItem) {
	addBase.AC = 0
	addBase.MAC = 0
	addBase.DC = 0
	addBase.MC = 0
	addBase.SC = 0

	for i := 0; i <= int(protocol.U_CHARM); i++ {
		userItem := useItems[i]
		if userItem.Dura == 0 {
			continue
		}

		attr := GetItemByIndex(int(userItem.Index))
		if attr == nil {
			continue
		}

		addBase.AC += attr.AC
		addBase.MAC += attr.MAC
		addBase.DC += attr.DC
		addBase.MC += attr.MC
		addBase.SC += attr.SC
		addBase.HitPoint += uint16(attr.Hit)
		addBase.SpeedPoint += uint16(attr.Speed)
	}

	base.MaxWeight = uint16(50 + level*3 + addBase.WHP/10)
	base.MaxWearWeight = uint16(15 + level)
	base.MaxHandWeight = uint16(50 + level*2)
}

type ItemDrop struct {
	Item      *TUserItem
	MakeIndex int32
	MapName   string
	X, Y      int
	DropTime  int64
	OwnerID   int32
	CanPick   bool
}

type ItemManager struct {
	Items     map[int32]*ItemDrop
	NextIndex int32
	Mutex     sync.RWMutex
}

var DefaultItemManager *ItemManager

func init() {
	DefaultItemManager = NewItemManager()
}

func NewItemManager() *ItemManager {
	return &ItemManager{
		Items:     make(map[int32]*ItemDrop),
		NextIndex: 1,
	}
}

func (im *ItemManager) GetNextIndex() int32 {
	im.Mutex.Lock()
	defer im.Mutex.Unlock()

	id := im.NextIndex
	im.NextIndex++
	return id
}

func (im *ItemManager) AddItem(item *ItemDrop) int32 {
	im.Mutex.Lock()
	defer im.Mutex.Unlock()

	item.MakeIndex = im.NextIndex
	im.NextIndex++
	im.Items[item.MakeIndex] = item
	return item.MakeIndex
}

func (im *ItemManager) GetItem(makeIndex int32) *ItemDrop {
	im.Mutex.RLock()
	defer im.Mutex.RUnlock()
	return im.Items[makeIndex]
}

func (im *ItemManager) DelItem(makeIndex int32) bool {
	im.Mutex.Lock()
	defer im.Mutex.Unlock()

	if _, ok := im.Items[makeIndex]; ok {
		delete(im.Items, makeIndex)
		return true
	}
	return false
}

func (im *ItemManager) GetMapItems(mapName string, x, y int) []*ItemDrop {
	im.Mutex.RLock()
	defer im.Mutex.RUnlock()

	var result []*ItemDrop
	for _, item := range im.Items {
		if item.MapName == mapName {
			if x > 0 && y > 0 {
				dx := item.X - x
				dy := item.Y - y
				if dx < 0 {
					dx = -dx
				}
				if dy < 0 {
					dy = -dy
				}
				if dx <= 1 && dy <= 1 {
					result = append(result, item)
				}
			} else {
				result = append(result, item)
			}
		}
	}
	return result
}

func (im *ItemManager) DropItem(item *TUserItem, mapName string, x, y int, ownerID int32, canPick bool) int32 {
	drop := &ItemDrop{
		Item:     item,
		MapName:  mapName,
		X:        x,
		Y:        y,
		DropTime: 0,
		OwnerID:  ownerID,
		CanPick:  canPick,
	}

	return im.AddItem(drop)
}

func GetItemManager() *ItemManager {
	return DefaultItemManager
}
