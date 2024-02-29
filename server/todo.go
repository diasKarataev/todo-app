package main

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/joho/godotenv"
	"github.com/jordan-wright/email"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/time/rate"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"net/smtp"
	"os"
	"strconv"
	"time"
)

var (
	limiter        = rate.NewLimiter(300, 1) // Rate limit of 1 request
	db             *gorm.DB
	log            *logrus.Logger
	jwtSecret      = []byte(os.Getenv("JWT_SECRET"))
	tokenExpiresIn = time.Hour * 24
)

const (
	dsn = "host=localhost user=postgres password=Infinitive dbname=go-todo-app port=5432 sslmode=disable"
)

type Task struct {
	ID          uuid.UUID `gorm:"primaryKey"`
	Name        string    `json:"name"`
	Details     string    `json:"details"`
	CreatedDate time.Time `json:"createdDate"`
	HaveStar    bool      `json:"star" gorm:"default:false"`
	LastUpdated time.Time `json:"lastUpdated" gorm:"column:lastupdated"`
	UserId      uint      `json:"userId"`
}

type User struct {
	ID             uint   `gorm:"primaryKey"`
	Username       string `gorm:"uniqueIndex"`
	Email          string `gorm:"uniqueIndex"`
	Password       string
	IsActivated    bool   `json:"isActivated"`
	ActivationLink string `json:"activationLink"`
	ROLE           string `json:"-"`
}

type Claims struct {
	Username    string `json:"username"`
	IsActivated bool   `json:"isActivated"`
	Email       string `json:"email"`
	UserId      uint   `json:"userId"`
	ROLE        string `json:"role"`
	jwt.StandardClaims
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

func (t *Task) ToggleHaveStar() {
	t.HaveStar = !t.HaveStar
	t.LastUpdated = time.Now()
}

func checkLimiter(c *gin.Context) {
	if !limiter.Allow() {
		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
		return
	}
}

func main() {
	log = logrus.New()
	var err error
	log.SetFormatter(&logrus.JSONFormatter{})

	err = godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	file, err := os.OpenFile("logfile.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(file)
		log.Info("Log file opened successfully")
	} else {
		log.WithError(err).Fatal("Failed to open log file")
	}
	defer file.Close()

	log.WithFields(logrus.Fields{
		"action": "start",
		"status": "success",
	}).Info("Application started successfully")

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	db.AutoMigrate(&Task{})
	db.AutoMigrate(&User{})

	if err := CreateAdminUser(); err != nil {
		log.Fatal("Failed to create admin user:", err)
	}

	r := gin.Default()

	// CORS middleware
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{os.Getenv("CLIENT_URL")}
	config.AllowMethods = []string{"GET", "PATCH", "POST", "PUT", "DELETE", "OPTIONS"}                                                   // Разрешить все методы
	config.AllowHeaders = []string{"Origin", "Authorization", "Content-Type", "Access-Control-Allow-Headers", "Accept, Accept-Language"} // Разрешить определенные заголовки
	r.Use(cors.New(config))

	// Public routes
	r.POST("/register", Register)
	r.POST("/login", Login)
	r.GET("/activate/:activationLink", Activate)
	r.GET("/resend-activation-link", ResendActivationLink)
	// Auth middleware
	auth := r.Group("/api")
	auth.Use(AuthMiddleware())
	{

		auth.GET("/user-info", UserInfo)
		auth.GET("/tasks", GetTasks)
		auth.GET("/tasks/:id", GetTask)
		auth.POST("/tasks", CreateTask)
		auth.PUT("/tasks/:id", UpdateTask)
		auth.DELETE("/tasks/:id", DeleteTask)
		auth.PUT("/tasks/:id/toggle-star", ToggleStarTask)

	}

	auth.Use(AdminAuthMiddleware())
	{
		auth.GET("/admin/users", GetAllUsers)
		auth.POST("/admin/mailing", SendEmailToAllUsers)
	}

	// Start server
	log.Println("Сервер запущен на порту :8000")
	log.Fatal(http.ListenAndServe(":8000", r))
}

func GetAllUsers(c *gin.Context) {
	var users []User
	if err := db.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	var usersResponse []gin.H
	for _, user := range users {
		userResponse := gin.H{
			"ID":          user.ID,
			"Username":    user.Username,
			"Email":       user.Email,
			"IsActivated": user.IsActivated,
			"ROLE":        user.ROLE,
		}
		usersResponse = append(usersResponse, userResponse)
	}

	c.JSON(http.StatusOK, usersResponse)
}

func Register(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if user.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	var existingEmailUser User
	if err := db.Where("email = ?", user.Email).First(&existingEmailUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		return
	}

	// Проверка наличия username в базе данных
	var existingUsernameUser User
	if err := db.Where("username = ?", user.Username).First(&existingUsernameUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	user.ActivationLink = uuid.New().String()
	user.IsActivated = false
	user.ROLE = "USER"
	user.Password = string(hashedPassword)
	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	if err := SendActivationEmail(user.Email, user.ActivationLink); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send activation email"})
		return
	}

	c.Status(http.StatusCreated)
}

func SendActivationEmail(to, activationLink string) error {
	from := "karataev020902@gmail.com"
	pass := os.Getenv("SMTP_KEY")

	e := email.NewEmail()
	e.From = from
	e.To = []string{to}
	e.Subject = "Activate your account"
	e.HTML = []byte(fmt.Sprintf("Click <a href=\"%s/activate/%s\">here</a> to activate your account", os.Getenv("API_URL"), activationLink))

	return e.Send("smtp.gmail.com:587", smtp.PlainAuth("", from, pass, "smtp.gmail.com"))
}

func ResendActivationLink(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
		return
	}
	tokenString := authHeader[len("Bearer "):]

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	userId := claims.UserId
	if userId == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "UserId is required"})
		return
	}

	var user User
	if err := db.Where("id = ?", userId).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	newActivationLink := uuid.New().String()

	user.ActivationLink = newActivationLink
	if err := db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update ActivationLink"})
		return
	}

	if err := SendActivationEmail(user.Email, user.ActivationLink); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send activation email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Activation link resent successfully"})
}

func Login(c *gin.Context) {
	var loginRequest LoginRequest
	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var user User
	if err := db.Where("email = ?", loginRequest.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := GenerateToken(user.ID, user.Username, user.Email, user.IsActivated, user.ROLE)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, TokenResponse{Token: token})
}

func CreateAdminUser() error {
	var admin User
	result := db.First(&admin, "role = ?", "ADMIN")
	if result.Error == nil {
		// Пользователь admin уже существует
		return nil
	}
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// Произошла ошибка при поиске пользователя
		return result.Error
	}

	// Создание пользователя admin
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin = User{
		Username:       "admin",
		Email:          "admin",
		Password:       string(hashedPassword),
		IsActivated:    true,
		ActivationLink: "",
		ROLE:           "ADMIN",
	}

	if err := db.Create(&admin).Error; err != nil {
		return err
	}

	return nil
}
func Activate(c *gin.Context) {
	activationLink := c.Param("activationLink")

	var user User
	if err := db.Where("activation_link = ?", activationLink).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Activation link not found"})
		return
	}

	user.IsActivated = true
	if err := db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to activate user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User activated successfully"})
}

func GenerateToken(userId uint, username, email string, isActivated bool, role string) (string, error) {
	expirationTime := time.Now().Add(tokenExpiresIn)
	claims := &Claims{
		UserId:      userId,
		Username:    username,
		IsActivated: isActivated,
		Email:       email,
		ROLE:        role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
			return
		}

		tokenString := authHeader[len("Bearer "):]
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		c.Next()
	}
}

func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
			return
		}

		tokenString := authHeader[len("Bearer "):]
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// Проверка роли пользователя
		if claims.ROLE != "ADMIN" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Insufficient permissions"})
			return
		}

		c.Next()
	}
}

func GetTasks(c *gin.Context) {
	checkLimiter(c)
	if c.IsAborted() {
		return
	}
	var tasks []Task

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
		return
	}
	tokenString := authHeader[len("Bearer "):]

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	userId := claims.UserId

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	nameFilter := c.Query("name")
	detailsFilter := c.Query("details")
	starFilter, _ := strconv.ParseBool(c.Query("star"))
	sortField := c.Query("sortField")
	sortOrder := c.DefaultQuery("sortOrder", "asc")

	if sortField == "" {
		sortField = "ID"
	}

	offset := (page - 1) * pageSize

	query := db.Offset(offset).Limit(pageSize).Where("user_id = ?", userId)

	if nameFilter != "" {
		query = query.Where("name LIKE ?", "%"+nameFilter+"%")
	}

	if detailsFilter != "" {
		query = query.Where("details LIKE ?", "%"+detailsFilter+"%")
	}

	if starFilter {
		query = query.Where("have_star = ?", true)
	}

	orderClause := sortField + " " + sortOrder
	query = query.Order(orderClause)

	if err := query.Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения задач"})
		return
	}
	log.WithFields(logrus.Fields{
		"action":     "getTasks",
		"page":       page,
		"pageSize":   pageSize,
		"nameFilter": nameFilter,
	}).Info("GetTasks executed successfully")

	c.JSON(http.StatusOK, tasks)
}

func GetTask(c *gin.Context) {
	checkLimiter(c)
	if c.IsAborted() {
		return
	}
	var task Task
	id := c.Param("id")
	if err := db.First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Задача не найдена"})
		return
	}

	log.WithFields(logrus.Fields{
		"action": "getTasks",
	}).Info("GetTask executed successfully")

	c.JSON(http.StatusOK, task)
}

func SendEmailToAllUsers(c *gin.Context) {
	var request struct {
		Subject string `json:"subject"`
		Body    string `json:"body"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var users []User
	if err := db.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	for _, user := range users {
		if err := SendEmail(user.Email, request.Subject, request.Body); err != nil {
			continue
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email sent to all users successfully"})
}

func SendEmail(to, subject, body string) error {
	from := "karataev020902@gmail.com"
	pass := os.Getenv("SMTP_KEY")

	e := email.NewEmail()
	e.From = from
	e.To = []string{to}
	e.Subject = subject
	e.HTML = []byte(body)

	return e.Send("smtp.gmail.com:587", smtp.PlainAuth("", from, pass, "smtp.gmail.com"))
}

func UserInfo(c *gin.Context) {
	// Получаем токен из заголовка Authorization
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
		return
	}
	tokenString := authHeader[len("Bearer "):]

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	userInfo := gin.H{
		"userId":      claims.UserId,
		"username":    claims.Username,
		"email":       claims.Email,
		"isActivated": claims.IsActivated,
		"ROLE":        claims.ROLE,
	}
	c.JSON(http.StatusOK, userInfo)
}

func CreateTask(c *gin.Context) {
	checkLimiter(c)
	if c.IsAborted() {
		return
	}
	var newTask Task
	if err := c.BindJSON(&newTask); err != nil {
		log.WithFields(logrus.Fields{
			"action": "createTask",
			"error":  err.Error(),
		}).Error("Error binding JSON for creating task")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка входных данных"})
		return
	}

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
		return
	}
	tokenString := authHeader[len("Bearer "):]

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}
	newTask.ID = uuid.New()
	newTask.CreatedDate = time.Now()
	newTask.LastUpdated = newTask.CreatedDate
	newTask.HaveStar = false
	newTask.UserId = claims.UserId // Назначить userId в поле newTask.UserId

	if err := db.Create(&newTask).Error; err != nil {
		log.WithFields(logrus.Fields{
			"action": "createTask",
			"error":  err.Error(),
		}).Error("Error creating task in the database")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания задачи"})
		return
	}

	log.WithFields(logrus.Fields{
		"action": "createTask",
		"taskID": newTask.ID,
	}).Info("Task created successfully")

	c.JSON(http.StatusCreated, newTask)
}

func UpdateTask(c *gin.Context) {
	checkLimiter(c)
	if c.IsAborted() {
		return
	}
	var updatedTask Task
	id := c.Param("id")

	taskID, err := uuid.Parse(id)
	if err != nil {
		log.WithFields(logrus.Fields{
			"action": "updateTask",
			"error":  err.Error(),
		}).Error("Error parsing task ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный идентификатор задачи"})
		return
	}

	if err := db.First(&updatedTask, taskID).Error; err != nil {
		log.WithFields(logrus.Fields{
			"action": "updateTask",
			"error":  err.Error(),
		}).Error("Error retrieving task for update")
		c.JSON(http.StatusNotFound, gin.H{"error": "Задача не найдена"})
		return
	}

	if err := c.BindJSON(&updatedTask); err != nil {
		log.WithFields(logrus.Fields{
			"action": "updateTask",
			"error":  err.Error(),
		}).Error("Error binding JSON for updating task")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка входных данных"})
		return
	}

	updatedTask.LastUpdated = time.Now()

	if err := db.Save(&updatedTask).Error; err != nil {
		log.WithFields(logrus.Fields{
			"action": "updateTask",
			"error":  err.Error(),
		}).Error("Error updating task in the database")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления задачи"})
		return
	}

	log.WithFields(logrus.Fields{
		"action": "updateTask",
		"taskID": updatedTask.ID,
	}).Info("Task updated successfully")

	c.JSON(http.StatusOK, updatedTask)
}

func DeleteTask(c *gin.Context) {
	checkLimiter(c)
	if c.IsAborted() {
		return
	}
	var task Task
	id := c.Param("id")

	// Преобразовать id в uuid.UUID
	taskID, err := uuid.Parse(id)
	if err != nil {
		log.WithFields(logrus.Fields{
			"action": "deleteTask",
			"error":  err.Error(),
		}).Error("Error parsing task ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный идентификатор задачи"})
		return
	}

	if err := db.First(&task, taskID).Error; err != nil {
		log.WithFields(logrus.Fields{
			"action": "deleteTask",
			"error":  err.Error(),
		}).Error("Error retrieving task for delete")
		c.JSON(http.StatusNotFound, gin.H{"error": "Задача не найдена"})
		return
	}

	if err := db.Delete(&task).Error; err != nil {
		log.WithFields(logrus.Fields{
			"action": "deleteTask",
			"error":  err.Error(),
		}).Error("Error deleting task from the database")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления задачи"})
		return
	}

	log.WithFields(logrus.Fields{
		"action": "deleteTask",
		"taskID": task.ID,
	}).Info("Task deleted successfully")

	c.JSON(http.StatusOK, gin.H{"message": "Задача успешно удалена"})
}

func ToggleStarTask(c *gin.Context) {
	checkLimiter(c)
	if c.IsAborted() {
		return
	}
	var task Task
	var taskID uuid.UUID // Объявление переменной taskID типа uuid.UUID
	id := c.Param("id")
	taskID, err := uuid.Parse(id) // Преобразование строки id в тип uuid.UUID

	if err != nil {
		log.WithFields(logrus.Fields{
			"action": "toggleStarTask",
			"error":  err.Error(),
		}).Error("Error parsing task ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат идентификатора задачи"})
		return
	}

	if err := db.First(&task, "id = ?", taskID).Error; err != nil {
		log.WithFields(logrus.Fields{
			"action": "toggleStarTask",
			"error":  err.Error(),
		}).Error("Error retrieving task for toggle star")
		c.JSON(http.StatusNotFound, gin.H{"error": "Задача не найдена"})
		return
	}

	task.ToggleHaveStar()

	if err := db.Save(&task).Error; err != nil {
		log.WithFields(logrus.Fields{
			"action": "toggleStarTask",
			"error":  err.Error(),
		}).Error("Error updating task for toggle star")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления задачи"})
		return
	}

	log.WithFields(logrus.Fields{
		"action":      "toggleStarTask",
		"taskID":      task.ID,
		"haveStar":    task.HaveStar,
		"lastUpdated": task.LastUpdated,
	}).Info("Task star status toggled successfully")

	c.JSON(http.StatusOK, gin.H{"haveStar": task.HaveStar, "lastUpdated": task.LastUpdated})
}
