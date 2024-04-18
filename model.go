package gowebkit

import (
	"fmt"
	"time"

	"github.com/ccb1900/gocommon/logger"
	"github.com/gin-gonic/gin"
)

type Result struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func ResultOk() Result {
	return Result{Code: 0, Msg: "ok"}
}

func ResultOkData(data interface{}) Result {
	return Result{Code: 0, Msg: "ok", Data: data}
}

func ResultPageData(total int64, data interface{}) Result {
	return Result{Code: 0, Msg: "ok", Data: map[string]interface{}{
		"total": total,
		"list":  data,
	}}
}

func ResultError(msg string) Result {
	return Result{Code: -1, Msg: msg}
}

type PageInfo struct {
	Page     int `json:"page" query:"page" uri:"page" form:"page"`
	Pagesize int `json:"pagesize" query:"pagesize" uri:"pagesize" form:"pagesize"`
}

func (p *PageInfo) Default() {
	if p.Page == 0 {
		p.Page = 1
	}

	if p.Pagesize == 0 {
		p.Pagesize = 10
	}
}

func Bind(ctx *gin.Context, req interface{}) error {
	if err := ctx.ShouldBind(req); err != nil {
		logger.Default().Warn("bind", "err", err)
	}
	if err := ctx.ShouldBindJSON(req); err != nil {
		logger.Default().Warn("bind json", "err", err)
	}
	if err := ctx.ShouldBindQuery(req); err != nil {
		logger.Default().Warn("bind query", "err", err)
	}
	if err := ctx.ShouldBindUri(req); err != nil {
		logger.Default().Warn("bind uri", "err", err)
		return err
	}
	return nil
}

type DefaultDatetimeWithDeletedAt struct {
	DefaultDatetime
	DeletedAt *time.Time `gorm:"column:deleted_at" json:"deleted_at"`
}
type DefaultDatetime struct {
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

var DEFAULT_TIMEFORMAT = fmt.Sprintf("%s %s", time.DateOnly, time.TimeOnly)

type DefaultID struct {
	ID DefaultIDType `gorm:"primaryKey" json:"id"`
}
type DefaultID64 struct {
	ID DefaultID64Type `gorm:"primaryKey" json:"id"`
}

type (
	DefaultIDType   uint
	DefaultID64Type uint64
)
