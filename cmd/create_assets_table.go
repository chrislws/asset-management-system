package main

import (
	"asset-management-system/pkg/model"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

func main() {
	log.Println("开始创建资产表和索引...")

	// 初始化数据库连接
	db, err := model.InitDB()
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer db.Close()

	// 创建 assets 表
	log.Println("创建 assets 表...")
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS assets (
            id INT AUTO_INCREMENT PRIMARY KEY,
            serial_number VARCHAR(100),
            name VARCHAR(255) NOT NULL,
            category VARCHAR(50) NOT NULL,
            brand VARCHAR(50) NOT NULL,
            application_date DATE,
            specification VARCHAR(255),
            asset_code VARCHAR(100),
            order_date DATE,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            department VARCHAR(100),
            location VARCHAR(100),
            supplier VARCHAR(100),
            recipient VARCHAR(100),
            recipient_department VARCHAR(100),
            remarks TEXT
        )
    `)
	if err != nil {
		log.Fatalf("创建 assets 表失败: %v", err)
	}
	log.Println("assets 表创建成功")

	// 创建索引
	log.Println("开始创建索引...")
	_, err = db.Exec("CREATE INDEX idx_assets_serial_number ON assets(serial_number)")
	if err != nil {
		log.Fatalf("创建 idx_assets_serial_number 索引失败: %v", err)
	}

	_, err = db.Exec("CREATE INDEX idx_assets_name ON assets(name)")
	if err != nil {
		log.Fatalf("创建 idx_assets_name 索引失败: %v", err)
	}

	_, err = db.Exec("CREATE INDEX idx_assets_created_at ON assets(created_at)")
	if err != nil {
		log.Fatalf("创建 idx_assets_created_at 索引失败: %v", err)
	}

	log.Println("成功创建资产表和索引！")
}