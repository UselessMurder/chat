package routes

import (
	"../db/pgwrapper"
	"../hashgenerator"
	"../models"
	"../sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

func GetLoginHandler(c *gin.Context) {

	intr, _ := c.Get("currentSession")
	currentSession := intr.(*sessions.Session)

	if currentSession.Status != 0 {
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	inputError := c.DefaultQuery("error", "0")

	type ls struct {
		Il    int
		Stage uint8
	}

	errValue, err := strconv.Atoi(inputError)

	if err != nil {
		errValue = 0
	}

	lg := ls{errValue, currentSession.Status}

	c.HTML(http.StatusOK, "login", lg)
}

func GetRegisterHandler(c *gin.Context) {

	intr, _ := c.Get("currentSession")
	currentSession := intr.(*sessions.Session)

	if currentSession.Status != 0 {
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	inputError := c.DefaultQuery("error", "0")

	type rs struct {
		Ir    int
		Stage uint8
	}

	errValue, err := strconv.Atoi(inputError)

	if err != nil {
		errValue = 0
	}

	reg := rs{errValue, currentSession.Status}

	c.HTML(http.StatusOK, "register", reg)
}

func PostLoginHandler(c *gin.Context) {

	username := c.PostForm("username")

	password := c.PostForm("password")

	intr, _ := c.Get("currentSession")
	currentSession := intr.(*sessions.Session)

	if currentSession.Status != 0 {
		c.Redirect(http.StatusMovedPermanently, "/")
		return
	}

	intr, _ = c.Get("database")
	db := intr.(*pgwrapper.DataBaseWrapper)

	row, err := db.QueryRow("GetUserByName", username)

	if err != nil {
		c.Redirect(http.StatusMovedPermanently, "/login?error=1")
		return
	}

	currentUser := models.User{}

	err = row.Scan(&currentUser.UserId, &currentUser.Login, &currentUser.Password, &currentUser.RegisterDate)

	if err != nil {
		c.Redirect(http.StatusMovedPermanently, "/login?error=1")
		return
	}

	pashash, _ := hashgenerator.GetHashSum28(password, "password")

	if currentUser.Password != pashash {
		c.Redirect(http.StatusMovedPermanently, "/login?error=1")
		return
	}

	currentSession.Status = 1
	currentSession.UserId = currentUser.UserId
	currentSession.UserName = currentUser.Login

	c.Redirect(http.StatusMovedPermanently, "/")
}

func PostRegisterHandler(c *gin.Context) {

	username := c.PostForm("username")

	password := c.PostForm("password")

	confirmPassword := c.PostForm("confirm-password")

	intr, _ := c.Get("currentSession")
	currentSession := intr.(*sessions.Session)

	if currentSession.Status != 0 {
		c.Redirect(http.StatusMovedPermanently, "/")
		return
	}

	if username == "" || password == "" || confirmPassword == "" {
		c.Redirect(http.StatusMovedPermanently, "/register?error=1")
		return
	}

	if len(username) < 3 || len(username) > 50 {
		c.Redirect(http.StatusMovedPermanently, "/register?error=1")
		return
	}

	if len(password) < 5 || len(password) > 50 {
		c.Redirect(http.StatusMovedPermanently, "/register?error=1")
		return
	}

	if password != confirmPassword {
		c.Redirect(http.StatusMovedPermanently, "/register?error=1")
		return
	}

	intr, _ = c.Get("database")
	db := intr.(*pgwrapper.DataBaseWrapper)

	pashash, _ := hashgenerator.GetHashSum28(password, "password")

	err := db.ExecTransact("RegisterUser", username, pashash, time.Now())

	if err != nil {
		c.Redirect(http.StatusMovedPermanently, "/register?error=1")
		return
	}

	c.Redirect(http.StatusMovedPermanently, "/login")
}

func LogoutHandler(c *gin.Context) {

	intr, _ := c.Get("currentSession")
	currentSession := intr.(*sessions.Session)

	if currentSession.Status == 0 {
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	currentSession.RoomId = 0
	currentSession.Status = 0
	currentSession.UserId = 0
	currentSession.UserName = "Guest"

	c.Redirect(http.StatusTemporaryRedirect, "/")
	return
}
