package db

import (
	"encoding/gob"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/mir2go/mir2/pkg/game/actor"
)

type Account struct {
	AccountID   int32
	LoginID     string
	LoginPW     string
	UserName    string
	SSNo        string
	PhoneNum    string
	Quiz        string
	Answer      string
	EMail       string
	JoinDate    time.Time
	LastLogin   time.Time
	IPAddr      string
	SafeCode    string
	BlockDate   time.Time
	FailedCount int
	Admin       int
}

type Character struct {
	CharID       int32
	AccountID    int32
	Name         string
	Job          actor.Job
	Gender       actor.Gender
	Level        int32
	Gold         int32
	MapName      string
	X, Y         int
	HP, MP       int32
	Exp          uint32
	Hair         byte
	Clothes      uint16
	Weapon       uint16
	Direction    byte
	PKPoint      int32
	DeleteTime   time.Time
	CreateTime   time.Time
	LastLogin    time.Time
	LoginStatus  int
}

type TAccountDB struct {
	Accounts map[string]*Account
	ByID     map[int32]*Account
	Mutex    sync.RWMutex
	FileName string
}

type TCharacterDB struct {
	Characters map[int32]*Character
	ByName     map[string]*Character
	Mutex      sync.RWMutex
	FileName   string
}

func NewAccountDB(filename string) *TAccountDB {
	db := &TAccountDB{
		Accounts: make(map[string]*Account),
		ByID:     make(map[int32]*Account),
		FileName: filename,
	}
	db.Load()
	return db
}

func (db *TAccountDB) Load() error {
	if db.FileName == "" {
		return nil
	}
	
	file, err := os.Open(db.FileName)
	if err != nil {
		return nil
	}
	defer file.Close()
	
	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(&db.Accounts); err != nil {
		return err
	}
	
	db.ByID = make(map[int32]*Account)
	for _, acc := range db.Accounts {
		db.ByID[acc.AccountID] = acc
	}
	
	return nil
}

func (db *TAccountDB) Save() error {
	if db.FileName == "" {
		return nil
	}
	
	file, err := os.Create(db.FileName)
	if err != nil {
		return err
	}
	defer file.Close()
	
	encoder := gob.NewEncoder(file)
	return encoder.Encode(db.Accounts)
}

func (db *TAccountDB) FindAccount(loginID string) *Account {
	db.Mutex.RLock()
	defer db.Mutex.RUnlock()
	return db.Accounts[loginID]
}

func (db *TAccountDB) AddAccount(acc *Account) error {
	db.Mutex.Lock()
	defer db.Mutex.Unlock()
	
	if _, exists := db.Accounts[acc.LoginID]; exists {
		return fmt.Errorf("account already exists: %s", acc.LoginID)
	}
	
	acc.AccountID = int32(len(db.Accounts) + 1)
	db.Accounts[acc.LoginID] = acc
	db.ByID[acc.AccountID] = acc
	
	return db.Save()
}

func (db *TAccountDB) UpdateAccount(acc *Account) error {
	db.Mutex.Lock()
	defer db.Mutex.Unlock()
	
	db.Accounts[acc.LoginID] = acc
	db.ByID[acc.AccountID] = acc
	
	return db.Save()
}

func NewCharacterDB(filename string) *TCharacterDB {
	db := &TCharacterDB{
		Characters: make(map[int32]*Character),
		ByName:     make(map[string]*Character),
		FileName:   filename,
	}
	db.Load()
	return db
}

func (db *TCharacterDB) Load() error {
	if db.FileName == "" {
		return nil
	}
	
	file, err := os.Open(db.FileName)
	if err != nil {
		return nil
	}
	defer file.Close()
	
	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(&db.Characters); err != nil {
		return err
	}
	
	db.ByName = make(map[string]*Character)
	for _, ch := range db.Characters {
		db.ByName[ch.Name] = ch
	}
	
	return nil
}

func (db *TCharacterDB) Save() error {
	if db.FileName == "" {
		return nil
	}
	
	file, err := os.Create(db.FileName)
	if err != nil {
		return err
	}
	defer file.Close()
	
	encoder := gob.NewEncoder(file)
	return encoder.Encode(db.Characters)
}

func (db *TCharacterDB) FindCharacter(charID int32) *Character {
	db.Mutex.RLock()
	defer db.Mutex.RUnlock()
	return db.Characters[charID]
}

func (db *TCharacterDB) FindCharacterByName(name string) *Character {
	db.Mutex.RLock()
	defer db.Mutex.RUnlock()
	return db.ByName[name]
}

func (db *TCharacterDB) GetCharactersByAccount(accountID int32) []*Character {
	db.Mutex.RLock()
	defer db.Mutex.RUnlock()
	
	var result []*Character
	for _, ch := range db.Characters {
		if ch.AccountID == accountID {
			result = append(result, ch)
		}
	}
	return result
}

func (db *TCharacterDB) AddCharacter(ch *Character) error {
	db.Mutex.Lock()
	defer db.Mutex.Unlock()
	
	if _, exists := db.ByName[ch.Name]; exists {
		return fmt.Errorf("character name already exists: %s", ch.Name)
	}
	
	ch.CharID = int32(len(db.Characters) + 1)
	db.Characters[ch.CharID] = ch
	db.ByName[ch.Name] = ch
	
	return db.Save()
}

func (db *TCharacterDB) UpdateCharacter(ch *Character) error {
	db.Mutex.Lock()
	defer db.Mutex.Unlock()
	
	db.Characters[ch.CharID] = ch
	db.ByName[ch.Name] = ch
	
	return db.Save()
}

func (db *TCharacterDB) DeleteCharacter(charID int32) error {
	db.Mutex.Lock()
	defer db.Mutex.Unlock()
	
	ch, exists := db.Characters[charID]
	if !exists {
		return fmt.Errorf("character not found: %d", charID)
	}
	
	ch.DeleteTime = time.Now()
	delete(db.ByName, ch.Name)
	
	return nil
}

type TGameDataDB struct {
	MonsterDB    map[string]interface{}
	MagicDB      map[string]interface{}
	ItemDB       map[string]interface{}
	QuestDB      map[string]interface{}
	MerchantDB  map[string]interface{}
	Mutex        sync.RWMutex
}

func NewGameDataDB() *TGameDataDB {
	return &TGameDataDB{
		MonsterDB:   make(map[string]interface{}),
		MagicDB:    make(map[string]interface{}),
		ItemDB:     make(map[string]interface{}),
		QuestDB:    make(map[string]interface{}),
		MerchantDB: make(map[string]interface{}),
	}
}

func (db *TGameDataDB) LoadMonsterDB() error {
	db.Mutex.Lock()
	defer db.Mutex.Unlock()
	return nil
}

func (db *TGameDataDB) LoadMagicDB() error {
	db.Mutex.Lock()
	defer db.Mutex.Unlock()
	return nil
}

func (db *TGameDataDB) LoadItemsDB() error {
	db.Mutex.Lock()
	defer db.Mutex.Unlock()
	return nil
}

func (db *TGameDataDB) LoadMerchant() error {
	db.Mutex.Lock()
	defer db.Mutex.Unlock()
	return nil
}

func (db *TGameDataDB) LoadQuestDiary() error {
	db.Mutex.Lock()
	defer db.Mutex.Unlock()
	return nil
}

func (db *TGameDataDB) LoadMapQuest() error {
	db.Mutex.Lock()
	defer db.Mutex.Unlock()
	return nil
}

func (db *TGameDataDB) LoadAdminList() (map[string]string, error) {
	db.Mutex.RLock()
	defer db.Mutex.RUnlock()
	result := make(map[string]string)
	return result, nil
}

func (db *TGameDataDB) LoadMonGen() error {
	db.Mutex.Lock()
	defer db.Mutex.Unlock()
	return nil
}

func (db *TGameDataDB) LoadUnbindList() error {
	db.Mutex.Lock()
	defer db.Mutex.Unlock()
	return nil
}

func (db *TGameDataDB) GetMonster(name string) interface{} {
	db.Mutex.RLock()
	defer db.Mutex.RUnlock()
	return db.MonsterDB[name]
}

func (db *TGameDataDB) GetMagic(name string) interface{} {
	db.Mutex.RLock()
	defer db.Mutex.RUnlock()
	return db.MagicDB[name]
}

func (db *TGameDataDB) GetItem(name string) interface{} {
	db.Mutex.RLock()
	defer db.Mutex.RUnlock()
	return db.ItemDB[name]
}
