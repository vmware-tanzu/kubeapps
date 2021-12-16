package mockstore

import (
	"github.com/globalsign/mgo"
	"github.com/kubeapps/kubeapps/pkg/datastore"
	"github.com/stretchr/testify/mock"
)

// MockSession acts as a mock datastore.Session
type mockSession struct {
	*mock.Mock
}

// DB returns a mocked datastore.Database and empty closer function
func (s mockSession) DB() (datastore.Database, func()) {
	return mockDatabase{s.Mock}, func() {}
}

func (s mockSession) Use(name string) datastore.Session {
	return s
}

func (s mockSession) Fsync(async bool) error {
	return nil
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

func (c mockCollection) Bulk() datastore.Bulk {
	return mockBulk{c.Mock}
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
	c.Called(selector, update)
	return nil, nil
}

func (c mockCollection) UpsertId(selector interface{}, update interface{}) (*mgo.ChangeInfo, error) {
	c.Called(selector, update)
	return nil, nil
}

func (c mockCollection) UpdateId(selector, update interface{}) error {
	args := c.Called(selector, update)
	if len(args) > 0 {
		return args.Error(0)
	}
	return nil
}

func (c mockCollection) Remove(selector interface{}) error {
	return nil
}

func (c mockCollection) RemoveAll(selector interface{}) (*mgo.ChangeInfo, error) {
	c.Called(selector)
	return nil, nil
}

func (c mockCollection) Count() (int, error) {
	return 0, nil
}

func (c mockCollection) Pipe(pipeline interface{}) datastore.Pipe {
	return mockPipe{c.Mock}
}

func (c mockCollection) DropCollection() error {
	return nil
}

func (c mockCollection) EnsureIndex(index mgo.Index) error {
	return nil
}

// mockBulk acts as a mock datastore.Bulk
type mockBulk struct {
	*mock.Mock
}

func (b mockBulk) Upsert(pairs ...interface{}) {
	b.Called(pairs)
}

func (b mockBulk) RemoveAll(selectors ...interface{}) {
	b.Called(selectors)
}

func (b mockBulk) Run() (*mgo.BulkResult, error) {
	return nil, nil
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
	if len(args) > 0 {
		return args.Error(0)
	}
	return nil
}

func (q mockQuery) Sort(fields ...string) datastore.Query {
	return q
}

func (q mockQuery) Select(selector interface{}) datastore.Query {
	return q
}

// mockPipe acts as a mock datastore.Pipe
type mockPipe struct {
	*mock.Mock
}

func (p mockPipe) All(result interface{}) error {
	p.Called(result)
	return nil
}

func (p mockPipe) One(result interface{}) error {
	p.Called(result)
	return nil
}

// NewMockSession returns a mocked Session
func NewMockSession(m *mock.Mock) datastore.Session {
	return mockSession{m}
}
