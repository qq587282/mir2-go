package protocol

import "encoding/binary"

type TDefaultMessage struct {
	Recog  int32
	Ident  uint16
	Param  uint16
	Tag    uint16
	Series uint16
}

func (m *TDefaultMessage) Pack() []byte {
	buf := make([]byte, 14)
	binary.LittleEndian.PutUint32(buf[0:4], uint32(m.Recog))
	binary.LittleEndian.PutUint16(buf[4:6], m.Ident)
	binary.LittleEndian.PutUint16(buf[6:8], m.Param)
	binary.LittleEndian.PutUint16(buf[8:10], m.Tag)
	binary.LittleEndian.PutUint16(buf[10:12], m.Series)
	buf[12] = 0
	buf[13] = 0
	return buf
}

func UnpackDefaultMessage(data []byte) *TDefaultMessage {
	if len(data) < 14 {
		return nil
	}
	return &TDefaultMessage{
		Recog:  int32(binary.LittleEndian.Uint32(data[0:4])),
		Ident:  binary.LittleEndian.Uint16(data[4:6]),
		Param:  binary.LittleEndian.Uint16(data[6:8]),
		Tag:    binary.LittleEndian.Uint16(data[8:10]),
		Series: binary.LittleEndian.Uint16(data[10:12]),
	}
}

type TMessageBodyW struct {
	Param1 uint16
	Param2 uint16
	Tag1   uint16
	Tag2   uint16
}

type TMessageBodyWL struct {
	LParam1 int32
	LParam2 int32
	LTag1   int32
	LTag2   int32
}

type TCharDesc struct {
	Feature   int32
	Status    int32
	Level     int32
	HP        int32
	MaxHP     int32
	AddStatus int32
}

func (c *TCharDesc) Pack() []byte {
	buf := make([]byte, 20)
	binary.LittleEndian.PutUint32(buf[0:4], uint32(c.Feature))
	binary.LittleEndian.PutUint32(buf[4:8], uint32(c.Status))
	binary.LittleEndian.PutUint32(buf[8:12], uint32(c.Level))
	binary.LittleEndian.PutUint32(buf[12:16], uint32(c.HP))
	binary.LittleEndian.PutUint32(buf[16:20], uint32(c.MaxHP))
	return buf
}

type THealth struct {
	HP    int32
	MP    int32
	MaxHP int32
}

type TAbility struct {
	Level         int32
	AC            int32
	MAC           int32
	DC            int32
	MC            int32
	SC            int32
	CC            int32
	HP            int32
	MP            int32
	MaxHP         int32
	MaxMP         int32
	Exp           uint32
	MaxExp        uint32
	Weight        uint16
	MaxWeight     uint16
	WearWeight    uint16
	MaxWearWeight uint16
	HandWeight    uint16
	MaxHandWeight uint16
}

func (a *TAbility) Pack() []byte {
	buf := make([]byte, 44)
	binary.LittleEndian.PutUint32(buf[0:4], uint32(a.Level))
	binary.LittleEndian.PutUint32(buf[4:8], uint32(a.AC))
	binary.LittleEndian.PutUint32(buf[8:12], uint32(a.MAC))
	binary.LittleEndian.PutUint32(buf[12:16], uint32(a.DC))
	binary.LittleEndian.PutUint32(buf[16:20], uint32(a.MC))
	binary.LittleEndian.PutUint32(buf[20:24], uint32(a.SC))
	binary.LittleEndian.PutUint32(buf[24:28], uint32(a.CC))
	binary.LittleEndian.PutUint32(buf[28:32], uint32(a.HP))
	binary.LittleEndian.PutUint32(buf[32:36], uint32(a.MP))
	binary.LittleEndian.PutUint32(buf[36:40], uint32(a.MaxHP))
	binary.LittleEndian.PutUint32(buf[40:44], uint32(a.MaxMP))
	return buf
}

type TAddAbility struct {
	WHP           int32
	WMP           int32
	HitPoint      uint16
	SpeedPoint    uint16
	AC            int32
	MAC           int32
	DC            int32
	MC            int32
	SC            int32
	CC            int32
	Holy          byte
	UnHoly        byte
	AntiPoison    uint16
	PoisonRecover uint16
	HealthRecover uint16
	SpellRecover  uint16
	AntiMagic     uint16
	Luck          byte
	UnLuck        byte
	HitSpeed      int32
}

type TUserItem struct {
	MakeIndex int32
	Index     uint16
	Dura      uint16
	DuraMax   uint16
	btValue   [14]byte
	AddValue  [14]byte
	AddPoint  [14]byte
	MaxDate   int64
}

type TClientItem struct {
	s         TStdItem
	MakeIndex int32
	Dura      uint16
	DuraMax   uint16
}

type TStdItem struct {
	Name         [31]byte
	StdMode      byte
	Shape        byte
	Weight       byte
	AniCount     byte
	Source       int8
	Reserved     byte
	NeedIdentify byte
	Looks        uint16
	DuraMax      uint16
	Reserved1    uint16
	AC           int32
	MAC          int32
	DC           int32
	MC           int32
	SC           int32
	CC           int32
	Need         int32
	NeedLevel    int32
	Price        int32
}

type TMagic struct {
	MagicID     uint16
	MagicName   [31]byte
	EffectType  byte
	Effect      byte
	unk1        byte
	Spell       uint16
	Power       uint16
	TrainLevel  [4]byte
	unk2        uint16
	MaxTrain    [4]int32
	TrainLv     byte
	Job         byte
	MagicIdx    uint16
	DelayTime   uint32
	DefSpell    byte
	DefPower    byte
	MaxPower    uint16
	DefMaxPower byte
	Descr       [19]byte
}

type TClientMagic struct {
	Key      byte
	Level    byte
	CurTrain int32
	Def      TMagic
}

type THumMagic struct {
	MagIdx    uint16
	Level     byte
	Key       byte
	TranPoint int32
}

type TQuestFlag [16]byte

type TCond事变 struct {
	Flags [16]byte
}

type TOSessInfo struct {
	Account         [13]byte
	IPaddr          [16]byte
	SessionID       int32
	Payment         int32
	PayMode         int32
	SessionStatus   int32
	StartTick       uint32
	ActiveTick      uint32
	MakeAccountTick uint32
	RefCount        int32
}

type TSessInfo struct {
	Account         string
	IPaddr          [16]byte
	SessionID       int32
	Payment         int32
	PayMode         int32
	SessionStatus   int32
	StartTick       uint32
	ActiveTick      uint32
	MakeAccountTick uint32
	RefCount        int32
}

type TLoadHuman struct {
	Account   [ACCOUNTLEN]byte
	ChrName   [ACTORNAMELEN]byte
	UserAddr  [16]byte
	SessionID int32
}

type TOLoadHuman struct {
	Account   [13]byte
	ChrName   [ACTORNAMELEN]byte
	UserAddr  [16]byte
	SessionID int32
}

type TNakedAbility struct {
	DC    uint16
	MC    uint16
	SC    uint16
	CC    uint16
	AC    uint16
	MAC   uint16
	HP    uint16
	MP    uint16
	Hit   uint16
	Speed uint16
	X2    uint16
}

type TMapInfo struct {
	Name            string
	MapNO           string
	L               int32
	ServerIndex     int32
	NEEDONOFFFlag   int32
	boNEEDONOFFFlag bool
	ShowName        string
	ReConnectMap    string
	SAFE            bool
	DARK            bool
	FIGHT           bool
	FIGHT3          bool
	DAY             bool
	QUIZ            bool
	NORECONNECT     bool
	NEEDHOLE        bool
	NORECALL        bool
	NORANDOMMOVE    bool
	NODRUG          bool
	MINE            bool
	NOPOSITIONMOVE  bool
}

type TMerchantInfo struct {
	Script  [15]byte
	MapName [15]byte
	X       int32
	Y       int32
	NPCName [41]byte
	Face    int32
	Body    int32
	Castle  bool
}

type TSendMessage struct {
	wIdent         uint16
	wParam         int32
	nParam1        int32
	nParam2        int32
	nParam3        int32
	dwAddTime      uint32
	dwDeliveryTime uint32
	boLateDelivery bool
	Buff           []byte
}

type TProcessMessage struct {
	wIdent         uint16
	wParam         int32
	nParam1        int32
	nParam2        int32
	nParam3        int32
	nParam4        int32
	nParam5        int32
	nParam6        int32
	boLateDelivery bool
	dwDeliveryTime uint32
	sMsg           string
}

type TClientPacket struct {
	Ident   uint16
	Socket  int32
	Socket2 int32
	Buff    [256]byte
	DefMsg  TDefaultMessage
	Length  int
}

type TServerPacket struct {
	Ident  uint16
	Socket int32
	Buff   [8192]byte
	DefMsg TDefaultMessage
	Length int
}

type TCheckOnline struct {
	Account string
	ChrName string
	Result  int32
}

type TUserInfo struct {
	Account      string
	IPAddress    string
	ConnectionID int32
	SessionID    int32
	Ready        bool
}

const (
	DB_LOADCHAR    = 1
	DB_SAVECHAR    = 2
	DB_CREATECHAR  = 3
	DB_DELCHAR     = 4
	DB_GETCHARLIST = 5
	DB_REFRESHCHAR = 6
)

const (
	LG_QUERY_NAME = 1
	LG_QUERY_PASS = 2
	LG_LOGIN_OK   = 3
	LG_LOGIN_FAIL = 4
	LG_CLOSE      = 5
)

const (
	SS_SERVERINFO  = 1
	SS_KICKUSER    = 2
	SS_USERSOCKET  = 3
	SS_RELOADGUILD = 10
)

type TServerInfo struct {
	ServerID    int32
	ServerName  string
	ServerIP    string
	Port        int32
	PlayerCount int32
	ServerURL   string
}
