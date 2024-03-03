package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func runTestServer() *httptest.Server {
	db = initDB()
	log = initLogger()

	r := setupRoutes(db, log)

	return httptest.NewServer(r)
}

func TestRegister(t *testing.T) {
	ts := runTestServer()
	defer ts.Close()

	t.Run("it should return 200", func(t *testing.T) {
		requestBody := []byte(`{
           "email": "testRegister@gmail.com",
           "username": "testRegisterUser",
           "password": "password"
       }`)

		resp, err := http.Post(ts.URL+"/register", "application/json", bytes.NewBuffer(requestBody))
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		if err := removeUserByUsername("testRegisterUser"); err != nil {
			t.Fatalf("Failed to remove user from the database: %v", err)
		}
	})
}

func TestLogin(t *testing.T) {
	ts := runTestServer()
	defer ts.Close()

	createUserRequestBody := []byte(`{
       "email": "test@gmail.com",
       "username": "testLoginUser",
       "password": "password"
   }`)
	resp, err := http.Post(ts.URL+"/register", "application/json", bytes.NewBuffer(createUserRequestBody))
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Failed to register user: expected status code %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	loginRequestBody := []byte(`{
       "email": "test@gmail.com",
       "password": "password"
   }`)

	resp, err = http.Post(ts.URL+"/login", "application/json", bytes.NewBuffer(loginRequestBody))
	if err != nil {
		t.Fatalf("Failed to send POST request to login: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var tokenResponse struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		t.Fatalf("Failed to decode login response body: %v", err)
	}

	if tokenResponse.Token == "" {
		t.Fatalf("Received empty token")
	}

	if err := removeUserByUsername("testLoginUser"); err != nil {
		t.Fatalf("Failed to remove user from the database: %v", err)
	}
}

func getAuthToken() string {
	ts := runTestServer()

	loginRequestBody := []byte(`{
        "email": "otabek.shadimatov@gmail.com",
        "password": "91926499"
    }`)
	resp, err := http.Post(ts.URL+"/login", "application/json", bytes.NewBuffer(loginRequestBody))
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var tokenResponse struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return ""
	}

	return tokenResponse.Token
}

func TestUserInfo(t *testing.T) {
	ts := runTestServer()
	defer ts.Close()

	token := getAuthToken()
	if token == "" {
		t.Fatalf("Token is empty")
	}

	req, err := http.NewRequest("GET", ts.URL+"/api/user-info", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestGetTasks(t *testing.T) {
	ts := runTestServer()
	defer ts.Close()
	token := getAuthToken()

	t.Run("it should return 200", func(t *testing.T) {
		req, err := http.NewRequest("GET", ts.URL+"/api/tasks", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var tasks []Task
		err = json.NewDecoder(resp.Body).Decode(&tasks)
		if err != nil {
			t.Fatalf("Failed to decode response JSON: %v", err)
		}

		if len(tasks) == 0 {
			t.Fatalf("Expected tasks, got empty response")
		}
	})
}

func TestGetTask(t *testing.T) {
	ts := runTestServer()
	defer ts.Close()
	token := getAuthToken()
	lastTaskID, err := getLastTaskID(ts, token)
	if err != nil {
		t.Fatalf("Error getting last task ID: %v", err)
	}

	t.Run("it should return 200 for correct task ID", func(t *testing.T) {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/tasks/%d", ts.URL, lastTaskID), nil)
		if err != nil {
			t.Fatalf("Failed to create GET request: %v", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("it should return 404 for incorrect task ID", func(t *testing.T) {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/tasks/%d", ts.URL, lastTaskID+1), nil)
		if err != nil {
			t.Fatalf("Failed to create GET request: %v", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, 404, resp.StatusCode)
	})
}

func TestCreateTask(t *testing.T) {
	ts := runTestServer()
	defer ts.Close()

	token := getAuthToken()

	newTask := Task{
		Name:        "Test Task",
		Details:     "This is a test task",
		CreatedDate: time.Now(),
	}

	taskJSON, err := json.Marshal(newTask)
	if err != nil {
		t.Fatalf("Failed to marshal task to JSON: %v", err)
	}

	req, err := http.NewRequest("POST", ts.URL+"/api/tasks", bytes.NewBuffer(taskJSON))
	if err != nil {
		t.Fatalf("Failed to create HTTP request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to send POST request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status code %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	var createdTask Task
	err = json.NewDecoder(resp.Body).Decode(&createdTask)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	foundTask, err := getTaskByNameFromDB(newTask.Name)
	if err != nil {
		t.Fatalf("Failed to fetch task from the database: %v", err)
	}

	if foundTask.Name != createdTask.Name || foundTask.Details != createdTask.Details {
		t.Fatalf("Retrieved task does not match created task")
	}

	err = removeTaskFromDB(foundTask.Name, foundTask.Details)
	if err != nil {
		t.Fatalf("Failed to remove task from the database: %v", err)
	}
}

func TestUpdateTask(t *testing.T) {
	ts := runTestServer()
	defer ts.Close()
	token := getAuthToken()

	fmt.Println("Generated token:", token)

	newTask := Task{
		Name:        "Test Task",
		Details:     "This is a test task",
		CreatedDate: time.Now(),
		UserId:      65,
	}

	taskJSON, err := json.Marshal(newTask)
	if err != nil {
		t.Fatalf("Failed to marshal task to JSON: %v", err)
	}

	req, err := http.NewRequest("POST", ts.URL+"/api/tasks", bytes.NewBuffer(taskJSON))
	if err != nil {
		t.Fatalf("Failed to create POST request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send POST request: %v", err)
	}
	defer resp.Body.Close()

	fmt.Println("POST Response Status Code:", resp.StatusCode)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status code %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	var createdTask Task
	err = json.NewDecoder(resp.Body).Decode(&createdTask)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	updatedTask := Task{
		Name:        "Updated Test Task",
		Details:     "This is an updated test task",
		CreatedDate: createdTask.CreatedDate,
		LastUpdated: time.Now(),
	}

	updatedTaskJSON, err := json.Marshal(updatedTask)
	if err != nil {
		t.Fatalf("Failed to marshal updated task to JSON: %v", err)
	}

	req, err = http.NewRequest("PUT", fmt.Sprintf("%s/api/tasks/%d", ts.URL, createdTask.ID), bytes.NewBuffer(updatedTaskJSON))
	if err != nil {
		t.Fatalf("Failed to create PUT request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send PUT request: %v", err)
	}
	defer resp.Body.Close()

	fmt.Println("PUT Response Status Code:", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	err = removeTaskFromDB(updatedTask.Name, updatedTask.Details)
	err = removeTaskFromDB(createdTask.Name, createdTask.Details)
	if err != nil {
		t.Fatalf("Failed to remove task from the database: %v", err)
	}
}

func TestDeleteTask(t *testing.T) {
	// Create a test server
	ts := runTestServer()
	defer ts.Close()
	token := getAuthToken()

	newTask := Task{
		Name:        "Test Task",
		Details:     "This is a test task",
		CreatedDate: time.Now(),
	}
	taskJSON, err := json.Marshal(newTask)
	if err != nil {
		t.Fatalf("Failed to marshal task to JSON: %v", err)
	}

	req, err := http.NewRequest("POST", ts.URL+"/api/tasks", bytes.NewBuffer(taskJSON))
	if err != nil {
		t.Fatalf("Failed to create POST request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send POST request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status code %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	var createdTask Task
	err = json.NewDecoder(resp.Body).Decode(&createdTask)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	req, err = http.NewRequest("DELETE", fmt.Sprintf("%s/api/tasks/%d", ts.URL, createdTask.ID), nil)
	if err != nil {
		t.Fatalf("Failed to create DELETE request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token) // Add token to the header
	client = &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send DELETE request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = removeTaskFromDB(newTask.Name, newTask.Details)
		if err != nil {
			t.Fatalf("Failed to remove task from the database: %v", err)
		}

		t.Fatalf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

}

func TestToggleStarTask(t *testing.T) {
	ts := runTestServer()
	defer ts.Close()
	token := getAuthToken()

	newTask := Task{
		Name:        "Test Task",
		Details:     "This is a test task",
		CreatedDate: time.Now(),
		// HaveStar and LastUpdated fields will be populated automatically by the server
	}
	taskJSON, err := json.Marshal(newTask)
	if err != nil {
		t.Fatalf("Failed to marshal task to JSON: %v", err)
	}

	req, err := http.NewRequest("POST", ts.URL+"/api/tasks", bytes.NewBuffer(taskJSON))
	if err != nil {
		t.Fatalf("Failed to create POST request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token) // Add token to the header
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send POST request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status code %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	var createdTask Task
	err = json.NewDecoder(resp.Body).Decode(&createdTask)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	req, err = http.NewRequest("PUT", fmt.Sprintf("%s/api/tasks/%d/toggle-star", ts.URL, createdTask.ID), nil)
	if err != nil {
		t.Fatalf("Failed to create PATCH request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token) // Add token to the header
	client = &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send PUT request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	updatedTask, err := getTaskByNameFromDB(createdTask.Name)
	if err != nil {
		t.Fatalf("Failed to fetch updated task from the database: %v", err)
	}

	if updatedTask.HaveStar == createdTask.HaveStar {
		t.Fatalf("Expected star status to be toggled, but it remains unchanged")
	}

	err = removeTaskFromDB(newTask.Name, newTask.Details)
	if err != nil {
		t.Fatalf("Failed to remove task from the database: %v", err)
	}
}
