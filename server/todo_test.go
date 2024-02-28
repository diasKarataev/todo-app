package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TestDB is a global variable to hold the database connection
var TestDB *gorm.DB

// SetupTestDB initializes the test database connection
func SetupTestDB() {
	dsn := "your_test_db_connection_string"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to test database: " + err.Error())
	}
	TestDB = db
}

func TestMain(m *testing.M) {
	// Setup test database
	SetupTestDB()
	defer TestDB.Close()

	// Run tests
	m.Run()
}

func TestGetTasksStatusCode(t *testing.T) {
	// Create a request to the /tasks endpoint
	req, err := http.NewRequest("GET", "/tasks", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a response recorder to record the response
	rr := httptest.NewRecorder()

	// Mocking Gin context
	router := setupServer(TestDB)
	router.ServeHTTP(rr, req)

	// Check if the status code is 200 OK
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d but got %d", http.StatusOK, rr.Code)
	}
}
