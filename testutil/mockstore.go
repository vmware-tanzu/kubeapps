package testutil

import (
	"github.com/kubeapps/common/datastore"
	"github.com/stretchr/testify/mock"
	mgo "gopkg.in/mgo.v2"
)

// MockSession acts as a mock datastore.Session
type mockSession struct {
	*mock.Mock
}

// DB returns a mocked datastore.Database and empty closer function
func (s mockSession) DB() (datastore.Database, func()) {
	return mockDatabase{s.Mock}, func() {}
}

// mockDatabase acts as a mock datastore.Database
type mockDatabase struct {
	*mock.Mock
}

func (d mockDatabase) C(name string) datastore.Collection {
	return mockCollection{d.Mock}
}

// mockCollection acts as a mock datastore.Collection
type mockCollection struct {
	*mock.Mock
}

func (c mockCollection) Find(query interface{}) datastore.Query {
	return mockQuery{c.Mock}
}

func (c mockCollection) FindId(id interface{}) datastore.Query {
	return mockQuery{c.Mock}
}

func (c mockCollection) Insert(docs ...interface{}) error {
	c.Called(docs...)
	return nil
}

func (c mockCollection) Upsert(selector interface{}, update interface{}) (*mgo.ChangeInfo, error) {
	return nil, nil
}

func (c mockCollection) UpsertId(selector interface{}, update interface{}) (*mgo.ChangeInfo, error) {
	return nil, nil
}

func (c mockCollection) UpdateId(selector, update interface{}) error {
	c.Called(selector, update)
	return nil
}

func (c mockCollection) Remove(selector interface{}) error {
	return nil
}

func (c mockCollection) Count() (int, error) {
	return 0, nil
}

// mockQuery acts as a mock datastore.Query
type mockQuery struct {
	*mock.Mock
}

func (q mockQuery) All(result interface{}) error {
	q.Called(result)
	return nil
}

func (q mockQuery) One(result interface{}) error {
	args := q.Called(result)
	return args.Error(0)
}

// NewMockSession returns a mocked Session
func NewMockSession(m *mock.Mock) datastore.Session {
	return mockSession{m}
}
