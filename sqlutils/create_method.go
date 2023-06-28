package utils

import (
	"fmt"
	"gin_template/common/res"
	"gin_template/project/configs/database"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/copier"
	"regexp"
)

// CreateObject T 请求体的参数，传入类型 obj 创建的对象,传入指针
func CreateObject[T any](c *gin.Context, obj any) bool {
	var requestData T

	if err := c.BindJSON(&requestData); err != nil {
		res.FailMsg(c, "请求参数错误")
		return false
	}

	if err := copier.Copy(obj, &requestData); err != nil {
		res.Error(c, err)
		return false
	}

	err := database.DB.Create(obj).Error
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
