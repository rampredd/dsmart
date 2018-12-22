package main

import (
	cmn "dsmart/common"
	db "dsmart/maria"
	"strconv"
	"strings"

	"io/ioutil"

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
	path := "html/login.html"
	b, err := ioutil.ReadFile(path)
	if cmn.IsError(err) {
		return c.JSON(http.StatusInternalServerError, cmn.FILE_ERROR)
	}
	s := string(b)
	return c.HTML(http.StatusOK, s)
}

func Verify(c echo.Context) error {
	login := &cmn.Login{}
	err := c.Bind(login)
	if cmn.IsError(err) {
		return c.JSON(http.StatusInternalServerError, cmn.BIND_ERROR)
	}
	rsp, ret := db.VerifyUser(login.Username, login.Password)
	if ret == false {
		return c.JSON(http.StatusInternalServerError, cmn.INVALID_CREDENTIALS)
	}
	return c.JSON(http.StatusOK, rsp)
}

func isContactValid(contact *cmn.Contact) (ret string) {
	ret = ""
	//	check phone number should have only 0-9 ascii
	_, err := strconv.Atoi(contact.Phone_number)
	if cmn.IsError(err) {
		ret = "PHONE NUMBER IS NOT A NUMBER"
		return
	}
	if len(contact.Phone_number) != 10 {
		ret = "PHONE NUMBER DOESN'T HAVE 10 DIGITS"
		return
	}

	if false == strings.ContainsRune(contact.Email, 64) {
		ret = "@ MISING FROM EMAIL"
	}
	if false == strings.ContainsRune(contact.Email, 46) {
		ret = ". MISING FROM EMAIL"
	}
	return
}

func CreateContact(c echo.Context) error {
	tok := getBearerToken(c)

	if tok == "" {
		return c.JSON(http.StatusBadRequest, cmn.EMPTY_TOKEN)
	}
	contact := &cmn.Contact{}
	err := c.Bind(contact)
	if cmn.IsError(err) {
		return c.JSON(http.StatusInternalServerError, cmn.BIND_ERROR)
	}
	//validate the contact parameters
	if ret := isContactValid(contact); ret != "" {
		return c.JSON(http.StatusBadRequest, ret)
	}
	if ret := db.CreateContact(tok, contact); ret != "" {
		return c.JSON(http.StatusBadRequest, ret)
	}
	return c.JSON(http.StatusOK, contact)
}

func GetContact(c echo.Context) error {
	tok := getBearerToken(c)

	if tok == "" {
		return c.JSON(http.StatusBadRequest, cmn.EMPTY_TOKEN)
	}
	contact, ret := db.GetContact(tok)
	if ret != "" {
		return c.JSON(http.StatusBadRequest, ret)
	}
	return c.JSON(http.StatusOK, contact)
}

func EditContact(c echo.Context) error {
	tok := getBearerToken(c)

	if tok == "" {
		return c.JSON(http.StatusBadRequest, cmn.EMPTY_TOKEN)
	}
	contact := &cmn.Contact{}
	err := c.Bind(contact)
	if cmn.IsError(err) {
		return c.JSON(http.StatusInternalServerError, cmn.BIND_ERROR)
	}
	//validate the contact parameters
	if ret := isContactValid(contact); ret != "" {
		return c.JSON(http.StatusBadRequest, ret)
	}
	if ret := db.EditContact(tok, contact); ret != "" {
		return c.JSON(http.StatusBadRequest, ret)
	}
	return c.JSON(http.StatusOK, contact)
}

func DeleteContact(c echo.Context) error {
	tok := getBearerToken(c)
	if tok == "" {
		return c.JSON(http.StatusBadRequest, cmn.EMPTY_TOKEN)
	}
	if ret := db.DeleteContact(tok); ret != "" {
		return c.JSON(http.StatusBadRequest, ret)
	}
	return c.JSON(http.StatusOK, "")
}

func main() {
	//	load config
	cmn.LoadConfig()

	e := echo.New()

	// Named routes
	e.GET("/Login", Login)
	e.POST("/Verify", Verify)
	e.POST("/CreateContact", CreateContact)
	e.PATCH("/EditContact", EditContact)
	e.GET("/GetContact", GetContact)
	e.DELETE("/DeleteContact", DeleteContact)
	e.Logger.Fatal(e.Start(":80"))
}
