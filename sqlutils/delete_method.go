package utils

import (
	"gin_template/common/res"
	"gin_template/project/configs/database"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"strconv"
)

// DeleteOneObject T 删除数据的结构体类型
func DeleteOneObject[T any](c *gin.Context, pkStr string) bool {
	pk, _ := strconv.Atoi(c.Param(pkStr))
	if pk <= 0 {
		res.FailMsg(c, "参数错误")
		return false
	}

	return DeleteMethod[T](c, []int64{int64(pk)})
}

type IDListReq struct {
	IDList []int64 `json:"id_list"`
}

func DeleteAllObject[T any](c *gin.Context) bool {
	var idList IDListReq
	err := c.ShouldBindJSON(&idList)
	if err != nil {
		res.FailMsg(c, "请求参数错误")
		return false
	}

	return DeleteMethod[T](c, idList.IDList)
}

func DeleteTreeObject[T any](c *gin.Context) bool {
	var idList IDListReq
	err := c.ShouldBindJSON(&idList)
	if err != nil {
		res.FailMsg(c, "请求参数错误")
		return false
	}

	for {
		if len(idList.IDList) <= 0 {
			return true
		}
		if ok := DeleteMethod[T](c, idList.IDList); !ok {
			return false
		}
		parentIDList := idList.IDList
		idList.IDList = []int64{}
		var modelType T
		if err := database.DB.Model(&modelType).Where("parent_id IN (?)", parentIDList).Pluck("id", &idList.IDList).Error; err != nil {
			res.Error(c, err)
			return false
		}
	}
}

// DeleteMethod T 删除的数据库模型
func DeleteMethod[T any](c *gin.Context, idList []int64) bool {
	var objList []T
	findResult := database.DB.Where(idList).Find(&objList)
	if findResult.Error != nil {
		res.Error(c, findResult.Error)
		return false
	}
	if len(idList) == 1 && findResult.RowsAffected != 1 {
		res.FailMsg(c, "数据不存在")
		return false
	}

	result := database.DB.Delete(&objList)
	if result.Error != nil {
		if result.Error == gorm.ErrMissingWhereClause {
			res.FailMsg(c, "无可删数据")
			return false
		}
		res.FailMsg(c, result.Error)
		return false
	}

	return true
}
