package databases

import "errors"

// Config 数据库连接配置
type Config struct {
	ShowSql        bool     `json:"show_sql" hcl:"show_sql"`
	MaxIdle        int      `json:"max_idle" hcl:"max_idle"`
	MaxConn        int      `json:"max_conn" hcl:"max_conn"`
	Master         string   `json:"master" hcl:"master"`
	Slaves         []string `json:"slaves" hcl:"slaves"`
	UseMasterSlave bool     `json:"use_master_slave" hcl:"use_master_slave"`
	DbType         string   `json:"db_type" hcl:"db_type"`
}

var ErrGetEmpty = errors.New("found 0 rows")
var ErrUpdatedEmpty = errors.New("update affected 0 rows")
var ErrDeletedEmpty = errors.New("update affected 0 rows")
var ErrInsertedEmpty = errors.New("update affected 0 rows")
var ErrMigrateTableIDEmpty = errors.New("migrate table id nil")
var ErrMigrateTableNameEmpty = errors.New("migrate table name nil")

const (
	SessionDoctorCommit   SessionDoctor = 0
	SessionDoctorRollback SessionDoctor = 1
)
