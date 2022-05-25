package controllers

import (
	"sync"

	"github.com/lmika/awstools/internal/dynamo-browse/models"
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

func (s *State) setResultSetAndFilter(resultSet *models.ResultSet, filter string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.resultSet = resultSet
	s.filter = filter
}
