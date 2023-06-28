package utils

import (
	"fmt"
	"gin_template/common/res"
	"gin_template/project/configs/database"
	"gin_template/project/configs/settings"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"reflect"
	"strconv"
	"strings"
)

// GetAllObject
// mapValue={"disablePage": true, ...}   objList 查询的返回值类型切片
func GetAllObject(c *gin.Context, mapValue map[string]any, objList any) (map[string]any, bool) {
	// 处理 mapValue 数据库查询和筛选
	var querySet *gorm.DB
	var ok bool
	var err error
	querySet, ok = getQuerySet(mapValue)
	if !ok {
		querySet = database.DB
	}
	querySet = UseMapValue(c, querySet, mapValue)
	// 处理分页
	disablePage, ok := mapValue["disablePage"]
	var pageMap map[string]any
	if ok == false || disablePage == false {
		pageMap = Paginate(c, querySet)
	}
	// 查询结果
	err = querySet.Find(objList).Error
	if err != nil {
		res.FailMsg(c, err)
		return nil, false
	}
	pageMap["list"] = objList
	return pageMap, true
}

// GetRelationObject
// RT 关联类型  T 查询结果类型   objList 查询的返回值类型切片
// mapValue=mapValue = {"id/pk": "post_id", "filter_id": "user_id", "where": {}， "filter_where": {}}
func GetRelationObject[RT, T any](c *gin.Context, mapValue map[string]any, objList any) bool {
	filterIDStr, ok := mapValue["update_id"]
	if !ok {
		res.BusinessError(c)
		return false
	}

	var IDList []int64
	var relationObj RT
	relationQuerySet, ok := getQuerySet(mapValue)
	if !ok {
		relationQuerySet = database.DB.Model(&relationObj)
	}
	relationQuerySet = UseMapValue(c, relationQuerySet, mapValue)
	pkStr, ok := mapValue["pk"]
	if !ok {
		pkStr, ok = mapValue["id"]
		if !ok {
			pkStr = ""
		}
	}
	if pkStr != "" {
		pk, _ := strconv.Atoi(c.Param(pkStr.(string)))
		if pk <= 0 {
			res.FailMsg(c, "参数错误")
			return false
		}
		relationQuerySet = relationQuerySet.Where(fmt.Sprintf("%s = ?", pkStr.(string)), pk)
	}
	relationQuerySet.Pluck(filterIDStr.(string), &IDList)
	if len(IDList) > 0 {
		var filterObj T
		filterWhere, ok := mapValue["filter_where"]
		if ok {
			mapValue["where"] = filterWhere
		} else {
			mapValue["where"] = nil
		}
		filterQuerySet, ok := getQuerySet(mapValue)
		if !ok {
			filterQuerySet = database.DB.Model(&filterObj)
		}
		filterQuerySet = UseMapValue(c, filterQuerySet, mapValue)
		if err := filterQuerySet.Where(IDList).Scan(&objList).Error; err != nil {
			res.FailMsg(c, err)
			return false
		}
	}

	return true
}

// QueryWhere 构建where条件, 用于列表筛选查询
func QueryWhere(c *gin.Context, query *gorm.DB, where map[string]any) *gorm.DB {
	// condition => 数据库字段匹配条件;  args => 从url中获取值的key
	for condition, args := range where {
		queryValT := reflect.TypeOf(args).Kind()
		if queryValT == reflect.String {
			argsList := handlerArgs(c, args)
			if argsList == nil {
				continue
			}
			query = query.Where(handlerConditionString(condition), argsList...)
		} else if queryValT == reflect.Slice {
			if strings.Count(condition, "IN (?)") == 1 {
				query = query.Where(condition, args)
			} else {
				var whereValList []any
				for _, queryListVal := range args.([]any) {
					queryListValT := reflect.TypeOf(queryListVal).Kind()
					if queryListValT == reflect.String {
						argsList := handlerArgs(c, queryListVal)
						if argsList == nil {
							continue
						}
						whereValList = append(whereValList, argsList...)
					} else {
						whereValList = append(whereValList, queryListVal)
					}
				}
				query = query.Where(handlerConditionString(condition), whereValList...)
			}
		} else {
			query = query.Where(handlerConditionString(condition), args)
		}
	}
	return query
}

// BuildWhere 构建where条件, 用于值筛选查询
func BuildWhere(c *gin.Context, query *gorm.DB, where any) error {
	var err error
	t := reflect.TypeOf(where).Kind()
	if t == reflect.Struct || t == reflect.Map {
		query.Where(where)
	} else if t == reflect.Slice {
		for _, item := range where.([]any) {
			item := item.([]any)
			column := item[0]
			if reflect.TypeOf(column).Kind() == reflect.String {
				count := len(item)
				if count == 1 {
					return errors.New("切片长度不能小于2")
				}
				columnStr := column.(string)
				// 拼接参数形式
				if strings.Index(columnStr, "?") > -1 {
					query.Where(column, item[1:]...)
				} else {
					cond := "and" //cond
					opt := "="
					_opt := " = "
					var val any
					if count == 2 {
						opt = "="
						val = item[1]
					} else {
						opt = strings.ToLower(item[1].(string))
						_opt = " " + strings.ReplaceAll(opt, " ", "") + " "
						val = item[2]
					}

					if count == 4 {
						cond = strings.ToLower(strings.ReplaceAll(item[3].(string), " ", ""))
					}

					/*
					   '=', '<', '>', '<=', '>=', '<>', '!=', '<=>',
					   'like', 'like binary', 'not like', 'ilike',
					   '&', '|', '^', '<<', '>>',
					   'rlike', 'regexp', 'not regexp',
					   '~', '~*', '!~', '!~*', 'similar to',
					   'not similar to', 'not ilike', '~~*', '!~~*',
					*/

					if strings.Index(" in notin ", _opt) > -1 {
						// val 是数组类型
						column = columnStr + " " + opt + " (?)"
					} else if strings.Index(" = < > <= >= <> != <=> like likebinary notlike ilike rlike regexp notregexp", _opt) > -1 {
						column = columnStr + " " + opt + " ?"
					}

					if cond == "and" {
						query.Where(column, val)
					} else {
						query.Or(column, val)
					}
				}
			} else if t == reflect.Map /*Map*/ {
				query.Where(item)
			} else {
				/*
					// 解决and 与 or 混合查询，但这种写法有问题，会抛出 invalid query condition
					db = db.Where(func(db *gorm.DB) *gorm.DB {
						db, err = BuildWhere(db, item)
						if err != nil {
							panic(err)
						}
						return db
					})*/

				err = BuildWhere(c, query, item)
				if err != nil {
					return err
				}
			}
		}
	} else {
		return errors.New("参数有误")
	}
	return nil
}

// Paginate 默认分页
func Paginate(c *gin.Context, query *gorm.DB) map[string]any {
	var disablePage bool
	var pageNum int
	var pageSize int
	var err error

	disablePageStr, ok := c.GetQuery("disable_page")
	if ok {
		disablePage, _ = strconv.ParseBool(disablePageStr)
	}
	if disablePage == true {
		return map[string]any{}
	}

	pageNumStr, ok := c.GetQuery("page_num")
	if ok {
		pageNum, err = strconv.Atoi(pageNumStr)
		if err != nil {
			pageNum = 1
		} else if pageNum <= 0 {
			pageNum = 1
		}
	} else {
		pageNum = 1
	}

	pageSizeStr, ok := c.GetQuery("page_size")
	if ok {
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil {
			pageSize = settings.AppSettings.PageSize
		}
		switch {
		case pageSize <= 0:
			pageSize = 10
		case pageSize > settings.AppSettings.MaxPageSize:
			pageSize = settings.AppSettings.MaxPageSize
		}
	} else {
		pageSize = settings.AppSettings.PageSize
	}
	var total int64
	query.Count(&total)
	query.Offset((pageNum - 1) * pageSize).Limit(pageSize)
	return map[string]any{
		"page_num":  pageNum,
		"page_size": pageSize,
		"total":     total,
	}
}

// UseMapValue  mapValue={"table": string, "model": model.XXX{}, "order": string, "where": map[string]any}
func UseMapValue(c *gin.Context, querySet *gorm.DB, mapValue map[string]any) *gorm.DB {
	if querySet == nil || mapValue == nil {
		return querySet
	}
	order, ok := mapValue["order"]
	if ok {
		orderValT := reflect.TypeOf(order).Kind()
		if orderValT == reflect.String {
			querySet = querySet.Order(order)
		} else if orderValT == reflect.Slice {
			for _, orderVal := range order.([]string) {
				querySet = querySet.Order(orderVal)
			}
		}
	}
	where, ok := mapValue["where"]
	if ok {
		querySet = QueryWhere(c, querySet, where.(map[string]any))
	}
	preloadVal, ok := mapValue["preload"]
	if ok {
		preloadValT := reflect.TypeOf(preloadVal).Kind()
		if preloadValT == reflect.String {
			querySet = querySet.Preload(preloadVal.(string))
		} else if preloadValT == reflect.Slice {
			for _, preloadString := range preloadVal.([]string) {
				querySet = querySet.Preload(preloadString)
			}
		}
	}
	return querySet
}

///////////////////////////// 不对外使用 /////////////////////////////

// 获取带有表信息的 *gorm.DB
func getQuerySet(mapValue map[string]any) (*gorm.DB, bool) {
	querySet := database.DB
	table, ok := mapValue["table"]
	if ok {
		querySet = querySet.Table(table.(string))
	} else {
		dataModel, ok := mapValue["model"]
		if ok {
			querySet = querySet.Model(dataModel)
		} else {
			//res.Error(c, "列表查询未传入table或者model")
			return nil, false
		}
	}
	return querySet, true
}

// handlerConditionString 数据库字段匹配条件；无？自动匹配 = ？
func handlerConditionString(condition string) string {
	var stringBuild strings.Builder
	stringBuild.WriteString(condition)
	if strings.Index(condition, "?") == -1 {
		stringBuild.WriteString(" = ?")
	}
	return stringBuild.String()
}

// handlerConditionAndArgs
// 当QueryWhere条件是字符串时， 拼接 = ? 和字符串类型处理等
func handlerArgs(c *gin.Context, argsString any) []any {
	// 匹配条件（多个？）多个值 ; 分割，值类型 , 分割
	var whereValList []any
	if strings.Count(argsString.(string), ";") > 0 {
		argsList := strings.Split(argsString.(string), ";")
		for _, args := range argsList {
			whereVal, ok := parseQueryValue(c, args)
			if !ok {
				continue
			}
			whereValList = append(whereValList, whereVal)
		}
	} else {
		whereVal, ok := parseQueryValue(c, argsString.(string))
		if !ok {
			return nil
		}
		whereValList = append(whereValList, whereVal)
	}
	return whereValList
}

// parseQueryValue 把字符串 "id_delete;bool" 通过url获取值， 并转换为 bool或者nil
func parseQueryValue(c *gin.Context, queryVal string) (any, bool) {
	valueTrList := strings.Split(queryVal, ",")
	switch len(valueTrList) {
	case 1:
		//whereVal, ok := c.GetQuery(queryVal)
		//if !ok || whereVal == "" {
		//	return nil, false
		//}
		return queryVal, true
	case 2:
		queryValue := valueTrList[0]
		valueType := valueTrList[1]
		if valueType == "@" {
			queryValue, ok := c.GetQuery(queryValue)
			if !ok || queryValue == "" {
				return nil, false
			}
			return queryValue, true
		}
		whereVal, err := queryTranslateType(queryValue, valueType)
		if err != nil {
			return nil, false
		}
		return whereVal, true
	default:
		queryValue := valueTrList[0]
		valueType := valueTrList[1]
		isQuery := valueTrList[2]
		if isQuery == "@" {
			queryValue, ok := c.GetQuery(queryValue)
			if !ok || queryValue == "" {
				return nil, false
			}
		}
		whereVal, err := queryTranslateType(queryValue, valueType)
		if err != nil {
			return nil, false
		}
		return whereVal, true
	}
}
