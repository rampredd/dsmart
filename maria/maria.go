package maria

import (
	"database/sql"
	cmn "dsmart/common"
	"fmt"

	//	"reflect"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	_ "github.com/go-sql-driver/mysql"
)

var (
	db    *sql.DB
	err   error
	dbCfg cmn.DbCfg
)

const (
	SECRET      = "canubreakthis?"
	AUTHTAIL    = "AB12CD#$"
	REFRESHTAIL = "EF56GH&*"
)

func Connect() {
	dbCfg = cmn.GetDbConfig()
	str := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", dbCfg.Username, dbCfg.Password, dbCfg.Address, dbCfg.Port, dbCfg.DbName)
	db, err = sql.Open("mysql", str)
}

func CreateUsers() {
	Connect()

	if cmn.IsError(err) {
		return
	}
	defer db.Close()

	i := 0
	query := fmt.Sprintf("insert into user (username,password) values ")
	for ; i < 9; i++ {
		query += fmt.Sprintf("('user%d','pass%d'),", i, i)
	}
	query += fmt.Sprintf("('user%d','pass%d');", i, i)

	insert, err := db.Query(query)
	if cmn.IsError(err) {
		fmt.Println("MYSQL QUERY ERROR")
		return
	}
	insert.Close()
}

func CreateContact(tok string, contact *cmn.Contact) (ret string) {
	Connect()

	if cmn.IsError(err) {
		return
	}
	defer db.Close()
	ret = ""
	id := IsTokenValid(tok)
	if id == 0 {
		ret = "INVALID TOKEN"
		return
	}
	addContact := fmt.Sprintf("insert into %s (id, first_name, last_name, organization, phone_number, email, website) values(%d,'%s','%s','%s','%s','%s','%s')", dbCfg.Contact, id, contact.First_name, contact.Last_name, contact.Organization, contact.Phone_number, contact.Email, contact.Website)
	add, err := db.Query(addContact)
	if cmn.IsError(err) {
		ret = "DB INTERNAL ERROR CONTACT CREATE FAIL"
		return
	}
	defer add.Close()
	return
}

func getColVal(valType string, i interface{}) (res string, yes bool) {
	res = ""
	yes = false

	switch valType {
	case "string":
		res = i.(string)
		yes = true
	case "int32":
		i32 := i.(int32)
		if i32 > 0 {
			res = fmt.Sprintf("%s", i32)
			yes = true
		}
	}
	return res, yes
}

func EditContact(tok string, contact *cmn.Contact) (ret string) {
	Connect()

	if cmn.IsError(err) {
		return "DB CONNECT ERROR"
	}
	defer db.Close()
	ret = ""
	id := IsTokenValid(tok)
	if id == 0 {
		ret = "INVALID TOKEN"
		return
	}
	str := "update contact set first_name=?, last_name=?, organization=?, phone_number=?, email=?, website=? where id=?"
	//	str := fmt.Sprintf("update contact set ")

	//split structure electment by reflect
	//	v := reflect.ValueOf(contact).Elem()
	//	typeOfRequest := v.Type()

	//	for i = 0; i < v.NumField()-1; i++ {
	//		f := v.Field(i)
	//		name := typeOfRequest.Field(i).Name
	//		valType := f.Type().String()
	//		val := f.Interface()

	//		//		fmt.Printf("%d: %s %s = %v %v\n", i, name, valType, val, f.IsValid())
	//		if f.IsValid() {
	//			if v, yes := getColVal(valType, val); yes == true {
	//				colNames += fmt.Sprintf("%s,", strings.ToLower(name))
	//				colValues += fmt.Sprintf("%s,", v)
	//			}
	//		}
	//	}

	editContact, err := db.Prepare(str)
	if cmn.IsError(err) {
		return "EDIT CONTACT PREPARE ERROR"
	}

	res, err := editContact.Exec(contact.First_name, contact.Last_name, contact.Organization, contact.Phone_number, contact.Email, contact.Website, id)
	if cmn.IsError(err) {
		return "EDIT CONTACT ERROR"
	}

	affect, err := res.RowsAffected()
	if cmn.IsError(err) {
		return "EDIT CONTACT GET NOOFROWS ERROR"
	}

	if affect != 1 {
		ret = "INTERNAL DB ERROR MORE THAN ONE ROW AFFECTED"
		cmn.Log(cmn.DB, ret)
		return
	}
	return
}

func DeleteContact(tok string) (ret string) {
	Connect()

	if cmn.IsError(err) {
		return
	}
	defer db.Close()
	ret = ""
	id := IsTokenValid(tok)
	if id == 0 {
		ret = "INVALID TOKEN"
		return
	}

	delContact, err := db.Prepare("delete from contact where id=?")
	if cmn.IsError(err) {
		return "DELETE CONTACT PREPARE ERROR"
	}

	res, err := delContact.Exec(id)
	if cmn.IsError(err) {
		return "EDIT CONTACT ERROR"
	}

	affect, err := res.RowsAffected()
	if cmn.IsError(err) {
		return "DELETE CONTACT GET NOOFROWS ERROR"
	}

	if affect != 1 {
		ret = "INTERNAL DB ERROR MORE THAN ONE ROW AFFECTED"
		cmn.Log(cmn.DB, ret)
		return
	}
	return
}

//returns id if token is valid
func IsTokenValid(tok string) (id int) {
	getId := fmt.Sprintf("select id from %s where auth_token='%s'", dbCfg.Token, tok)

	// Execute the query to get id
	err = db.QueryRow(getId).Scan(&id)
	if cmn.IsError(err) {
		id = 0
	}
	return
}

func RefreshToken(req cmn.Token) (rsp cmn.Token, ret bool) {
	Connect()
	if cmn.IsError(err) {
		return
	}
	defer db.Close()

	ret = false
	queryStr := ""
	u, p := "", ""
	id := 0
	queryStr = fmt.Sprintf("SELECT id FROM %s where refresh_token = '%s'", dbCfg.Token, req.Refresh_token)
	// Execute the query to get id
	err = db.QueryRow(queryStr).Scan(&id)
	queryStr = fmt.Sprintf("SELECT username,password FROM %s where id = %d", dbCfg.User, id)
	// Execute the query to get id
	err = db.QueryRow(queryStr).Scan(&u, &p)

	if ret == false {
		return
	}
	rsp.Auth_token, rsp.Refresh_token, ret = GetTokens(u, p, id)
	return
}

func UpdatePassword(pass string, req cmn.Token) (ret bool) {
	Connect()
	if cmn.IsError(err) {
		return
	}
	defer db.Close()

	ret = false
	id := 0

	query := fmt.Sprintf("select id from %s where auth_token='%s'", dbCfg.Token, req.Auth_token)
	cmn.Log(cmn.DB, query)

	// Execute the query to get id
	err = db.QueryRow(query).Scan(&id)
	if cmn.IsError(err) {
		return
	}
	updatePass := fmt.Sprintf("update %s set password = '%s' where id = %d", dbCfg.User, pass, id)

	count, err := db.Exec(updatePass)

	if cmn.IsError(err) {
		return
	}
	rows, err := count.RowsAffected()

	if cmn.IsError(err) {
		return
	}
	if rows != 1 {
		cmn.Log(cmn.DB, "invalid token ")
		return
	}

	ret = true
	return
}

func VerifyUser(username, password string) (rsp cmn.Token, ret bool) {
	Connect()
	if cmn.IsError(err) {
		return
	}
	defer db.Close()

	ret = false
	str := fmt.Sprintf("SELECT password,id FROM %s where username = '%s'", dbCfg.User, username)
	pass := ""
	id := 0

	err = db.QueryRow(str).Scan(&pass, &id)

	if cmn.IsError(err) {
		return
	}

	if password != pass {
		cmn.Log(cmn.DB, "password mismatch for ", username)
		return
	}

	/*	lets say secret is fixed. this cannot work because users randomly sings out
			if we apply same secret then secret should be changed at single point username
			lets say if we cange secret for any user then at the time of token generation
			apply this secret and save token in database
		  if we have secret for each user then need to save many secrets
		so in database its like this: secret is enough to store
		case1: validate every http request with secret in database
		get secret
		decode username and password
		match username and password

		case2: validate every http request on arrival
		generate token every time
		match token

		so its like this:
		common secret for all users
		seperate secret for every user
	*/
	rsp.Auth_token, rsp.Refresh_token, ret = GetTokens(username, password, id)
	if false == ret {
		return
	}
	ret = true
	return
}

func GetTokens(username, password string, id int) (aToken, rToken string, ret bool) {
	ret = false

	//create auth and refresh tokens
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"password": password,
	})

	timeStr := strconv.Itoa(int(time.Now().UnixNano()))

	authSecret := SECRET + timeStr + AUTHTAIL
	refreshSecret := SECRET + timeStr + REFRESHTAIL

	aToken, err = token.SignedString([]byte(authSecret))
	if nil != err {
		return
	}
	rToken, err = token.SignedString([]byte(refreshSecret))
	if nil != err {
		return
	}

	//	save auth and refresh tokens in database
	cDate := time.Now().UnixNano()
	eDate := cDate + 31536000
	updateToken := fmt.Sprintf("insert into %s  (id,active,auth_token,refresh_token,createddate,expiredate) values(%d,1,'%s','%s',%v,%v)", dbCfg.Token, id, aToken, rToken, cDate, eDate)
	update, err := db.Query(updateToken)
	if cmn.IsError(err) {
		return
	}
	defer update.Close()
	ret = true
	return
}
