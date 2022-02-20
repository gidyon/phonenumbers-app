package phoneutils

import (
	"sync"

	"github.com/Pallinder/go-randomdata"
)

type Pagination interface {
	GetBackPageInfo(sessionId, pageToken string) *PageInfo
	SetNextPageInfo(sessionId, currentPageToken, nextPageToken string) *PageInfo
	SessionExist(sessionId string) bool
	SetNewSession(collectionCount int32, pageToken string) string
}

func NewPaginationAPI() Pagination {
	return &paginationStruct{
		mu:              sync.RWMutex{},
		sessionStore:    map[string]map[string]*PageInfo{},
		collectionCount: 0,
	}
}

type paginationStruct struct {
	mu                 sync.RWMutex // guards sessionStore
	sessionStore       map[string]map[string]*PageInfo
	collectionCount    int32
	currentPageCounter int32
}

func (ps *paginationStruct) GetBackPageInfo(sessionId, pageToken string) *PageInfo {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	_, ok := ps.sessionStore[sessionId]
	if ok {
		if v, ok := ps.sessionStore[sessionId][pageToken]; ok {
			return v
		}
	}

	return &PageInfo{
		PageToken:       pageToken,
		PageNumber:      0,
		CollectionCount: ps.collectionCount,
	}
}

func (ps *paginationStruct) SetNextPageInfo(sessionId, currentPageToken, nextPageToken string) *PageInfo {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	_, ok := ps.sessionStore[sessionId]
	if ok {
		ps.currentPageCounter++
		pi := &PageInfo{
			PageToken:       nextPageToken,
			PageNumber:      ps.currentPageCounter,
			CollectionCount: ps.collectionCount,
		}
		ps.sessionStore[sessionId][currentPageToken] = pi
		return pi
	}
	return nil
}

func (ps *paginationStruct) SessionExist(sessionId string) bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	_, ok := ps.sessionStore[sessionId]
	return ok
}

func (ps *paginationStruct) SetNewSession(collectionCount int32, pageToken string) string {
	sessionId := randomdata.RandStringRunes(10)
	ps.mu.Lock()
	ps.collectionCount = collectionCount
	ps.currentPageCounter = 1
	ps.sessionStore[sessionId] = map[string]*PageInfo{
		pageToken: {
			PageNumber: 1,
		},
	}
	ps.mu.Unlock()
	return sessionId
}

type PageInfo struct {
	PageToken       string
	PageNumber      int32
	CollectionCount int32
}
