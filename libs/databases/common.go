package databases

import (
	"github.com/go-xorm/xorm"
)

func DoGet(call func() (bool, error)) error {
	has, err := call()
	if nil != err {
		return err
	}
	if !has {
		return ErrGetEmpty
	}
	return nil
}
func DoUpdate(call func() (int64, error)) error {
	ret, err := call()
	if nil != err {
		return err
	}
	if 0 == ret {
		return ErrUpdatedEmpty
	}
	return nil
}

func DoInsert(call func() (int64, error)) error {
	ret, err := call()
	if nil != err {
		return err
	}
	if 0 == ret {
		return ErrInsertedEmpty
	}
	return nil
}

func DoDelete(call func() (int64, error)) error {
	ret, err := call()
	if nil != err {
		return err
	}
	if 0 == ret {
		return ErrDeletedEmpty
	}
	return nil
}

type SessionDoctor int

type SessionHandler func(session *xorm.Session) (interface{}, SessionDoctor, error)

// WarpSession 事务装饰器
// error 为真事务回滚
// SessionDoctor 决定回滚还是提交
// func WarpSession(wrapper SessionWrapper) (interface{}, int, error) {
func WarpSession(s *xorm.Session, h SessionHandler) (interface{}, int, error) {
	defer s.Close()
	if err := s.Begin(); err != nil {
		return nil, 0, err
	}
	value, doctor, err := h(s)
	if err != nil {
		return nil, 0, err
	}

	if doctor == SessionDoctorCommit {
		if err2 := s.Commit(); err2 != nil {
			return nil, 0, err2
		}
	} else {
		if err2 := s.Rollback(); err2 != nil {
			return nil, 0, err2
		}
	}

	return value, 0, nil
}

type SessionWrapper struct {
	dao     BaseDao
	session SessionDao
	err     error
}

type ExecuteFunc func(session SessionDao) error

func NewSessionWrapper(s BaseDao) SessionWrapper {
	sw := SessionWrapper{
		dao:     s,
		session: s.NewSession(),
		err:     nil,
	}
	sw.err = sw.session.Begin()
	return sw
}

func (sw SessionWrapper) Execute(call ExecuteFunc) SessionWrapper {
	// 有错误直接返回
	if sw.err != nil {
		return sw
	}
	if sw.err = call(sw.session); sw.err != nil {
		_ = sw.session.Rollback()
	}
	return sw
}

func (sw SessionWrapper) CommitAndClose() error {
	defer sw.session.Close()
	if sw.err == nil {
		sw.err = sw.Commit()
	}
	return sw.err
}

func (sw SessionWrapper) Commit() error {
	// 没有错误，提交commit
	if sw.err == nil {
		sw.err = sw.session.Commit()
	} else {
		// 有错误，回滚
		_ = sw.session.Rollback()
	}

	return sw.err
}

func (sw SessionWrapper) Close() error {
	// 有错误则回滚后更新
	if sw.err != nil {
		_ = sw.session.Rollback()
	}
	sw.session.Close()
	return sw.err
}

func (sw SessionWrapper) GetErr() error {
	return sw.err
}
