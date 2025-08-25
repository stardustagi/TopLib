package databases

import (
	"errors"

	"xorm.io/xorm"
	"xorm.io/xorm/migrate"
)

type Dao interface {
	Count(bean interface{}) (int64, error)
	Exists(bean interface{}) (bool, error)

	InsertOne(entry interface{}) (int64, error)
	InsertMany(entries ...interface{}) (int64, error)

	Update(bean interface{}, where ...interface{}) (int64, error)
	UpdateById(id interface{}, bean interface{}) (int64, error)

	// 新增的 Upsert 方法
	Upsert(bean interface{}) (int64, error)
	UpsertById(id interface{}, bean interface{}) (int64, error)
	UpsertMany(beans ...interface{}) (int64, error)

	Delete(bean interface{}) (int64, error)
	DeleteById(id interface{}, bean interface{}) (int64, error)
	GetDBMetas() (map[string]interface{}, error)
	GetTableMetas(tableName string) ([]map[string]interface{}, error)
	FindById(id interface{}, bean interface{}) (bool, error)
	FindOne(bean interface{}) (bool, error)
	FindMany(rowsSlicePtr interface{}, orderBy string, condiBean ...interface{}) error
	FindAndCount(rowsSlicePtr interface{},
		pageable Pageable, condiBean ...interface{}) (int64, error)

	Query(rowsSlicePtr interface{}, sql string, Args ...interface{}) error
	Native() DBInterface
	Migrations(opt *migrate.Options, tables []map[string]interface{}) error
}

type SessionDao interface {
	Dao
	Session() *xorm.Session
	Begin() error
	Commit() error
	Rollback() error
	Close()
}

type BaseDao interface {
	Dao
	NewSession() SessionDao
}

type OrmBaseDao struct {
	conn    DBInterface
	session *xorm.Session
}

func NewBaseDao(conn DBInterface) BaseDao {
	return &OrmBaseDao{
		conn:    conn,
		session: conn.NewSession(),
	}
}

func (m *OrmBaseDao) Session() *xorm.Session {
	return m.session
}

// NewSession 创建一个session
func (m *OrmBaseDao) NewSession() SessionDao {
	sm := &OrmBaseDao{conn: m.conn}
	sm.session = m.conn.NewSession()
	return sm
}

// Begin 开启事务
func (m *OrmBaseDao) Begin() error {
	return m.session.Begin()
}

// Close 关闭事务
func (m *OrmBaseDao) Close() {
	m.session.Close()
}

func (m *OrmBaseDao) Commit() error {
	return m.session.Commit()
}

func (m *OrmBaseDao) Rollback() error {
	return m.session.Rollback()
}

func (m *OrmBaseDao) InsertOne(entry interface{}) (int64, error) {
	if m.session != nil {
		return m.session.InsertOne(entry)
	}
	return m.conn.InsertOne(entry)
}
func (m *OrmBaseDao) InsertMany(entries ...interface{}) (int64, error) {
	if m.session != nil {
		return m.session.Insert(entries...)
	}
	return m.conn.Insert(entries...)
}

func (m *OrmBaseDao) Update(bean interface{}, where ...interface{}) (int64, error) {
	if m.session != nil {
		return m.session.Update(bean, where...)
	}
	return m.conn.Update(bean, where...)
}
func (m *OrmBaseDao) UpdateById(id interface{}, bean interface{}) (int64, error) {
	if m.session != nil {
		return m.session.ID(id).Update(bean)
	}
	return m.conn.ID(id).Update(bean)
}

// Upsert 如果数据存在则更新，不存在则插入
func (m *OrmBaseDao) Upsert(bean interface{}) (int64, error) {
	// 检查数据是否存在
	exists, err := m.Exists(bean)
	if err != nil {
		return 0, err
	}

	if exists {
		// 数据存在，执行更新
		return m.Update(bean)
	} else {
		// 数据不存在，执行插入
		return m.InsertOne(bean)
	}
}

// UpsertById 根据ID进行Upsert操作
func (m *OrmBaseDao) UpsertById(id interface{}, bean interface{}) (int64, error) {
	// 检查指定ID的数据是否存在
	found, err := m.FindById(id, bean)
	if err != nil {
		return 0, err
	}

	if found {
		// 数据存在，执行更新
		return m.UpdateById(id, bean)
	} else {
		// 数据不存在，执行插入
		return m.InsertOne(bean)
	}
}

// UpsertMany 批量Upsert操作
func (m *OrmBaseDao) UpsertMany(beans ...interface{}) (int64, error) {
	var totalAffected int64 = 0

	for _, bean := range beans {
		affected, err := m.Upsert(bean)
		if err != nil {
			return totalAffected, err
		}
		totalAffected += affected
	}

	return totalAffected, nil
}

func (m *OrmBaseDao) Delete(bean interface{}) (int64, error) {
	if m.session != nil {
		return m.session.Delete(bean)
	}
	return m.conn.Delete(bean)
}
func (m *OrmBaseDao) DeleteById(id interface{}, bean interface{}) (int64, error) {
	if m.session != nil {
		return m.session.ID(id).Delete(bean)
	}
	return m.conn.ID(id).Delete(bean)
}

func (m *OrmBaseDao) Query(rowsSlicePtr interface{}, sql string, Args ...interface{}) error {

	if m.session != nil {
		return m.session.SQL(sql, Args...).Find(rowsSlicePtr)
	}
	return m.conn.SQL(sql, Args...).Find(rowsSlicePtr)
}

func (m *OrmBaseDao) FindById(id interface{}, bean interface{}) (bool, error) {
	if m.session != nil {
		return m.session.ID(id).Get(bean)
	}
	return m.conn.ID(id).Get(bean)
}

func (m *OrmBaseDao) FindOne(bean interface{}) (bool, error) {
	if m.session != nil {
		return m.session.Get(bean)
	}
	return m.conn.Get(bean)
}

func (m *OrmBaseDao) Count(bean interface{}) (int64, error) {
	if m.session != nil {
		return m.session.Count(bean)
	}
	return m.conn.Count(bean)
}

func (m *OrmBaseDao) Exists(bean interface{}) (bool, error) {
	if m.session != nil {
		return m.session.Exist(bean)
	}
	return m.conn.Exist(bean)
}
func (m *OrmBaseDao) FindMany(rowsSlicePtr interface{}, sort string, condiBean ...interface{}) error {
	if m.session != nil {
		return m.session.OrderBy(sort).Find(rowsSlicePtr, condiBean...)
	}
	return m.conn.OrderBy(sort).Find(rowsSlicePtr, condiBean...)
}

func (m *OrmBaseDao) FindAndCount(rowsSlicePtr interface{},
	pageable Pageable, condiBean ...interface{}) (int64, error) {

	if m.session != nil {
		return m.session.
			Limit(pageable.Limit(), pageable.Skip()).
			OrderBy(pageable.Sort()).
			FindAndCount(rowsSlicePtr, condiBean...)
	}
	return m.conn.
		Limit(pageable.Limit(), pageable.Skip()).
		OrderBy(pageable.Sort()).
		FindAndCount(rowsSlicePtr, condiBean...)
}

func (m *OrmBaseDao) Native() DBInterface {
	return m.conn
}

func (m *OrmBaseDao) Migrations(opt *migrate.Options, tables []map[string]interface{}) error {
	var migrations []*migrate.Migration
	for _, table := range tables {
		id, ok := table["id"]
		if !ok {
			return ErrMigrateTableIDEmpty
		}
		name, ok := table["name"]
		if !ok {
			return ErrMigrateTableNameEmpty
		}
		migration := &migrate.Migration{ID: id.(string),
			Migrate: func(tx *xorm.Engine) error {
				return tx.Sync2(name)
			},
			Rollback: func(tx *xorm.Engine) error {
				return tx.DropTables(name)
			}}
		migrations = append(migrations, migration)
	}

	x := migrate.New(m.conn.(*xorm.Engine), opt, migrations)
	x.InitSchema(func(tx *xorm.Engine) error {
		for _, table := range tables {
			name, ok := table["name"]
			if !ok {
				return ErrMigrateTableNameEmpty
			}
			err := tx.Sync2(name)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return nil
}

func (m *OrmBaseDao) GetDBMetas() (map[string]interface{}, error) {
	tables, err := m.conn.(*xorm.Engine).DBMetas()
	if err != nil {
		return nil, err
	}
	resultTable := make(map[string]interface{})
	for _, table := range tables {
		var resultTablesColumns []map[string]interface{}
		for _, col := range table.Columns() {
			colMap := make(map[string]interface{})
			colMap["name"] = col.Name
			colMap["sql_type"] = col.SQLType.Name
			colMap["length"] = col.Length
			colMap["nullable"] = col.Nullable
			resultTablesColumns = append(resultTablesColumns, colMap)
		}
		resultTable[table.Name] = resultTablesColumns
	}
	return resultTable, nil
}

func (m *OrmBaseDao) GetTableMetas(tableName string) ([]map[string]interface{}, error) {
	dbMetas, err := m.GetDBMetas()
	if err != nil {
		return nil, err
	}
	tableMetas, ok := dbMetas[tableName]
	if !ok {
		return nil, errors.New("table not found")
	}
	return tableMetas.([]map[string]interface{}), nil
}
