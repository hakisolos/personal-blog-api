package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type Blog struct {
	ID         uint      `json:"id"`
	Title      string    `json:"title"`
	Author     string    `json:"author"`
	Content    string    `json:"content"`
	Created_at time.Time `json:"created_at"`
}

var err error
var DB *sql.DB

func initDb() {
	var cstring string = "postgres://user_27b23fb2:eefc75a584f103014c75acc1b66b8d4a@db.pxxl.pro:56150/db_ab107e8b?sslmode=disable"
	DB, err = sql.Open("postgres", cstring)
	if err != nil {
		panic(err)
	}
	err := DB.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("db connected")
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS blogs (
			id SERIAL PRIMARY KEY,
			title TEXT NOT NULL,
			author TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)
	`)
	if err != nil {
		panic(err)
	}
}

func newblog(c *gin.Context) {
	var blog Blog
	err := c.BindJSON(&blog)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{
			"message": "an error occured",
			"Error":   err,
		})
	}
	q := `
		INSERT INTO blogs
		(title, author, content)
		VALUES ($1, $2, $3)
	`
	err = DB.QueryRow(q, &blog.Title, &blog.Author, &blog.Content).Scan(&blog.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]error{
			"error": err,
		})
		return
	}
	c.JSON(http.StatusCreated, map[string]any{
		"message": "blog created successfully",
		"data":    &blog,
	})
}

func getblogs(c *gin.Context) {
	rows, err := DB.Query("SELECT id, title, author, content, created_at FROM blogs ORDER BY id DESC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var blogs []Blog
	for rows.Next() {
		var b Blog
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Content, &b.Created_at); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		blogs = append(blogs, b)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    blogs,
	})
}

func getblog(c *gin.Context) {
	blogtitle := c.Query("title")
	if blogtitle == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "query parameter `title` required",
		})
	}
	var b Blog
	err := DB.QueryRow("SELECT id, title, author, content, created_at FROM blogs WHERE title=$1", blogtitle).
		Scan(&b.ID, &b.Title, &b.Author, &b.Content, &b.Created_at)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
	}
	c.JSON(http.StatusFound, gin.H{
		"success": true,
		"data":    b,
	})
}

func updateBlog(c *gin.Context) {
	id := c.Query("id") // get blog ID from ?id=
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing id"})
		return
	}

	var b Blog
	if err := c.BindJSON(&b); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	query := `UPDATE blogs SET title=$1, author=$2, content=$3 WHERE id=$4 RETURNING id, title, author, content, created_at`
	err := DB.QueryRow(query, b.Title, b.Author, b.Content, id).
		Scan(&b.ID, &b.Title, &b.Author, &b.Content, &b.Created_at)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    b,
	})
}

func deleteBlog(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing id"})
		return
	}

	result, err := DB.Exec("DELETE FROM blogs WHERE id=$1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No blog found with that id"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Blog deleted successfully",
	})
}

func main() {
	initDb()
	app := gin.Default()

	app.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]string{
			"message": "blog api running",
		})
	})
	app.POST("/newBlog", newblog)
	app.GET("/getBlogs", getblogs)
	app.GET("/getblog", getblog)
	app.PUT("/updateBlog", updateBlog)
	app.DELETE("/deleteBlog", deleteBlog)
	app.Run(":3001")
}
