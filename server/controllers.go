package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
)

func removeUserByUsername(username string) error {
	var user User
	if err := db.Where("username = ?", username).Delete(&user).Error; err != nil {
		return err
	}
	return nil
}

func removeUserByID(userID uint) error {
	var user User
	if err := db.Where("id = ?", userID).Delete(&user).Error; err != nil {
		return err
	}
	return nil
}

func getTaskByNameFromDB(name string) (Task, error) {
	var task Task
	if err := db.Where("name = ?", name).First(&task).Error; err != nil {
		return Task{}, err
	}
	return task, nil
}

func removeTaskFromDB(name, details string) error {
	var task Task
	if err := db.Where("name = ? AND details = ?", name, details).Delete(&task).Error; err != nil {
		return err
	}
	return nil
}

func getLastTaskID(ts *httptest.Server, token string) (uint, error) {
	// Create a new HTTP client with the provided token in the header
	client := &http.Client{}
	req, err := http.NewRequest("GET", ts.URL+"/api/tasks", nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error fetching tasks: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var tasks []Task
	if err := json.NewDecoder(resp.Body).Decode(&tasks); err != nil {
		return 0, fmt.Errorf("error decoding response body: %v", err)
	}

	if len(tasks) == 0 {
		return 0, errors.New("no tasks found")
	}

	// Return the ID of the last task
	return tasks[len(tasks)-1].ID, nil
}
