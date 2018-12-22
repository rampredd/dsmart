package common

import (
	"fmt"
	"os"

	"github.com/micro/go-config"
	"github.com/micro/go-config/source/file"
)

//define data structures for APIs
// Login
type Login struct {
	Username string `json:"username" form:"username" query:"username"`
	Password string `json:"password" form:"password" query:"password"`
}

//contact
type Contact struct {
	First_name   string `json:"first_name" form:"first_name" query:"first_name"`
	Last_name    string `json:"last_name" form:"last_name" query:"last_name"`
	Organization string `json:"organization" form:"organization" query:"organization"`
	Phone_number string `json:"phone_number" form:"phone_number" query:"phone_number"`
	Email        string `json:"email" form:"email" query:"email"`
	Website      string `json:"website" form:"website" query:"website"`
}

type Token struct {
	Auth_token    string `json:"auth_token" form:"auth_token" query:"auth_token"`
	Refresh_token string `json:"refresh_token" form:"refresh_token" query:"refresh_token"`
}

// define structure to store runtime config
type LogCfg struct {
	Api   bool `json:"api"`
	Db    bool `json:"db"`
	Error bool `json:"error"`
	Other bool `json:"other"`
}

// define structure to store database config
type DbCfg struct {
	Address  string `json:"address"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	DbName   string `json:"dbName"`
	User     string `json:"user"`
	Token    string `json:"token"`
	Contact  string `json:"contact"`
}

var (
	err    error
	logCfg LogCfg
	dbCfg  DbCfg
	stdOut = os.Stdout
)

const (
	API                 = 1
	DB                  = 2
	ERROR               = 3
	OTHER               = 50
	EMPTY_TOKEN         = "BEARER TOKEN IS EMPTY"
	BIND_ERROR          = "DATA BIND ERROR"
	FILE_ERROR          = "CANNOT LOAD HTML FILE"
	INVALID_CREDENTIALS = "INVALID CREDENTIALS"
)

func Log(who int, args ...interface{}) {
	switch who {
	case API:
		if logCfg.Api {
			fmt.Fprintln(stdOut, args)
		}
	case DB:
		if logCfg.Db {
			fmt.Fprintln(stdOut, args)
		}
	case ERROR:
		if logCfg.Error {
			fmt.Fprintln(stdOut, args)
		}
	case OTHER:
		fmt.Fprintln(stdOut, args)
	}
}

/* print errors*/
func IsError(err error) bool {
	if err != nil {
		Log(ERROR, err.Error())
		return true
	}

	return false
}

func loadConfigFile() error {
	//	read config file
	// load the config from a file source
	path := "./config/config.json"
	if err := config.Load(file.NewSource(
		file.WithPath(path),
	)); err != nil {
		return err
	}
	return nil
}

func GetDbConfig() DbCfg {
	return dbCfg
}

func LoadConfig() {
	err = loadConfigFile()
	if IsError(err) {
		os.Exit(1)
	}
	err = config.Get("config", "db").Scan(&dbCfg)
	if IsError(err) {
		os.Exit(1)
	}

	err = config.Get("config", "log").Scan(&logCfg)
	if IsError(err) {
		os.Exit(1)
	}
}

//func PrintError(str string, e error) {
//	fmt.Println(str, e)
//}
