package service

import (
	"errors"
	"fmt"
	"time"

	"competitor-backend/internal/model"
)

type CompetitorRepository interface {
	Create(c *model.Competitor) error
	GetByID(id string) (*model.Competitor, error)
	GetByName(name string) (*model.Competitor, error)
}

type CompetitorService struct {
	repo CompetitorRepository
}

func NewCompetitorService(repo CompetitorRepository) *CompetitorService {
	return &CompetitorService{repo: repo}
}

type CreateCompetitorInput struct {
	Name     string `json:"name"`
	Website  string `json:"website"`
	Category string `json:"category"`
}

func (s *CompetitorService) Create(input *CreateCompetitorInput) (*model.Competitor, error) {
	if input.Name == "" {
		return nil, errors.New("竞品名称不能为空")
	}

	existing, _ := s.repo.GetByName(input.Name)
	if existing != nil {
		return nil, fmt.Errorf("竞品已存在: %s", input.Name)
	}

	id := fmt.Sprintf("cmpt_%s_%03d", time.Now().Format("20060102"), 1)

	c := &model.Competitor{
		ID:       id,
		Name:     input.Name,
		Website:  input.Website,
		Category: input.Category,
		Status:   "active",
	}

	if err := s.repo.Create(c); err != nil {
		return nil, fmt.Errorf("创建竞品失败: %w", err)
	}

	return c, nil
}

func (s *CompetitorService) GetByID(id string) (*model.Competitor, error) {
	if id == "" {
		return nil, errors.New("竞品ID不能为空")
	}

	c, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("竞品不存在: %w", err)
	}

	return c, nil
}
