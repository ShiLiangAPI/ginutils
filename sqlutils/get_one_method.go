package utils

import (
	"gin_template/common/res"
	"gin_template/project/configs/database"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"strconv"
)

// GetObject
// T  返回结构体类型
// modelType 查询数据库类型 model.xxx{}
// pkStr  url上对应id字符串
func GetObject(c *gin.Context, mapValue map[string]any, obj any) bool {
	var err error
	var querySet *gorm.DB
	var ok bool
	pk := getPk(c, mapValue)
	if pk <= 0 {
		res.FailMsg(c, "参数错误")
		return false
	}

	querySet, ok = getQuerySet(mapValue)
	if !ok {
		querySet = database.DB
	}
	querySet = UseMapValue(c, querySet, mapValue)

	err = querySet.First(obj, pk).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			res.FailMsg(c, "数据不存在")
			return false
		} else {
			res.FailMsg(c, err)
			return false
		}
	}
	return true
}

// GetFirstObject T 返回结构体
// mapValue={"table": string, "model": model.XXX{}, "order": string, "where": map[string]any}
func GetFirstObject(c *gin.Context, mapValue map[string]any, obj any) bool {
	var querySet *gorm.DB
	var ok bool

	querySet, ok = getQuerySet(mapValue)
	if !ok {
		querySet = database.DB
	}
	querySet = UseMapValue(c, querySet, mapValue)

	err := querySet.First(&obj).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			res.FailMsg(c, "数据不存在")
			return false
		} else {
			res.FailMsg(c, err)
			return false
		}
	}
	return true
}

//////////////////////////// 不对外 ////////////////////////////

// 根据参数 获取id值，并转换为int64
func getPk(c *gin.Context, mapValue map[string]any) int64 {
	pkStr, ok := mapValue["pk"]
	if !ok {
		pkStr, ok = mapValue["id"]
		if !ok {
			pkStr = "pk"
		}
	}
	pk, _ := strconv.Atoi(c.Param(pkStr.(string)))
	return int64(pk)
}
