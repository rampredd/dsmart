package main

import (
	cmn "dsmart/common"
	db "dsmart/maria"
	"strconv"
	"strings"

	//	"encoding/json"
	//	"io/ioutil"

	//	"fmt"
	"net/http"

	"github.com/labstack/echo"
)

func getBearerToken(c echo.Context) (reqToken string) {
	reqToken = ""
	h := c.Request().Header
	reqToken = h.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer:")
	if len(splitToken) > 1 {
		reqToken = splitToken[1]
	}
	return reqToken
}

func Login(c echo.Context) error {
	login := &cmn.Login{}
	err := c.Bind(login)
	if cmn.IsError(err) {
		return c.String(http.StatusInternalServerError, "data bind error")
	}
	rsp, ret := db.VerifyUser(login.Username, login.Password)
	if ret == false {
		return c.String(http.StatusInternalServerError, "invalid username/password")
	}
	return c.JSON(http.StatusOK, rsp)
}

func isPhoneValid(phone string) bool {
	if len(phone) != 10 {
		return false
	}
	_, err := strconv.Atoi(phone)
	if cmn.IsError(err) {
		return false
	}
	return true
}

func CreateContact(c echo.Context) error {
	tok := getBearerToken(c)

	if tok == "" {
		return c.JSON(http.StatusBadRequest, "BEARER TOKEN IS EMPTY")
	}
	contact := &cmn.Contact{}
	err := c.Bind(contact)
	if cmn.IsError(err) {
		return c.String(http.StatusInternalServerError, "DATA BIND ERROR")
	}
	//	check phone number should have only 0-9 ascii
	if ret := isPhoneValid(contact.Phone_number); ret == false {
		return c.String(http.StatusBadRequest, "PHONE NUMBER IS NOT VALID")
	}
	if ret := db.CreateContact(tok, contact); ret != "" {
		return c.String(http.StatusBadRequest, ret)
	}
	return c.JSON(http.StatusOK, contact)
}

func EditContact(c echo.Context) error {
	tok := getBearerToken(c)

	if tok == "" {
		return c.JSON(http.StatusBadRequest, "BEARER TOKEN IS EMPTY")
	}
	contact := &cmn.Contact{}
	err := c.Bind(contact)
	if cmn.IsError(err) {
		return c.String(http.StatusInternalServerError, "DATA BIND ERROR")
	}
	if ret := db.EditContact(tok, contact); ret != "" {
		return c.String(http.StatusBadRequest, ret)
	}
	return c.JSON(http.StatusOK, contact)
}

func DeleteContact(c echo.Context) error {
	tok := getBearerToken(c)
	if tok == "" {
		return c.JSON(http.StatusBadRequest, "BEARER TOKEN IS EMPTY")
	}
	if ret := db.DeleteContact(tok); ret != "" {
		return c.String(http.StatusBadRequest, ret)
	}
	return c.JSON(http.StatusOK, "CONTACT DELETED")
}

func main() {
	//	load config
	cmn.LoadConfig()

	e := echo.New()

	// Named routes
	e.GET("/Login", Login)
	e.POST("/CreateContact", CreateContact)
	e.PATCH("/EditContact", EditContact)
	e.DELETE("/DeleteContact", DeleteContact)
	e.Logger.Fatal(e.Start(":80"))
}
