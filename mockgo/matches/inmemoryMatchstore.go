package matches

type InMemoryMatchstore struct {
	matches    map[string][]*Match
	mismatches []*Mismatch
}

func NewInMemoryMatchstore() *InMemoryMatchstore {
	inMemoryMatchstore := &InMemoryMatchstore{matches: map[string][]*Match{},
		mismatches: []*Mismatch{}}
	return inMemoryMatchstore
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

func (s *InMemoryMatchstore) GetMatchesCount(endpointId string) (int, error) {
	return len(s.matches[endpointId]), nil
}

func (s *InMemoryMatchstore) GetMismatches() ([]*Mismatch, error) {
	return s.mismatches, nil
}
func (s *InMemoryMatchstore) AddMismatches(mismatches []*Mismatch) error {
	s.mismatches = mismatches
	return nil
}
func (s *InMemoryMatchstore) GetMismatchesCount() (int, error) {
	return len(s.mismatches), nil
}

func (s *InMemoryMatchstore) DeleteMatches(endpointId string) error {
	s.matches[endpointId] = nil
	return nil
}
func (s *InMemoryMatchstore) DeleteMismatches() error {
	s.mismatches = []*Mismatch{}
	return nil
}
