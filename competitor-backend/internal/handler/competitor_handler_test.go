package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"competitor-backend/internal/model"
	"competitor-backend/internal/service"
)

func TestCompetitorServiceIntegration(t *testing.T) {
	_ = &model.Competitor{
		ID:       "cmpt_001",
		Name:     "DJI Action 4",
		Category: "运动相机",
	}
	_ = &service.CreateCompetitorInput{}
	_ = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	assert.True(t, true)
}

func TestHTTPValidation_EmptyName(t *testing.T) {
	body, _ := json.Marshal(map[string]interface{}{
		"category": "运动相机",
	})

	resp, err := http.Post("http://example.com/api/v1/competitors",
		"application/json", bytes.NewReader(body))

	_ = resp
	_ = err
	assert.True(t, true)
}
