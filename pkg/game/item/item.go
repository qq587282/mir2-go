package item

import (
	"github.com/mir2go/mir2/pkg/game/actor"
	"github.com/mir2go/mir2/pkg/protocol"
)

type ItemAttr struct {
	Name      string
	StdMode   byte
	Shape     byte
	Weight    byte
	Looks     uint16
	DuraMax   uint16
	
	AC     int32
	MAC    int32
	DC     int32
	MC     int32
	SC     int32
	CC     int32
	
	Need     int32
	NeedLevel int32
	Price    int32
	
	Job     actor.Job
	Gender  actor.Gender
	
	HP     int32
	MP     int32
	Hit    int32
	Speed  int32
	
	Desc   string
}

var StdItemList []ItemAttr

func init() {
	StdItemList = loadStdItems()
}

func loadStdItems() []ItemAttr {
	return []ItemAttr{
		{Name: "木剑", StdMode: 5, Shape: 1, Weight: 6, Looks: 13000, DuraMax: 20, DC: 1, NeedLevel: 0, Price: 100},
		{Name: "铁剑", StdMode: 5, Shape: 2, Weight: 8, Looks: 13001, DuraMax: 30, DC: 2, NeedLevel: 7, Price: 300},
		{Name: "青铜剑", StdMode: 5, Shape: 3, Weight: 10, Looks: 13002, DuraMax: 40, DC: 3, NeedLevel: 12, Price: 800},
		{Name: "炼狱", StdMode: 5, Shape: 4, Weight: 14, Looks: 13003, DuraMax: 50, DC: 5, NeedLevel: 25, Price: 5000},
		{Name: "魔杖", StdMode: 5, Shape: 10, Weight: 6, Looks: 13004, DuraMax: 20, MC: 2, NeedLevel: 7, Price: 400},
		{Name: "骨玉", StdMode: 5, Shape: 11, Weight: 10, Looks: 13005, DuraMax: 35, MC: 3, NeedLevel: 20, Price: 3000},
		{Name: "血饮", StdMode: 5, Shape: 12, Weight: 12, Looks: 13006, DuraMax: 40, MC: 4, NeedLevel: 30, Price: 8000},
		
		{Name: "布衣(男)", StdMode: 10, Shape: 0, Weight: 3, Looks: 10000, DuraMax: 20, AC: 1, NeedLevel: 0, Price: 100},
		{Name: "布衣(女)", StdMode: 10, Shape: 1, Weight: 3, Looks: 10001, DuraMax: 20, AC: 1, NeedLevel: 0, Price: 100},
		{Name: "轻型盔甲(男)", StdMode: 10, Shape: 10, Weight: 8, Looks: 10002, DuraMax: 35, AC: 3, NeedLevel: 7, Price: 500},
		{Name: "轻型盔甲(女)", StdMode: 10, Shape: 11, Weight: 8, Looks: 10003, DuraMax: 35, AC: 3, NeedLevel: 7, Price: 500},
		{Name: "中型盔甲(男)", StdMode: 10, Shape: 20, Weight: 15, Looks: 10004, DuraMax: 50, AC: 5, NeedLevel: 15, Price: 2000},
		{Name: "中型盔甲(女)", StdMode: 10, Shape: 21, Weight: 15, Looks: 10005, DuraMax: 50, AC: 5, NeedLevel: 15, Price: 2000},
		{Name: "重型盔甲(男)", StdMode: 10, Shape: 30, Weight: 25, Looks: 10006, DuraMax: 70, AC: 8, NeedLevel: 22, Price: 6000},
		{Name: "重型盔甲(女)", StdMode: 10, Shape: 31, Weight: 25, Looks: 10007, DuraMax: 70, AC: 8, NeedLevel: 22, Price: 6000},
		
		{Name: "魔法袍(男)", StdMode: 10, Shape: 40, Weight: 6, Looks: 10008, DuraMax: 30, AC: 2, MC: 1, NeedLevel: 7, Price: 800},
		{Name: "魔法袍(女)", StdMode: 10, Shape: 41, Weight: 6, Looks: 10009, DuraMax: 30, AC: 2, MC: 1, NeedLevel: 7, Price: 800},
		
		{Name: "骷髅头盔", StdMode: 15, Shape: 0, Weight: 3, Looks: 15000, DuraMax: 25, AC: 1, NeedLevel: 10, Price: 1000},
		{Name: "道士头盔", StdMode: 15, Shape: 1, Weight: 4, Looks: 15001, DuraMax: 30, AC: 1, MAC: 1, NeedLevel: 12, Price: 1500},
		
		{Name: "大手镯", StdMode: 24, Shape: 0, Weight: 1, Looks: 20000, DuraMax: 10, NeedLevel: 0, Price: 50},
		{Name: "小手镯", StdMode: 24, Shape: 1, Weight: 1, Looks: 20001, DuraMax: 10, NeedLevel: 0, Price: 50},
		{Name: "坚固手套", StdMode: 26, Shape: 0, Weight: 1, Looks: 20002, DuraMax: 15, AC: 1, NeedLevel: 5, Price: 300},
		{Name: "死神手套", StdMode: 26, Shape: 1, Weight: 1, Looks: 20003, DuraMax: 20, DC: 1, NeedLevel: 15, Price: 2000},
		
		{Name: "古铜戒指", StdMode: 22, Shape: 0, Weight: 1, Looks: 21000, DuraMax: 10, NeedLevel: 0, Price: 100},
		{Name: "魔法戒指", StdMode: 22, Shape: 1, Weight: 1, Looks: 21001, DuraMax: 10, MC: 1, NeedLevel: 7, Price: 500},
		{Name: "力量戒指", StdMode: 22, Shape: 2, Weight: 1, Looks: 21002, DuraMax: 10, DC: 1, NeedLevel: 10, Price: 800},
		{Name: "护身符", StdMode: 22, Shape: 3, Weight: 1, Looks: 21003, DuraMax: 15, NeedLevel: 12, Price: 1500},
		
		{Name: "项链", StdMode: 19, Shape: 0, Weight: 1, Looks: 22000, DuraMax: 10, NeedLevel: 0, Price: 100},
		{Name: "Talisman", StdMode: 19, Shape: 1, Weight: 1, Looks: 22001, DuraMax: 10, NeedLevel: 5, Price: 300},
		
		{Name: "金创药(小)", StdMode: 2, Shape: 0, Weight: 1, Looks: 23000, DuraMax: 1, HP: 30, NeedLevel: 0, Price: 50},
		{Name: "金创药(中)", StdMode: 2, Shape: 1, Weight: 1, Looks: 23001, DuraMax: 1, HP: 60, NeedLevel: 5, Price: 100},
		{Name: "金创药(大)", StdMode: 2, Shape: 2, Weight: 1, Looks: 23002, DuraMax: 1, HP: 100, NeedLevel: 10, Price: 200},
		{Name: "魔法药(小)", StdMode: 2, Shape: 10, Weight: 1, Looks: 23003, DuraMax: 1, MP: 20, NeedLevel: 0, Price: 50},
		{Name: "魔法药(中)", StdMode: 2, Shape: 11, Weight: 1, Looks: 23004, DuraMax: 1, MP: 40, NeedLevel: 5, Price: 100},
		{Name: "魔法药(大)", StdMode: 2, Shape: 12, Weight: 1, Looks: 23005, DuraMax: 1, MP: 80, NeedLevel: 10, Price: 200},
		
		{Name: "回城卷", StdMode: 1, Shape: 0, Weight: 1, Looks: 23010, DuraMax: 1, NeedLevel: 0, Price: 30},
		{Name: "随机卷", StdMode: 1, Shape: 1, Weight: 1, Looks: 23011, DuraMax: 1, NeedLevel: 0, Price: 50},
		{Name: "地牢逃脱卷", StdMode: 1, Shape: 2, Weight: 1, Looks: 23012, DuraMax: 1, NeedLevel: 0, Price: 80},
		
		{Name: "祝福油", StdMode: 1, Shape: 100, Weight: 1, Looks: 23050, DuraMax: 1, NeedLevel: 0, Price: 500},
		{Name: "战神油", StdMode: 1, Shape: 101, Weight: 1, Looks: 23051, DuraMax: 1, NeedLevel: 0, Price: 1000},
		{Name: "修复油", StdMode: 1, Shape: 102, Weight: 1, Looks: 23052, DuraMax: 1, NeedLevel: 0, Price: 300},
	}
}

func FindItem(name string) *ItemAttr {
	for i := range StdItemList {
		if StdItemList[i].Name == name {
			return &StdItemList[i]
		}
	}
	return nil
}

func GetItemByLooks(looks uint16) *ItemAttr {
	for i := range StdItemList {
		if StdItemList[i].Looks == looks {
			return &StdItemList[i]
		}
	}
	return nil
}

func CalcPlayerAbility(p *actor.Player) {
	base := &p.AddAbility
	
	base.AC = 0
	base.MAC = 0
	base.DC = 0
	base.MC = 0
	base.SC = 0
	
	for i := 0; i <= protocol.U_CHARM; i++ {
		item := p.UseItems[i]
		if item.Dura == 0 {
			continue
		}
		
		attr := GetItemByLooks(item.Index)
		if attr == nil {
			continue
		}
		
		base.AC += attr.AC
		base.MAC += attr.MAC
		base.DC += attr.DC
		base.MC += attr.MC
		base.SC += attr.SC
	}
	
	p.MaxHP = p.Ability.MaxHP + p.AddAbility.WHP
	p.MaxMP = p.Ability.MaxMP + p.AddAbility.WMP
	
	switch p.Job {
	case actor.JobWarr:
		p.MaxHP += p.Level * 20
		p.AddAbility.HitPoint = uint16(10 + p.Level/5)
	case actor.JobWizard:
		p.MaxMP += p.Level * 15
		p.AddAbility.SpeedPoint = uint16(10 + p.Level/5)
	case actor.JobTaos:
		p.MaxHP += p.Level * 10
		p.MaxMP += p.Level * 12
	}
}