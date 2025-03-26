package controller

import (
	"Final_Go/model"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var DB *gorm.DB

// ฟังก์ชันตั้งค่าตัวแปร DB
func SetDB(db *gorm.DB) {
	DB = db
}

func UserController(router *gin.Engine) {
	// ตั้งค่ากลุ่มสำหรับ /user
	routers := router.Group("/user")
	{
		routers.GET("/get", getUsers)
		routers.POST("/register", registerUser)
	}

	// เส้นทางใหม่สำหรับ login
	router.POST("/customer/login", loginUser)
}

// ฟังก์ชันดึงข้อมูลผู้ใช้ทั้งหมด
func getUsers(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database not initialized"})
		return
	}

	var users []model.Customer
	if err := DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": users})
}

// ฟังก์ชัน Hash Password
func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// ฟังก์ชันสมัครสมาชิก
func registerUser(c *gin.Context) {
	var input model.Customer

	// ใช้ ShouldBindJSON แทนการอ่าน Body ด้วย io.ReadAll
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// แฮชรหัสผ่านก่อนบันทึก
	hashedPassword, err := hashPassword(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	input.Password = hashedPassword
	input.CreatedAt = time.Now()
	input.UpdatedAt = time.Now()

	// บันทึกข้อมูลลงฐานข้อมูล
	if err := DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ตอบกลับข้อมูลที่สมัครสำเร็จ
	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"data":    input,
	})
}

// ฟังก์ชันตรวจสอบรหัสผ่าน
func checkPasswordHash(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// ฟังก์ชัน Login
func loginUser(c *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// รับข้อมูลจาก Request Body
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// ค้นหาผู้ใช้จากฐานข้อมูลโดยใช้ email
	var user model.Customer
	if err := DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// ตรวจสอบรหัสผ่าน
	if !checkPasswordHash(input.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// ถ้าผ่านการตรวจสอบ ลบข้อมูล Password ออกก่อนส่งกลับ
	user.Password = "" // ลบรหัสผ่านก่อนส่งกลับ

	c.JSON(http.StatusOK, gin.H{
		"CustomerID":  user.CustomerID,
		"FirstName":   user.FirstName,
		"LastName":    user.LastName,
		"Email":       user.Email,
		"PhoneNumber": user.PhoneNumber,
		"Address":     user.Address,
		"CreatedAt":   user.CreatedAt.Format(time.RFC3339),
		"UpdatedAt":   user.UpdatedAt.Format(time.RFC3339),
	})

}
