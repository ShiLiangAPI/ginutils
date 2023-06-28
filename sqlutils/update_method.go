package utils

import (
	"fmt"
	"gin_template/common/function"
	"gin_template/common/res"
	"gin_template/project/configs/database"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"reflect"
	"regexp"
	"strconv"
)

// UpdateObject T 请求体参数，传入类型 obj 创建的对象,传入指针
func UpdateObject[T any](c *gin.Context, obj any, pkStr string) (ok bool) {
	pk, _ := strconv.Atoi(c.Param(pkStr))
	if pk <= 0 {
		res.FailMsg(c, "参数错误")
		return false
	}
	var requestData T
	if err := c.ShouldBindJSON(&requestData); err != nil {
		res.FailMsg(c, "请求参数错误")
		return false
	}

	err := database.DB.First(obj, pk).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			res.FailMsg(c, "数据不存在")
			return false
		} else {
			res.FailMsg(c, err)
			return false
		}
	}

	if err := copier.Copy(obj, &requestData); err != nil {
		res.Error(c, err)
		return false
	}
	result := database.DB.Where("id=?", pk).Updates(obj)
	if result.Error != nil {
		mysqlError, ok := result.Error.(*mysql.MySQLError)
		if ok == true && mysqlError.Number == 1062 {
			compileRegex := regexp.MustCompile("Duplicate entry '(.*-)?(.*?)' for key .*")
			matchArrStr := compileRegex.FindStringSubmatch(result.Error.(*mysql.MySQLError).Message)
			res.FailMsg(c, fmt.Sprintf("%s 已存在", matchArrStr[len(matchArrStr)-1]))
			return false
		}
		res.Error(c, result.Error)
		return false
	}
	// 当未修改数据时，影响也是0
	//if result.RowsAffected != 1 {
	//	res.FailMsg(c, "数据不存在")
	//	return false
	//}
	reflect.ValueOf(obj).Elem().FieldByName("ID").SetInt(int64(pk))
	return true
}

// UpdateFilterObject T 请求体参数，传入类型 obj 创建的对象,传入指针
// mapValue = {}
func UpdateFilterObject[T any](c *gin.Context, mapValue map[string]any, obj any) bool {
	if mapValue == nil {
		res.Error(c, "修改未添加筛选项")
		return false
	}

	querySet := database.DB
	querySet = QueryWhere(c, querySet, mapValue)

	var requestData T
	if err := c.ShouldBindJSON(&requestData); err != nil {
		res.FailMsg(c, "请求参数错误")
		return false
	}

	if err := copier.Copy(obj, &requestData); err != nil {
		res.Error(c, err)
		return false
	}
	err := querySet.Updates(obj).Error
	if err != nil {
		mysqlError, ok := err.(*mysql.MySQLError)
		if ok == true && mysqlError.Number == 1062 {
			compileRegex := regexp.MustCompile("Duplicate entry '(.*-)?(.*?)' for key .*")
			matchArrStr := compileRegex.FindStringSubmatch(err.(*mysql.MySQLError).Message)
			res.FailMsg(c, fmt.Sprintf("%s 已存在", matchArrStr[len(matchArrStr)-1]))
			return false
		}
		res.Error(c, err)
		return false
	}
	return true
}

// UpdateRelationObject T 关联的类型
// mapValue = {"id/pk": "post_id", "update_id": "user_id", "filter_field": "PostID", "update_field": "UserID"}
func UpdateRelationObject[T any](c *gin.Context, mapValue map[string]any) bool {
	var requestData IDListReq
	if err := c.ShouldBindJSON(&requestData); err != nil {
		res.FailMsg(c, "请求参数错误")
		return false
	}
	pkStr, ok := mapValue["pk"]
	if !ok {
		pkStr, ok = mapValue["id"]
		if !ok {
			pkStr = "pk"
		}
	}
	pk, _ := strconv.Atoi(c.Param(pkStr.(string)))
	if pk <= 0 {
		res.FailMsg(c, "参数错误")
		return false
	}
	updateIDStr, ok := mapValue["update_id"]
	if !ok {
		res.FailMsg(c, "缺少参数update_id")
		return false
	}
	filterFieldStr, ok := mapValue["filter_field"]
	if !ok {
		res.FailMsg(c, "缺少参数filter_field")
		return false
	}
	updateFieldStr, ok := mapValue["update_field"]
	if !ok {
		res.FailMsg(c, "缺少参数update_field")
		return false
	}

	var modelType T
	var activeIDList []int64
	var addingIDList []int64
	var deletingIDList []int64
	var relationList []T
	err := database.DB.Model(&modelType).Where(fmt.Sprintf("%s = ?", pkStr.(string)), pk).Pluck(updateIDStr.(string), &activeIDList).Error
	if err != nil {
		res.FailMsg(c, err)
		return false
	}
	addingIDList = function.Different[int64](requestData.IDList, activeIDList)
	deletingIDList = function.Different[int64](activeIDList, requestData.IDList)
	for _, value := range addingIDList {
		var model T
		modelReflect := reflect.ValueOf(&model).Elem()
		modelReflect.FieldByName(filterFieldStr.(string)).SetInt(int64(pk))
		modelReflect.FieldByName(updateFieldStr.(string)).SetInt(value)
		relationList = append(relationList, model)
	}
	if len(relationList) > 0 {
		if err = database.DB.Model(&modelType).CreateInBatches(relationList, len(relationList)).Error; err != nil {
			res.FailMsg(c, err)
			return false
		}
	}
	if len(deletingIDList) > 0 {
		if err = database.DB.Unscoped().Where(deletingIDList).Delete(&modelType).Error; err != nil {
			res.FailMsg(c, err)
			return false
		}
	}
	return true
}
