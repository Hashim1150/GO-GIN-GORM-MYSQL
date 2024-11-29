package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type User struct {
	UID      uint   `gorm:"primary_key;autoIncrement"`
	Emailid  string `json:"emailid" gorm:"unique"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

var db *gorm.DB

func initdb() {
	var err error
	dsn := "root:==bitstek@700@tcp(127.0.0.1:3306)/exp?charset=utf8mb4&parseTime=True&loc=Local"
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to db ", err)
	}
	db.AutoMigrate(&User{})
	defer dbmiddleware()
}

// middlware to verify db connection
func dbmiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sqldb, err := db.DB()
		if err != nil || sqldb.Ping() != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "database connection failed",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

func main() {
	initdb()
	router := gin.Default()
	router.GET("/users", getallusers)
	router.POST("/users", createusers)
	router.GET("/users/:emailid", getusers)
	router.PUT("/users/:emailid", updateusers)
	router.DELETE("/users/:uid", deleteusers)
	router.Run(":747")
}
func getallusers(c *gin.Context) {
	var cred []User

	if err := db.Find(&cred).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cred)
}

func createusers(c *gin.Context) {
	var cred User

	if err := c.ShouldBindJSON(&cred); err != nil {
		c.JSON(500, gin.H{
			"error": err.Error()})
		return
	}
	if err := db.Create(&cred).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error()})
	}
	c.JSON(500, gin.H{"details": cred})
}

func getusers(c *gin.Context) {
	emailid := c.Param("emailid")
	var cred User

	if err := db.First(&cred, "emailid =?", emailid).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if err := db.First(&cred, "password =?", cred.Password).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cred.Password)
}

func updateusers(c *gin.Context) {
	emailid := c.Param("emailid")
	var cred User
	if err := c.ShouldBindJSON(&cred); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if emailid != cred.Emailid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name in url is not same as payload"})
		return
	}
	if err := db.Model(&cred).Where("emailid = ?", emailid).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"details updated": cred})
}
func deleteusers(c *gin.Context) {
	uid := c.Param("uid")
	if err := db.Delete(&User{}, "uid = ?", uid).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}
