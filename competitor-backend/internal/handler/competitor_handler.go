package handler

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"

	"competitor-backend/internal/service"
)

type CompetitorHandler struct {
	svc *service.CompetitorService
}

func NewCompetitorHandler(svc *service.CompetitorService) *CompetitorHandler {
	return &CompetitorHandler{svc: svc}
}

func (h *CompetitorHandler) Create(ctx context.Context, c *app.RequestContext) {
	var input service.CreateCompetitorInput
	if err := c.BindAndValidate(&input); err != nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"code":    40001,
			"message": "参数错误: 竞品名称不能为空",
			"data":    nil,
		})
		return
	}

	result, err := h.svc.Create(&input)
	if err != nil {
		hlog.CtxErrorf(ctx, "create competitor failed: %v", err)
		c.JSON(http.StatusConflict, map[string]interface{}{
			"code":    50004,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"code":    0,
		"message": "success",
		"data":    result,
	})
}

func (h *CompetitorHandler) Get(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")

	result, err := h.svc.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, map[string]interface{}{
			"code":    40401,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"code":    0,
		"message": "success",
		"data":    result,
	})
}
