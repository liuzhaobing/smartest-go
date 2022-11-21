package util

import (
	"smartest-go/pkg/setting"
)

func GetPaginationParams(startPage int, pageSize int) (offset int, limit int) {
	limit = defaultLimitPagination(pageSize, setting.AppSetting.PageSize)
	offset = defaultOffsetPagination(startPage, limit, 0)
	return
}

func defaultOffsetPagination(startPage int, limit int, defaultValue int) int {
	offset := defaultValue

	if startPage > 0 {
		offset = (startPage - 1) * limit
	}

	return offset
}

func defaultLimitPagination(pageSize int, defaultValue int) int {
	if pageSize <= 0 {
		return defaultValue
	}

	limit := pageSize

	sizeLimit := setting.AppSetting.PageSizeLimit
	if limit > sizeLimit {
		return sizeLimit
	}

	return limit
}
