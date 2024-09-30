package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

type User struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var (
	users  = []User{}
	nextID = uint(1)
	mu     sync.Mutex
)

const dataFile = "data.json"

func main() {
	var err error
	users, err = LoadUsers()
	if err != nil {
		fmt.Println("Error loading users:", err)
	}

	router := gin.Default()
	router.POST("/users", createUser)
	router.GET("/users", getUsers)
	router.GET("/users/:id", getUser)
	router.PUT("/users/:id", updateUser)
	router.DELETE("/users/:id", deleteUser)
	router.Run(":8080")
}

func LoadUsers() ([]User, error) {
	var users []User
	file, err := ioutil.ReadFile(dataFile)
	if err != nil {
		return users, err
	}
	json.Unmarshal(file, &users)
	return users, nil
}

func SaveUsers(users []User) error {
	data, err := json.Marshal(users)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dataFile, data, 0644)
}

func createUser(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mu.Lock()
	user.ID = nextID
	nextID++
	users = append(users, user)
	mu.Unlock()

	if err := SaveUsers(users); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save user"})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func getUsers(c *gin.Context) {
	c.JSON(http.StatusOK, users)
}

func getUser(c *gin.Context) {
	id := c.Param("id")
	for _, user := range users {
		if fmt.Sprint(user.ID) == id {
			c.JSON(http.StatusOK, user)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
}

func updateUser(c *gin.Context) {
	id := c.Param("id")
	var updatedUser User

	if err := c.ShouldBindJSON(&updatedUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mu.Lock()
	for i, user := range users {
		if fmt.Sprint(user.ID) == id {
			users[i].Name = updatedUser.Name
			users[i].Email = updatedUser.Email
			mu.Unlock()
			if err := SaveUsers(users); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save user"})
				return
			}
			c.JSON(http.StatusOK, users[i])
			return
		}
	}
	mu.Unlock()
	c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
}

func deleteUser(c *gin.Context) {
	id := c.Param("id")

	mu.Lock()
	defer mu.Unlock()
	for i, user := range users {
		if fmt.Sprint(user.ID) == id {
			users = append(users[:i], users[i+1:]...)
			if err := SaveUsers(users); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save user"})
				return
			}
			c.JSON(http.StatusNoContent, nil)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
}
