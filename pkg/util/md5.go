package util

import (
	"crypto/md5"
	"encoding/hex"
	"smartest-go/pkg/setting"
)

// EncodeMD5 md5 encryption
func EncodeMD5(value string) string {
	m := md5.New()
	m.Write([]byte(value))
	m.Write([]byte(setting.AppSetting.MD5Salt))

	return hex.EncodeToString(m.Sum(nil))
}

// MD5Equals verify str is or not md5Str
func MD5Equals(str, md5Str string) bool {
	return EncodeMD5(str) == md5Str
}
