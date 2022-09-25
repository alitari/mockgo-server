package matches

type InMemoryMatchstore struct {
	matchesCountOnly    bool
	mismatchesCountOnly bool
	matches             map[string][]*Match
	matchesCount        map[string]int64
	mismatches          []*Mismatch
	mismatchesCount     int64
}

func NewInMemoryMatchstore(matchesCountOnly, mismatchesCountOnly bool) *InMemoryMatchstore {
	inMemoryMatchstore := &InMemoryMatchstore{matchesCountOnly: matchesCountOnly, mismatchesCountOnly: mismatchesCountOnly,
		matches: make(map[string][]*Match), matchesCount: map[string]int64{}, mismatches: make([]*Mismatch, 0), mismatchesCount: 0}
	return inMemoryMatchstore
}

func (s *InMemoryMatchstore) HasMatchesCountOnly() bool {
	return s.matchesCountOnly
}
func (s *InMemoryMatchstore) HasMismatchesCountOnly() bool {
	return s.mismatchesCountOnly
}
func (s *InMemoryMatchstore) GetMatches(endpointId string) ([]*Match, error) {
	return s.matches[endpointId], nil
}

func (s *InMemoryMatchstore) AddMatches(matches map[string][]*Match) error {
	for k, v := range matches {
		s.matches[k] = append(s.matches[k], v...)
	}
	return nil
}

func (s *InMemoryMatchstore) GetMatchesCount(endpointId string) (int64, error) {
	return s.matchesCount[endpointId], nil
}

func (s *InMemoryMatchstore) AddMatchesCount(matchesCount map[string]int64) error {
	for k, v := range matchesCount {
		s.matchesCount[k] = s.matchesCount[k] + v
	}
	return nil
}

func (s *InMemoryMatchstore) GetMismatches() ([]*Mismatch, error) {
	return s.mismatches, nil
}
func (s *InMemoryMatchstore) AddMismatches(mismatches []*Mismatch) error {
	s.mismatches = mismatches
	return nil
}
func (s *InMemoryMatchstore) GetMismatchesCount() (int64, error) {
	return s.mismatchesCount, nil
}
func (s *InMemoryMatchstore) AddMismatchesCount(mismatchesCount int64) error {
	s.mismatchesCount = mismatchesCount
	return nil
}
func (s *InMemoryMatchstore) DeleteMatches() error {
	s.matches = make(map[string][]*Match)
	s.matchesCount = make(map[string]int64)
	return nil
}
func (s *InMemoryMatchstore) DeleteMismatches() error {
	s.mismatches = []*Mismatch{}
	s.mismatchesCount = 0
	return nil
}
