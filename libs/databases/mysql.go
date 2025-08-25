package databases

import (
	"bytes"
	"errors"
	"log"
	"text/template"

	"github.com/stardustagi/TopLib/utils"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"xorm.io/core"
	"xorm.io/xorm"
)

// DBInterface 数据库接口
type DBInterface interface {
	xorm.EngineInterface
}

var dbConn DBInterface

func Init(c []byte) (DBInterface, error) {
	var err error
	var conf *Config
	conf, err = utils.Bytes2Struct[*Config](c)
	if err != nil {
		panic("Failed to parse database configuration: " + err.Error())
	}
	if conf.UseMasterSlave {
		// return NewMSConn(conf)
		dbConn, err = NewMSConn(conf)
	}
	// return NewSingleConn(conf)
	dbConn, err = NewSingleConn(conf)
	return dbConn, err
}

// NewSingleConn 初始化数据库连接
// mysql fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=%s",user,pwd,host,db,charset)
// postgres fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s sslcert=%s sslkey=%s sslrootcert=%s",
//
//	host, port, user, name, pass, sslMode, SslCert, SslKey, SslRootCert)
func NewSingleConn(c *Config) (DBInterface, error) {
	if nil == c || "" == c.Master {
		return nil, errors.New("config or config.Url can not be null")
	}
	var conn *xorm.Engine
	var err error
	switch c.DbType {
	case "mysql":
		conn, err = xorm.NewEngine("mysql", c.Master)
	case "postgres":
		conn, err = xorm.NewEngine("postgres", c.Master)
	default:
		conn, err = xorm.NewEngine("mysql", c.Master)
	}
	if nil != err || nil == conn {
		log.Println("failed to initializing db connection:", err)
		return nil, err
	}
	conn.SetMapper(core.GonicMapper{})
	conn.ShowSQL(c.ShowSql)
	conn.SetMaxIdleConns(c.MaxIdle)
	conn.SetMaxOpenConns(c.MaxConn)
	return conn, nil
}

// NewMSConn 初始化主从数据库连接, master不能为空，slaves可以为空
func NewMSConn(c *Config) (DBInterface, error) {
	if nil == c || "" == c.Master {
		return nil, errors.New("config or config.Url can not be null")
	}
	conns := make([]string, len(c.Slaves)+1)
	conns[0] = c.Master
	for i, v := range c.Slaves {
		conns[i+1] = v
		if "" == v {
			return nil, errors.New("config or config.Url can not be null")
		}
	}

	var group *xorm.EngineGroup
	var err error
	switch c.DbType {
	case "mysql":
		group, err = xorm.NewEngineGroup("mysql", c.Master)
	case "postgres":
		group, err = xorm.NewEngineGroup("postgres", c.Master)
	default:
		group, err = xorm.NewEngineGroup("mysql", c.Master)
	}

	if nil != err || nil == group {
		log.Printf("failed to initializing db connection: %s\n", err)
		return nil, err
	}

	group.SetMapper(core.GonicMapper{})
	group.ShowSQL(c.ShowSql)
	group.SetMaxIdleConns(c.MaxIdle)
	group.SetMaxOpenConns(c.MaxConn)
	return group, nil
}

func GetMySqlDB() DBInterface {
	return dbConn
}

func GetDao() BaseDao {
	return NewBaseDao(dbConn)
}

// GetConnStr 检查传入字典和类型，返回数据库连接字符串，
// @params c map[string]interface{}{ "host":"127.0.0.1","port":3306,"user":"root","passwd":"123456","dbname":"csv"}
// @param t string "mysql" or "postgres"
// @return string, error
func GetConnStr(c map[string]interface{}, t string) (string, error) {
	var constr bytes.Buffer
	var tpr *template.Template
	var err error
	switch t {
	case "mysql":
		{
			tmpl := `{{.user}}:{{.pwd}}@tcp({{.host}}:{{.port}})/{{.db}}?charset=utf8mb4&parseTime=true&loc=Local`
			tpr, err = template.New("config").Parse(tmpl)
		}
	case "postgres":
		{
			tmpl := `host={{.host}} port={{.port}} user={{.user}} dbname={{.db}} password={{.pwd}} sslmode=disable`
			tpr, err = template.New("config").Parse(tmpl)
		}
	default:
		return "", errors.New("unsupported db type")
	}
	// 错误检查
	if err != nil {
		return "", err
	}
	if err = tpr.Execute(&constr, c); err != nil {
		return "", err
	}
	return constr.String(), nil
}
