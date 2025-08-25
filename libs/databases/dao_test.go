package databases

import (
	"encoding/json"
	"testing"
)

type TestTrans struct {
	Id int `json:"id" xorm:"id"`
}

func (m *TestTrans) TableName() string {
	return "test_trans"
}
func TestEntityNew_Create(t *testing.T) {
	//logger := hclog.Default()
	c := Config{
		ShowSql: true,
		Master:  "root:123456@tcp(127.0.0.1)/csv?charset=utf8mb4&parseTime=true&loc=Local",
	}
	// Marshal Config to JSON if Init expects []byte
	configBytes, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
		return
	}
	conn, _ := Init(configBytes)
	createTable := `create table if not exists test_trans(id int)engine=innodb`
	conn.Query(createTable)
	dao := NewBaseDao(conn)
	session := dao.NewSession()
	defer session.Close()
	session.Begin()
	t1 := TestTrans{Id: 1}
	t2 := TestTrans{Id: 2}
	t3 := TestTrans{Id: 3}
	_, err = session.InsertOne(&t1)
	if err != nil {
		t.Fatal(err)
		session.Rollback()
		return
	}
	session.Begin()
	session.InsertOne(&t2)
	if _, err := dao.InsertOne(&t3); err != nil {
		t.Fatal(err)
		return
	}
	session.Commit()
}
