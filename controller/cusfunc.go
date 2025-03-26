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
		routers.PUT("/update-address", updateAddress)
		routers.PUT("/change-password", changePassword)
	}

	// เส้นทางใหม่สำหรับ login
	router.POST("/customer/login", loginUser)
	router.GET("/showPD", searchProducts)
	router.POST("/cart/add-item", addItemToCart)
	router.GET("cart/showall/:customer_id", getAllCarts)

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

// ฟังก์ชันอัปเดตที่อยู่ โดยใช้ CustomerID
func updateAddress(c *gin.Context) {
	var input struct {
		CustomerID uint   `json:"customer_id"` // ใช้ CustomerID แทน Email
		Address    string `json:"address"`
	}

	// รับข้อมูลจาก Request Body
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// ค้นหาผู้ใช้จากฐานข้อมูลโดยใช้ CustomerID
	var user model.Customer
	if err := DB.Where("customer_id = ?", input.CustomerID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// อัปเดตที่อยู่
	user.Address = input.Address
	user.UpdatedAt = time.Now()

	// บันทึกการอัปเดตลงในฐานข้อมูล
	if err := DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update address"})
		return
	}

	// ส่งข้อมูลที่อัปเดตกลับไป
	c.JSON(http.StatusOK, gin.H{
		"message": "Address updated successfully",
		"data": gin.H{
			"CustomerID":  user.CustomerID,
			"FirstName":   user.FirstName,
			"LastName":    user.LastName,
			"Email":       user.Email,
			"PhoneNumber": user.PhoneNumber,
			"Address":     user.Address,
			"CreatedAt":   user.CreatedAt,
			"UpdatedAt":   user.UpdatedAt,
		},
	})
}

// ฟังก์ชันเปลี่ยนรหัสผ่าน
func changePassword(c *gin.Context) {
	var input struct {
		CustomerID  uint   `json:"customer_id"`  // CustomerID ของผู้ใช้
		OldPassword string `json:"old_password"` // รหัสผ่านเก่า
		NewPassword string `json:"new_password"` // รหัสผ่านใหม่
	}

	// รับข้อมูลจาก Request Body
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// ค้นหาผู้ใช้จากฐานข้อมูลโดยใช้ CustomerID
	var user model.Customer
	if err := DB.Where("customer_id = ?", input.CustomerID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// ตรวจสอบรหัสผ่านเก่ากับฐานข้อมูล
	if !checkPasswordHash(input.OldPassword, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Old password is incorrect"})
		return
	}

	// แฮชรหัสผ่านใหม่ก่อนบันทึก
	hashedPassword, err := hashPassword(input.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash new password"})
		return
	}

	// อัปเดตรหัสผ่านใหม่ในฐานข้อมูล
	user.Password = hashedPassword
	user.UpdatedAt = time.Now()

	// บันทึกการเปลี่ยนแปลงลงในฐานข้อมูล
	if err := DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to change password"})
		return
	}

	// ส่งข้อมูลที่อัปเดตกลับไป
	c.JSON(http.StatusOK, gin.H{
		"message": "Password updated successfully",
		"data": gin.H{
			"CustomerID":  user.CustomerID,
			"FirstName":   user.FirstName,
			"LastName":    user.LastName,
			"Email":       user.Email,
			"PhoneNumber": user.PhoneNumber,
			"Address":     user.Address,
			"CreatedAt":   user.CreatedAt,
			"UpdatedAt":   user.UpdatedAt,
		},
	})
}
func searchProducts(c *gin.Context) {
	var input struct {
		Description string  `json:"description"` // รายละเอียดสินค้า
		MinPrice    float64 `json:"min_price"`   // ราคาต่ำสุด
		MaxPrice    float64 `json:"max_price"`   // ราคาสูงสุด
	}

	// รับข้อมูลจาก request
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var products []model.Product
	// ค้นหาข้อมูลจากฐานข้อมูล
	query := DB.Model(&model.Product{})
	if input.Description != "" {
		query = query.Where("description LIKE ?", "%"+input.Description+"%")
	}
	if input.MinPrice > 0 {
		query = query.Where("price >= ?", input.MinPrice)
	}
	if input.MaxPrice > 0 {
		query = query.Where("price <= ?", input.MaxPrice)
	}

	// ดึงข้อมูลสินค้าที่ตรงตามเงื่อนไข
	if err := query.Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ส่งผลลัพธ์กลับไป
	c.JSON(http.StatusOK, gin.H{
		"products": products,
	})
}
func addItemToCart(c *gin.Context) {
	var input struct {
		CustomerID int    `json:"customer_id"`
		CartName   string `json:"cart_name"`
		ProductID  int    `json:"product_id"`
		Quantity   int    `json:"quantity"`
	}

	// รับข้อมูลจาก request
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var cart model.Cart
	// ค้นหารถเข็นตามชื่อที่ลูกค้ากำหนด
	if err := DB.Where("customer_id = ? AND cart_name = ?", input.CustomerID, input.CartName).First(&cart).Error; err != nil {
		// ถ้าไม่พบรถเข็น ให้สร้างใหม่
		cart = model.Cart{
			CustomerID: input.CustomerID,
			CartName:   input.CartName,
		}
		if err := DB.Create(&cart).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot create cart"})
			return
		}
	}

	var cartItem model.CartItem
	// ค้นหาว่ามีสินค้านี้อยู่ในรถเข็นแล้วหรือไม่
	if err := DB.Where("cart_id = ? AND product_id = ?", cart.CartID, input.ProductID).First(&cartItem).Error; err == nil {
		// ถ้ามีอยู่แล้วให้เพิ่มจำนวนสินค้า
		cartItem.Quantity += input.Quantity
		if err := DB.Save(&cartItem).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot update cart item"})
			return
		}
	} else {
		// ถ้ายังไม่มี ให้สร้างรายการสินค้าใหม่ในรถเข็น
		cartItem = model.CartItem{
			CartID:    cart.CartID,
			ProductID: input.ProductID,
			Quantity:  input.Quantity,
		}
		if err := DB.Create(&cartItem).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot add item to cart"})
			return
		}
	}

	// ส่ง response กลับ
	c.JSON(http.StatusOK, gin.H{
		"message":    "Item added to cart successfully",
		"cart_id":    cart.CartID,
		"product_id": input.ProductID,
		"quantity":   cartItem.Quantity,
	})
}
func getAllCarts(c *gin.Context) {
	customerID := c.Param("customer_id") // รับค่า customer_id จาก URL

	var carts []model.Cart
	// ค้นหารถเข็นทั้งหมดของลูกค้า
	if err := DB.Where("customer_id = ?", customerID).Find(&carts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot retrieve carts"})
		return
	}

	var result []gin.H
	for _, cart := range carts {
		var cartItems []struct {
			ProductID   int    `json:"product_id"`
			ProductName string `json:"product_name"`
			Quantity    int    `json:"quantity"`
			Price       string `json:"price"`
		}

		// ดึงข้อมูลสินค้าในรถเข็น
		query := `
            SELECT p.product_id, p.product_name, ci.quantity, p.price
            FROM cart_item ci
            JOIN product p ON ci.product_id = p.product_id
            WHERE ci.cart_id = ?
        `
		if err := DB.Raw(query, cart.CartID).Scan(&cartItems).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot retrieve cart items"})
			return
		}

		// เพิ่มข้อมูลลงใน JSON response
		result = append(result, gin.H{
			"cart_id":   cart.CartID,
			"cart_name": cart.CartName,
			"items":     cartItems,
		})
	}

	c.JSON(http.StatusOK, gin.H{"carts": result})
}
