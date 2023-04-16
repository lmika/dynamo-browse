package controllers

import (
	"sync"

	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models"
)

type State struct {
	mutex     *sync.Mutex
	resultSet *models.ResultSet
	filter    string
}

func NewState() *State {
	return &State{
		mutex: new(sync.Mutex),
	}
}

func (s *State) ResultSet() *models.ResultSet {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.resultSet
}

func (s *State) Filter() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.filter
}

func (s *State) withResultSet(rs func(*models.ResultSet)) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	rs(s.resultSet)
}

func (s *State) withResultSetReturningError(rs func(*models.ResultSet) error) (err error) {
	s.withResultSet(func(set *models.ResultSet) {
		err = rs(set)
	})
	return err
}

func (s *State) setResultSetAndFilter(resultSet *models.ResultSet, filter string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.resultSet = resultSet
	s.filter = filter
}

func (s *State) buildNewResultSetMessage(statusMessage string) NewResultSet {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var filteredCount int = 0
	if s.filter != "" {
		for i := range s.resultSet.Items() {
			if !s.resultSet.Hidden(i) {
				filteredCount += 1
			}
		}
	}

	return NewResultSet{s.resultSet, s.filter, filteredCount, statusMessage}
}
