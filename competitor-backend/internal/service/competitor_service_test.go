package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"competitor-backend/internal/model"
)

type MockCompetitorRepo struct {
	mock.Mock
}

func (m *MockCompetitorRepo) Create(c *model.Competitor) error {
	args := m.Called(c)
	return args.Error(0)
}

func (m *MockCompetitorRepo) GetByID(id string) (*model.Competitor, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Competitor), args.Error(1)
}

func (m *MockCompetitorRepo) GetByName(name string) (*model.Competitor, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Competitor), args.Error(1)
}

func TestCreateCompetitor_Success(t *testing.T) {
	mockRepo := new(MockCompetitorRepo)
	svc := NewCompetitorService(mockRepo)

	input := &CreateCompetitorInput{
		Name:     "DJI Osmo Action 4",
		Website:  "https://www.dji.com",
		Category: "运动相机",
	}

	mockRepo.On("GetByName", input.Name).Return(nil, assert.AnError)
	mockRepo.On("Create", mock.AnythingOfType("*model.Competitor")).Return(nil)

	result, err := svc.Create(input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "DJI Osmo Action 4", result.Name)
	assert.Equal(t, "运动相机", result.Category)
	assert.Equal(t, "active", result.Status)
	mockRepo.AssertExpectations(t)
}

func TestCreateCompetitor_MissingName(t *testing.T) {
	svc := NewCompetitorService(nil)

	input := &CreateCompetitorInput{
		Website:  "https://www.dji.com",
		Category: "运动相机",
	}

	result, err := svc.Create(input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "竞品名称不能为空")
}

func TestCreateCompetitor_DuplicateName(t *testing.T) {
	mockRepo := new(MockCompetitorRepo)
	svc := NewCompetitorService(mockRepo)

	input := &CreateCompetitorInput{
		Name:     "DJI Action 4",
		Category: "运动相机",
	}

	mockRepo.On("GetByName", input.Name).Return(&model.Competitor{
		ID:   "cmpt_001",
		Name: "DJI Action 4",
	}, nil)

	result, err := svc.Create(input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "竞品已存在")
}

func TestGetCompetitorByID_Success(t *testing.T) {
	mockRepo := new(MockCompetitorRepo)
	svc := NewCompetitorService(mockRepo)

	expected := &model.Competitor{
		ID:       "cmpt_001",
		Name:     "DJI Action 4",
		Category: "运动相机",
	}

	mockRepo.On("GetByID", "cmpt_001").Return(expected, nil)

	result, err := svc.GetByID("cmpt_001")

	assert.NoError(t, err)
	assert.Equal(t, expected.Name, result.Name)
}

func TestGetCompetitorByID_EmptyID(t *testing.T) {
	svc := NewCompetitorService(nil)

	result, err := svc.GetByID("")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "竞品ID不能为空")
}
