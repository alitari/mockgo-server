package matches

import (
	"container/list"
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

func (s *InMemoryMatchstore) GetMatches(endpointID string) ([]*Match, error) {
	matchesResult := []*Match{}
	matchesList := s.matches[endpointID]
	if matchesList != nil {
		for match := matchesList.Front(); match != nil; match = match.Next() {
			matchesResult = append(matchesResult, match.Value.(*Match))
		}
	}
	return matchesResult, nil
}

func (s *InMemoryMatchstore) GetMatchesCount(endpointID string) (uint64, error) {
	return s.matchesCount[endpointID], nil
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

func (s *InMemoryMatchstore) AddMatch(endpointID string, match *Match) error {
	if s.matches[endpointID] == nil {
		s.matches[endpointID] = list.New()
	}
	matches := s.matches[endpointID]
	matches.PushBack(match)
	if uint16(matches.Len()) > s.size {
		matches.Remove(matches.Front())
	}
	s.matchesCount[endpointID]++
	return nil
}

func (s *InMemoryMatchstore) GetMismatchesCount() (uint64, error) {
	return s.mismatchesCount, nil
}

func (s *InMemoryMatchstore) DeleteMatches(endpointID string) error {
	matchesList := s.matches[endpointID]
	if matchesList != nil {
		for match := matchesList.Front(); match != nil; match = matchesList.Front() {
			matchesList.Remove(match)
		}
	}
	s.matchesCount[endpointID] = uint64(0)
	return nil
}

func (s *InMemoryMatchstore) DeleteMismatches() error {
	for mismatch := s.mismatches.Front(); mismatch != nil; mismatch = s.mismatches.Front() {
		s.mismatches.Remove(mismatch)
	}
	s.mismatchesCount = uint64(0)
	return nil
}
