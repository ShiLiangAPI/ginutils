package utils

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type InitData struct {
	C           *gin.Context
	DB          *gorm.DB
	Model       any    // 模型，建议使用RespObj
	Table       string // 表名，建议使用RespObj
	PK          int64
	PKList      []int64
	ReqObj      any
	RespObj     any
	RespObjList any
	// 用于数据分页
	NoPage      bool
	CurrentPage int
	PageSize    int
	// 用于数据处理
	Order       string
	OrderList   []string
	Preload     string
	PreloadList []string
	Where       map[string]any
	// 用于查询树形列表
	NodeId int64
	// 用于关联查询
	RelationField string
	RelationObj   any
	RelationWhere map[string]any
	// 用于关联修改
	FilterField string
	UpdateField string
	// 内部使用
	query *gorm.DB
}
