package models

import "github.com/jinzhu/gorm"

const beginStatus = 1
const endStatus = 0

type Session struct {
	db           *gorm.DB // gorm db
	tx           *gorm.DB // 原生事务
	commitSign   int8     // 提交标记，控制是否提交事务
	rollbackSign bool     // 回滚标记，控制是否回滚事务
}

// GetSession 获取一个Session
func NewSession() *Session {
	session := new(Session)
	session.db = db
	return session
}

func GetSessionTx(session *Session) *gorm.DB {
	if session.tx != nil {
		return session.tx
	}

	return session.db
}

// Begin 开启事务
func (s *Session) Begin() {
	s.rollbackSign = true
	if s.tx == nil {
		s.tx = db.Begin()
		s.commitSign = beginStatus
	}
}

// Rollback 回滚事务
func (s *Session) Rollback() {
	if s.tx != nil && s.rollbackSign == true {
		s.tx.Rollback()
		s.tx = nil
	}
}

// Commit 提交事务
func (s *Session) Commit() {
	s.rollbackSign = false
	if s.tx != nil {
		if s.commitSign == beginStatus {
			s.tx.Commit()
			s.tx = nil
		} else {
			s.commitSign = endStatus
		}
	}
}
