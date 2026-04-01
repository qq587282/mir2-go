package actor

import (
	"sync"
	"time"
)

type TradeState int

const (
	TRADE_NONE      TradeState = 0
	TRADE_WAITING   TradeState = 1
	TRADE_ACCEPTED  TradeState = 2
	TRADE_LOCKED    TradeState = 3
	TRADE_COMPLETED TradeState = 4
	TRADE_CANCELLED TradeState = 5
)

type TTradeInfo struct {
	TradeID       int32
	State        TradeState
	Player1      *Player
	Player2      *Player
	
	Player1Items  []*TItem
	Player2Items  []*TItem
	Player1Gold   int32
	Player2Gold   int32
	
	Player1Locked bool
	Player2Locked bool
	
	CreateTime    time.Time
	CompleteTime  time.Time
}

type TradeManager struct {
	Trades    map[int32]*TTradeInfo
	Mutex     sync.RWMutex
	NextID    int32
}

var DefaultTradeManager *TradeManager

func init() {
	DefaultTradeManager = NewTradeManager()
}

func NewTradeManager() *TradeManager {
	return &TradeManager{
		Trades: make(map[int32]*TTradeInfo),
		NextID:  1,
	}
}

func (tm *TradeManager) GetNextID() int32 {
	id := tm.NextID
	tm.NextID++
	return id
}

func (tm *TradeManager) CreateTrade(player1, player2 *Player) (*TTradeInfo, error) {
	if player1 == nil || player2 == nil {
		return nil, ErrInvalidPlayer
	}
	
	if player1.TradeInfo != nil || player2.TradeInfo != nil {
		return nil, ErrAlreadyInTrade
	}
	
	if player1.GetMapName() != player2.GetMapName() {
		return nil, ErrNotSameMap
	}
	
	dist := Distance(player1.GetX(), player1.GetY(), player2.GetX(), player2.GetY())
	if dist > 3 {
		return nil, ErrTooFarAway
	}
	
	tm.Mutex.Lock()
	defer tm.Mutex.Unlock()
	
	trade := &TTradeInfo{
		TradeID:    tm.GetNextID(),
		State:      TRADE_WAITING,
		Player1:    player1,
		Player2:    player2,
		CreateTime: time.Now(),
	}
	
	trade.Player1Items = make([]*TItem, 0, 12)
	trade.Player2Items = make([]*TItem, 0, 12)
	
	player1.TradeInfo = trade
	player2.TradeInfo = trade
	
	tm.Trades[trade.TradeID] = trade
	
	return trade, nil
}

func (tm *TradeManager) AcceptTrade(trade *TTradeInfo) error {
	if trade == nil {
		return ErrInvalidTrade
	}
	
	if trade.State != TRADE_WAITING {
		return ErrInvalidState
	}
	
	trade.State = TRADE_ACCEPTED
	return nil
}

func (tm *TradeManager) AddItem(trade *TTradeInfo, player *Player, item *TItem) error {
	if trade == nil {
		return ErrInvalidTrade
	}
	
	if trade.State != TRADE_ACCEPTED {
		if trade.State != TRADE_WAITING {
			return ErrInvalidState
		}
	}
	
	if player != trade.Player1 && player != trade.Player2 {
		return ErrNotInTrade
	}
	
	var items *[]*TItem
	if player == trade.Player1 {
		items = &trade.Player1Items
		if trade.Player1Locked {
			return ErrAlreadyLocked
		}
	} else {
		items = &trade.Player2Items
		if trade.Player2Locked {
			return ErrAlreadyLocked
		}
	}
	
	if len(*items) >= 12 {
		return ErrTooManyItems
	}
	
	for _, existing := range *items {
		if existing.MakeIndex == item.MakeIndex {
			return ErrItemAlreadyAdded
		}
	}
	
	*items = append(*items, item)
	return nil
}

func (tm *TradeManager) RemoveItem(trade *TTradeInfo, player *Player, makeIndex int32) error {
	if trade == nil {
		return ErrInvalidTrade
	}
	
	if trade.State != TRADE_ACCEPTED {
		return ErrInvalidState
	}
	
	if player != trade.Player1 && player != trade.Player2 {
		return ErrNotInTrade
	}
	
	var items *[]*TItem
	var locked *bool
	if player == trade.Player1 {
		items = &trade.Player1Items
		locked = &trade.Player1Locked
	} else {
		items = &trade.Player2Items
		locked = &trade.Player2Locked
	}
	
	if *locked {
		return ErrAlreadyLocked
	}
	
	for i, item := range *items {
		if item.MakeIndex == makeIndex {
			*items = append((*items)[:i], (*items)[i+1:]...)
			return nil
		}
	}
	
	return ErrItemNotFound
}

func (tm *TradeManager) AddGold(trade *TTradeInfo, player *Player, gold int32) error {
	if trade == nil {
		return ErrInvalidTrade
	}
	
	if gold <= 0 {
		return ErrInvalidGold
	}
	
	if player != trade.Player1 && player != trade.Player2 {
		return ErrNotInTrade
	}
	
	if player == trade.Player1 {
		if trade.Player1Locked {
			return ErrAlreadyLocked
		}
		if gold > player.Gold {
			return ErrNotEnoughGold
		}
		trade.Player1Gold += gold
		player.Gold -= gold
	} else {
		if trade.Player2Locked {
			return ErrAlreadyLocked
		}
		if gold > player.Gold {
			return ErrNotEnoughGold
		}
		trade.Player2Gold += gold
		player.Gold -= gold
	}
	
	return nil
}

func (tm *TradeManager) LockTrade(trade *TTradeInfo, player *Player) error {
	if trade == nil {
		return ErrInvalidTrade
	}
	
	if trade.State != TRADE_ACCEPTED {
		return ErrInvalidState
	}
	
	if player != trade.Player1 && player != trade.Player2 {
		return ErrNotInTrade
	}
	
	if player == trade.Player1 {
		if trade.Player1Locked {
			return ErrAlreadyLocked
		}
		trade.Player1Locked = true
	} else {
		if trade.Player2Locked {
			return ErrAlreadyLocked
		}
		trade.Player2Locked = true
	}
	
	if trade.Player1Locked && trade.Player2Locked {
		return tm.CompleteTrade(trade)
	}
	
	return nil
}

func (tm *TradeManager) CompleteTrade(trade *TTradeInfo) error {
	if trade == nil {
		return ErrInvalidTrade
	}
	
	if !trade.Player1Locked || !trade.Player2Locked {
		return ErrNotLocked
	}
	
	for _, item := range trade.Player1Items {
		if !trade.Player2.AddItem(item) {
			trade.Player1.AddItem(item)
			trade.Player1Gold += trade.Player1Gold
			trade.Player2Gold += trade.Player2Gold
			tm.CancelTrade(trade)
			return ErrPlayerBagFull
		}
	}
	
	for _, item := range trade.Player2Items {
		if !trade.Player1.AddItem(item) {
			trade.Player2.AddItem(item)
			trade.Player1Gold += trade.Player1Gold
			trade.Player2Gold += trade.Player2Gold
			tm.CancelTrade(trade)
			return ErrPlayerBagFull
		}
	}
	
	trade.Player1.Gold += trade.Player2Gold
	trade.Player2.Gold += trade.Player1Gold
	
	trade.State = TRADE_COMPLETED
	trade.CompleteTime = time.Now()
	
	trade.Player1.TradeInfo = nil
	trade.Player2.TradeInfo = nil
	
	tm.Mutex.Lock()
	delete(tm.Trades, trade.TradeID)
	tm.Mutex.Unlock()
	
	return nil
}

func (tm *TradeManager) CancelTrade(trade *TTradeInfo) error {
	if trade == nil {
		return nil
	}
	
	for _, item := range trade.Player1Items {
		trade.Player1.AddItem(item)
	}
	
	for _, item := range trade.Player2Items {
		trade.Player2.AddItem(item)
	}
	
	trade.Player1.Gold += trade.Player1Gold
	trade.Player2.Gold += trade.Player2Gold
	
	trade.State = TRADE_CANCELLED
	
	if trade.Player1 != nil {
		trade.Player1.TradeInfo = nil
	}
	if trade.Player2 != nil {
		trade.Player2.TradeInfo = nil
	}
	
	tm.Mutex.Lock()
	delete(tm.Trades, trade.TradeID)
	tm.Mutex.Unlock()
	
	return nil
}

func (tm *TradeManager) GetTrade(tradeID int32) *TTradeInfo {
	tm.Mutex.RLock()
	defer tm.Mutex.RUnlock()
	return tm.Trades[tradeID]
}

func (p *Player) StartTrade(other *Player) (*TTradeInfo, error) {
	return DefaultTradeManager.CreateTrade(p, other)
}

func (p *Player) AcceptTrade() error {
	if p.TradeInfo == nil {
		return ErrNotInTrade
	}
	trade, _ := p.TradeInfo.(*TTradeInfo)
	return DefaultTradeManager.AcceptTrade(trade)
}

func (p *Player) AddTradeItem(item *TItem) error {
	if p.TradeInfo == nil {
		return ErrNotInTrade
	}
	trade, _ := p.TradeInfo.(*TTradeInfo)
	return DefaultTradeManager.AddItem(trade, p, item)
}

func (p *Player) RemoveTradeItem(makeIndex int32) error {
	if p.TradeInfo == nil {
		return ErrNotInTrade
	}
	trade, _ := p.TradeInfo.(*TTradeInfo)
	return DefaultTradeManager.RemoveItem(trade, p, makeIndex)
}

func (p *Player) AddTradeGold(gold int32) error {
	if p.TradeInfo == nil {
		return ErrNotInTrade
	}
	trade, _ := p.TradeInfo.(*TTradeInfo)
	return DefaultTradeManager.AddGold(trade, p, gold)
}

func (p *Player) LockTrade() error {
	if p.TradeInfo == nil {
		return ErrNotInTrade
	}
	trade, _ := p.TradeInfo.(*TTradeInfo)
	return DefaultTradeManager.LockTrade(trade, p)
}

func (p *Player) CancelTrade() error {
	if p.TradeInfo == nil {
		return nil
	}
	trade, _ := p.TradeInfo.(*TTradeInfo)
	return DefaultTradeManager.CancelTrade(trade)
}

func (p *Player) GetTradePartner() *Player {
	if p.TradeInfo == nil {
		return nil
	}
	trade, ok := p.TradeInfo.(*TTradeInfo)
	if !ok {
		return nil
	}
	if trade.Player1 == p {
		return trade.Player2
	}
	return trade.Player1
}

var (
	ErrInvalidPlayer     = &TradeError{"Invalid player"}
	ErrAlreadyInTrade   = &TradeError{"Already in trade"}
	ErrNotSameMap       = &TradeError{"Not in same map"}
	ErrTooFarAway       = &TradeError{"Too far away"}
	ErrInvalidTrade     = &TradeError{"Invalid trade"}
	ErrInvalidState     = &TradeError{"Invalid state"}
	ErrNotInTrade       = &TradeError{"Not in trade"}
	ErrTooManyItems     = &TradeError{"Too many items"}
	ErrItemAlreadyAdded = &TradeError{"Item already added"}
	ErrAlreadyLocked    = &TradeError{"Already locked"}
	ErrInvalidGold      = &TradeError{"Invalid gold amount"}
	ErrNotEnoughGold    = &TradeError{"Not enough gold"}
	ErrItemNotFound     = &TradeError{"Item not found"}
	ErrNotLocked        = &TradeError{"Not locked by both"}
	ErrPlayerBagFull    = &TradeError{"Player bag is full"}
)

type TradeError struct {
	Message string
}

func (e *TradeError) Error() string {
	return e.Message
}
