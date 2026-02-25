package services

import (
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockDatabase 是一个简单的数据库 mock
type MockDatabase struct {
	mock.Mock
}

// First 模拟 gorm 的 First 方法
func (m *MockDatabase) First(dest interface{}, conds ...interface{}) *gorm.DB {
	args := m.Called(dest, conds)
	if args.Get(0) != nil {
		return args.Get(0).(*gorm.DB)
	}
	return &gorm.DB{}
}

// Where 模拟 gorm 的 Where 方法
func (m *MockDatabase) Where(query interface{}, args ...interface{}) *gorm.DB {
	mockArgs := m.Called(query, args)
	if mockArgs.Get(0) != nil {
		return mockArgs.Get(0).(*gorm.DB)
	}
	return &gorm.DB{}
}

// Create 模拟 gorm 的 Create 方法
func (m *MockDatabase) Create(value interface{}) *gorm.DB {
	args := m.Called(value)
	if args.Get(0) != nil {
		return args.Get(0).(*gorm.DB)
	}
	return &gorm.DB{}
}

// Model 模拟 gorm 的 Model 方法
func (m *MockDatabase) Model(value interface{}) *gorm.DB {
	args := m.Called(value)
	if args.Get(0) != nil {
		return args.Get(0).(*gorm.DB)
	}
	return &gorm.DB{}
}

// Update 模拟 gorm 的 Update 方法
func (m *MockDatabase) Update(column string, value interface{}) *gorm.DB {
	args := m.Called(column, value)
	if args.Get(0) != nil {
		return args.Get(0).(*gorm.DB)
	}
	return &gorm.DB{}
}
