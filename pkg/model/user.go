package model

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

func InitDB() (*sql.DB, error) {
	log.Println("初始化数据库连接...")
	dataSource := "zabbix:admin123@tcp(10.0.0.96:3306)/asset_management"
	db, err := sql.Open("mysql", dataSource)
	if err != nil {
		log.Printf("打开数据库连接失败: %v", err)
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		log.Printf("数据库 Ping 失败: %v", err)
		return nil, err
	}

	log.Println("数据库连接成功")
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)

	return db, nil
}

// 验证用户（示例函数，需根据实际需求实现）
func ValidateUser(username, password string) bool {
	log.Printf("验证用户: username=%s, password=%s", username, password)
	return username == "admin" && password == "admin"
}