package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetTasksStatusCode(t *testing.T) {
	// Create a request to the /tasks endpoint
	req, err := http.NewRequest("GET", "http://localhost:8000/tasks", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a response recorder to record the response
	rr := httptest.NewRecorder()

	// Mocking Gin context
	router := gin.Default()
	router.GET("/tasks", GetTasks)

	// Serve the HTTP request and record the response
	router.ServeHTTP(rr, req)

	// Check if the status code is 200 OK
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d but got %d", http.StatusOK, rr.Code)
	}
}
