package databases

import (
	"encoding/json"
	"testing"
)

func TestPostgresEntity_Create(t *testing.T) {
	conStr, err := GetConnStr(map[string]interface{}{
		"host":   "127.0.0.1",
		"port":   5432,
		"user":   "root",
		"passwd": "rtl8139",
		"dbname": "csv",
	}, "postgres")
	if err != nil {
		t.Error(err)
		panic(err)
	}
	c := Config{
		Master: conStr,
		DbType: "postgres",
	}
	configBytes, err := json.Marshal(c)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	conn, _ := Init(configBytes)

	dao := NewBaseDao(conn)
	result, err := dao.GetDBMetas()
	if err != nil {
		t.Error(err)
	}
	b, err := json.Marshal(result)
	if err != nil {
		t.Error(err)
	}
	t.Log(string(b))
	tb, err := dao.GetTableMetas("test_transc")
	if err != nil {
		t.Log(err)
	}
	b, err = json.Marshal(tb)
	if err != nil {
		t.Error(err)
	}
	t.Log(string(b))
}
