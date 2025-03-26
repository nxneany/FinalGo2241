package main

import (
	"Final_Go/controller"
	"fmt"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func main() {
	// ตั้งค่า config file
	viper.SetConfigName("config") // ชื่อไฟล์คอนฟิก (config.yaml)
	viper.AddConfigPath(".")      // เส้นทางการค้นหาไฟล์คอนฟิก
	err := viper.ReadInConfig()   // อ่านค่าจากไฟล์คอนฟิก
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	// ดึงค่า DSN จากไฟล์คอนฟิก
	dsn := viper.GetString("mysql.dsn")
	fmt.Println("MySQL DSN:", dsn)

	// เชื่อมต่อกับฐานข้อมูล MySQL
	dialector := mysql.Open(dsn)
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		panic(fmt.Errorf("failed to connect to database: %w", err))
	}
	fmt.Println("Connection successful")

	// ตั้งค่า DB ให้กับ controller
	DB = db
	controller.SetDB(DB)

	// เริ่มเซิร์ฟเวอร์
	controller.StartServer()
}
