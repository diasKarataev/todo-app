package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	dsn = "host=localhost user=postgres password=Infinitive dbname=go-todo-app port=5432 sslmode=disable"
)

var db *gorm.DB

type Task struct {
	ID      uint   `gorm:"primaryKey"`
	Name    string `json:"name"`
	Details string `json:"details"`
}

func main() {
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	db.AutoMigrate(&Task{})

	r := gin.Default()

	r.GET("/tasks", GetTasks)
	r.GET("/tasks/:id", GetTask)
	r.POST("/tasks", CreateTask)
	r.PUT("/tasks/:id", UpdateTask)
	r.DELETE("/tasks/:id", DeleteTask)

	log.Println("Сервер запущен на порту :8000")
	log.Fatal(http.ListenAndServe(":8000", r))
}

func GetTasks(c *gin.Context) {
	var tasks []Task
	if err := db.Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения задач"})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

func GetTask(c *gin.Context) {
	var task Task
	id := c.Param("id")
	if err := db.First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Задача не найдена"})
		return
	}

	c.JSON(http.StatusOK, task)
}

func CreateTask(c *gin.Context) {
	var newTask Task
	if err := c.BindJSON(&newTask); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка входных данных"})
		return
	}

	if err := db.Create(&newTask).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания задачи"})
		return
	}

	c.JSON(http.StatusCreated, newTask)
}

func UpdateTask(c *gin.Context) {
	var updatedTask Task
	id := c.Param("id")

	if err := db.First(&updatedTask, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Задача не найдена"})
		return
	}

	if err := c.BindJSON(&updatedTask); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка входных данных"})
		return
	}

	if err := db.Save(&updatedTask).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления задачи"})
		return
	}

	c.JSON(http.StatusOK, updatedTask)
}

func DeleteTask(c *gin.Context) {
	var task Task
	id := c.Param("id")

	if err := db.First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Задача не найдена"})
		return
	}

	if err := db.Delete(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления задачи"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Задача успешно удалена"})
}
