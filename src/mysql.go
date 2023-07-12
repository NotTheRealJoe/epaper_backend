package epaper_backend

import (
	"database/sql"
	"errors"
	"log"
	"math/rand"
)

const AUTH_CODE_LENGTH = 8
const AUTH_CODE_ALLOWED_CHARACTERS = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-._~[]@!$()*,;"
const USER_COOKIE_LENGTH = 60

const CONNECTION_VERIFY_STRING = "Y|EPi0`^x~C,dvFhT(>C0&pWiXESk*'&g;V3o}Xu38U[89&t+19!6G+;C4j>S\\S6peO\"bo/=@p}qY\"xrAPkZyY!.v_EcKZ]Dq\\kn"

type MysqlRepository struct {
	db *sql.DB
}

func CreateMysqlRepository(db *sql.DB) MysqlRepository {
	return MysqlRepository{
		db: db,
	}
}

func (r MysqlRepository) CheckConnection() bool {
	row := r.db.QueryRow("SELECT * FROM connection_verify")
	var result string
	var comment string
	row.Scan(&result, &comment)
	if result == CONNECTION_VERIFY_STRING {
		log.Printf("%s: %s\n", "DB connection verify succeeded", comment)
		return true
	}
	return false
}

// CreateAuthorization generates and authorization code, saves it to the database, and returns it
func (r MysqlRepository) CreateAuthorization() string {
	authCode := ""
	for i := 0; i < AUTH_CODE_LENGTH; i++ {
		authCode = authCode + string(AUTH_CODE_ALLOWED_CHARACTERS[rand.Intn(len(AUTH_CODE_ALLOWED_CHARACTERS))])
	}

	_, err := r.db.Exec("INSERT INTO authorizations (authorization) VALUES (?)", authCode)
	if err != nil {
		log.Fatal(err)
	}

	return authCode
}

func (r MysqlRepository) UseAuthorization(authCode string) (ok bool, cookie *string) {
	row := r.db.QueryRow("SELECT * FROM authorizations WHERE authorization = ? AND user_cookie IS NULL", authCode)
	result := Authorzation{}
	err := row.Scan(&result.ID, &result.AuthCode, &result.UserCookie)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		} else {
			log.Fatal(err)
		}
	}

	userCookie := ""
	for i := 0; i < USER_COOKIE_LENGTH; i++ {
		authCode = userCookie + string(AUTH_CODE_ALLOWED_CHARACTERS[rand.Intn(len(AUTH_CODE_ALLOWED_CHARACTERS))])
	}

	_, err = r.db.Exec("UPDATE authorizations SET user_cookie = ? WHERE authorization = ?", userCookie, authCode)
	if err != nil {
		log.Fatal(err)
	}

	return true, &userCookie
}
