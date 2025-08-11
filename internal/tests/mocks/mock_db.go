package mocks

import (
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockDB 模拟数据库
type MockDB struct {
	mock.Mock
}

// First 模拟First方法
func (m *MockDB) First(dest interface{}, conds ...interface{}) *gorm.DB {
	args := m.Called(dest, conds)
	return args.Get(0).(*gorm.DB)
}

// Find 模拟Find方法
func (m *MockDB) Find(dest interface{}, conds ...interface{}) *gorm.DB {
	args := m.Called(dest, conds)
	return args.Get(0).(*gorm.DB)
}

// Create 模拟Create方法
func (m *MockDB) Create(value interface{}) *gorm.DB {
	args := m.Called(value)
	return args.Get(0).(*gorm.DB)
}

// Save 模拟Save方法
func (m *MockDB) Save(value interface{}) *gorm.DB {
	args := m.Called(value)
	return args.Get(0).(*gorm.DB)
}

// Update 模拟Update方法
func (m *MockDB) Update(column string, value interface{}) *gorm.DB {
	args := m.Called(column, value)
	return args.Get(0).(*gorm.DB)
}

// Delete 模拟Delete方法
func (m *MockDB) Delete(value interface{}, conds ...interface{}) *gorm.DB {
	args := m.Called(value, conds)
	return args.Get(0).(*gorm.DB)
}

// Where 模拟Where方法
func (m *MockDB) Where(query interface{}, args ...interface{}) *gorm.DB {
	a := m.Called(query, args)
	return a.Get(0).(*gorm.DB)
}

// Count 模拟Count方法
func (m *MockDB) Count(count *int64) *gorm.DB {
	args := m.Called(count)
	return args.Get(0).(*gorm.DB)
}

// Model 模拟Model方法
func (m *MockDB) Model(value interface{}) *gorm.DB {
	args := m.Called(value)
	return args.Get(0).(*gorm.DB)
}

// Preload 模拟Preload方法
func (m *MockDB) Preload(query string, args ...interface{}) *gorm.DB {
	a := m.Called(query, args)
	return a.Get(0).(*gorm.DB)
}

// Error 返回错误
func (m *MockDB) Error() error {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}
