package controller

import (
	"Final_Go/model"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var DB *gorm.DB

// ฟังก์ชันนี้จะใช้ในการตั้งค่าตัวแปร DB
func SetDB(db *gorm.DB) {
	DB = db
}

func UserController(router *gin.Engine) {
	routers := router.Group("/get")
	{
		routers.GET("/user", getUsers)

	}
}

func getUsers(c *gin.Context) {

	if DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database not initialized"})
		return
	}

	var users []model.Customer
	if err := DB.Find(&users).Error; err != nil { // ดึงข้อมูลทั้งหมด
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": users}) // ส่งข้อมูลในรูปแบบ JSON
}
