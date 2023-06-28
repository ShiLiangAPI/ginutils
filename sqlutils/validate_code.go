package utils

import (
	"fmt"
	"gin_template/project/configs/gredis"
	"gin_template/project/configs/settings"
	"github.com/goccy/go-json"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// HasInEffectCode 检查是否存在正在生效的验证码，限制发送频率
func HasInEffectCode(mobilePhone string, codeType int) bool {
	lastCodeValue, err := getLastCode(mobilePhone, codeType)
	if err != nil {
		return false
	}
	d := time.Now().Sub(lastCodeValue.SendTime)
	if d < settings.MessageSettings.CheckCodeResendTime {
		return false
	}
	return true
}

// MakeCode 生成指定类型的验证码
func MakeCode(mobilePhone string, codeType int) (string, error) {
	code := createRandom()
	key := getValidateCodeCacheKey(mobilePhone, codeType)
	value := codeCacheStruct{Code: "123456", SendTime: time.Now()}
	err := gredis.Set(key, value, 5*60)
	if err != nil {
		return "", err
	}
	return code, nil
}

// ValidateCode 校验验证码
func ValidateCode(mobilePhone string, code string, codeType int) bool {
	if code == settings.AppSettings.DefaultSMSCode {
		return true
	}
	lastValue, err := getLastCode(mobilePhone, codeType)
	if err != nil {
		return false
	}
	if lastValue.Code == code {
		ok, _ := deleteCode(mobilePhone, codeType)
		if ok {
			return true
		}
	}
	return false
}

// ******** 下面函数不对外 ***********

type codeCacheStruct struct {
	Code     string    `json:"code"`
	SendTime time.Time `json:"send_time"`
}

// getValidateCodeCacheKey 获取验证码缓存KEY
func getValidateCodeCacheKey(mobilePhone string, codeType int) string {
	return fmt.Sprintf("validate_code_%s_%d", mobilePhone, codeType)
}

// createRandom 生成6位数的随机短信验证码
func createRandom() string {
	var stringBuild strings.Builder
	for i := 0; i < 6; i++ {
		num := rand.Intn(10)
		stringBuild.WriteString(strconv.Itoa(num))
	}
	return stringBuild.String()
}

// getLastCode 获取指定手机号的最新验证码
func getLastCode(mobilePhone string, codeType int) (*codeCacheStruct, error) {
	key := getValidateCodeCacheKey(mobilePhone, codeType)
	value, _ := gredis.Get(key)
	var result codeCacheStruct
	err := json.Unmarshal(value, &result)
	if err != nil {
		return &codeCacheStruct{}, err
	}

	return &result, nil
}

// deleteCode 删除指定类型的验证码
func deleteCode(mobilePhone string, codeType int) (bool, error) {
	key := getValidateCodeCacheKey(mobilePhone, codeType)

	return gredis.Delete(key)
}
