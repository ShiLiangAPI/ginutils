package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"gin_template/project/configs/settings"
	"golang.org/x/crypto/pbkdf2"
	"strings"
)

/////////////////////////// pbkdf2_sha256加密 ///////////////////////////

// PasswordHash pbkdf2_sha256加密,与python同步
func PasswordHash(password string) (string, error) {
	// 根据项目替换，值越大加密时间越长
	iterations := 1000
	salt := settings.AppSettings.SecretKey

	// pbkdf2加密 <--- 关键
	hash := pbkdf2.Key([]byte(password), []byte(salt), iterations, sha256.Size, sha256.New)

	// base64编码成为固定长度的字符串
	b64SaltHash := strings.TrimRight(base64.StdEncoding.EncodeToString([]byte(salt)), "=")
	b64PasswordHash := strings.TrimRight(base64.StdEncoding.EncodeToString(hash), "=")

	// 最终字符串拼接成pbkdf2_sha256密钥格式
	pwdHash := fmt.Sprintf("$%s$%d$%s$%s", "pbkdf2-sha256", iterations, b64SaltHash, b64PasswordHash)

	return pwdHash, nil
}

// PasswordVerify 密码校验
func PasswordVerify(pwd, hash string) bool {
	pwdHash, _ := PasswordHash(pwd)

	return pwdHash == hash
}

/////////////////////////// bcrypt加密 ///////////////////////////

// PasswordHash bcrypt加密
//func PasswordHash(pwd string) (string, error) {
//	bytes, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
//	if err != nil {
//		return "", err
//	}
//
//	return string(bytes), err
//}

// PasswordVerify bcrypt解密
//func PasswordVerify(pwd, hash string) bool {
//	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pwd))
//
//	return err == nil
//}
