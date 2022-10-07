package matches

import (
	"container/list"
	"fmt"
)

type InMemoryMatchstore struct {
	size            uint16
	matches         map[string]*list.List
	mismatches      *list.List
	matchesCount    map[string]uint64
	mismatchesCount uint64
}

func NewInMemoryMatchstore(size uint16) *InMemoryMatchstore {
	list.New()
	inMemoryMatchstore := &InMemoryMatchstore{matches: map[string]*list.List{}, mismatches: list.New(), matchesCount: map[string]uint64{}, size: size}
	return inMemoryMatchstore
}

func (s *InMemoryMatchstore) GetAll() (map[string][]*Match, []*Mismatch, map[string]uint64, uint64) {
	matchesResult := map[string][]*Match{}
	for endpoint, matchList := range s.matches {
		for match := matchList.Front(); match != nil; match = match.Next() {
			matchesResult[endpoint] = append(matchesResult[endpoint], match.Value.(*Match))
		}
	}

	mismatchesResult := []*Mismatch{}
	for mismatch := s.mismatches.Front(); mismatch != nil; mismatch = mismatch.Next() {
		mismatchesResult = append(mismatchesResult, mismatch.Value.(*Mismatch))
	}
	return matchesResult, mismatchesResult, s.matchesCount, s.mismatchesCount
}

func (s *InMemoryMatchstore) GetMatches(endpointId string) ([]*Match, error) {
	matchesResult := []*Match{}
	matchesList := s.matches[endpointId]
	if matchesList != nil {
		for match := matchesList.Front(); match != nil; match = match.Next() {
			matchesResult = append(matchesResult, match.Value.(*Match))
		}
	}
	return matchesResult, nil
}

func (s *InMemoryMatchstore) Transfer() error {
	return fmt.Errorf("transfer not supported")
}

func (s *InMemoryMatchstore) GetMatchesCount(endpointId string) (uint64, error) {
	return s.matchesCount[endpointId], nil
}

func (s *InMemoryMatchstore) GetMismatches() ([]*Mismatch, error) {
	mismatchesResult := []*Mismatch{}
	for mismatch := s.mismatches.Front(); mismatch != nil; mismatch = mismatch.Next() {
		mismatchesResult = append(mismatchesResult, mismatch.Value.(*Mismatch))
	}
	return mismatchesResult, nil
}

func (s *InMemoryMatchstore) AddMismatch(mismatch *Mismatch) error {
	s.mismatches.PushBack(mismatch)
	if uint16(s.mismatches.Len()) > s.size {
		s.mismatches.Remove(s.mismatches.Front())
	}
	s.mismatchesCount++
	return nil
}

func (s *InMemoryMatchstore) AddMatch(endpointId string, match *Match) error {
	if s.matches[endpointId] == nil {
		s.matches[endpointId] = list.New()
	}
	matches := s.matches[endpointId]
	matches.PushBack(match)
	if uint16(matches.Len()) > s.size {
		matches.Remove(matches.Front())
	}
	s.matchesCount[endpointId]++
	return nil
}

func (s *InMemoryMatchstore) GetMismatchesCount() (uint64, error) {
	return s.mismatchesCount, nil
}

func (s *InMemoryMatchstore) DeleteMatches(endpointId string) error {
	matchesList := s.matches[endpointId]
	if matchesList != nil {
		for match := matchesList.Front(); match != nil; match = matchesList.Front() {
			matchesList.Remove(match)
		}
	}
	s.matchesCount[endpointId] = uint64(0)
	return nil
}

func (s *InMemoryMatchstore) DeleteMismatches() error {
	for mismatch := s.mismatches.Front(); mismatch != nil; mismatch = s.mismatches.Front() {
		s.mismatches.Remove(mismatch)
	}
	s.mismatchesCount = uint64(0)
	return nil
}
