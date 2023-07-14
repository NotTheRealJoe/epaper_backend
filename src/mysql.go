package epaper_backend

import (
	"database/sql"
	"errors"
	"log"
	"math/rand"
	"time"
)

const AUTH_CODE_LENGTH = 8
const AUTH_CODE_ALLOWED_CHARACTERS = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-._~[]@!$()*,"
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

// UseAuthorization takes in a URL authorization code and uses it. Using it means marking it so that it cannot be used
// by another client, and issuing a cookie that the client who presented the authCode can continue to be authorized by.
// If the authCode is invalid, function returns (ok=false, cookie=nil). If the authCode is valid, its status is updated
// in the database, a cookie is generated, and (ok=true, cookie=<the newly generated cookie>) is returned.
func (r MysqlRepository) UseAuthorization(authCode string) (ok bool, cookie *string) {
	row := r.db.QueryRow("SELECT `id`, `authorization` FROM `authorizations` WHERE `authorization` = ? AND user_cookie IS NULL", authCode)
	result := Authorzation{}
	err := row.Scan(&result.ID, &result.AuthCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		} else {
			log.Fatal(err)
		}
	}

	// generate a random userCookie
	userCookie := ""
	for i := 0; i < USER_COOKIE_LENGTH; i++ {
		userCookie = userCookie + string(AUTH_CODE_ALLOWED_CHARACTERS[rand.Intn(len(AUTH_CODE_ALLOWED_CHARACTERS))])
	}

	// save the new userCookie
	_, err = r.db.Exec("UPDATE authorizations SET user_cookie = ?, use_started=NOW() WHERE authorization = ?", userCookie, authCode)
	if err != nil {
		log.Fatal(err)
	}

	// return the generated cookie
	return true, &userCookie
}

func (r MysqlRepository) CookieIsValid(cookie string) bool {
	row := r.db.QueryRow("SELECT COUNT(*) FROM `authorizations` WHERE `user_cookie` = ?", cookie)
	var result int
	err := row.Scan(&result)
	if err != nil {
		log.Fatal(err)
	}
	return result > 0
}

func (r MysqlRepository) GetAuthorizationByCookie(cookie string) (ok bool, authorization *Authorzation) {
	authorization = &Authorzation{}
	row := r.db.QueryRow("SELECT `id`, `authorization`, `use_started`, `user_cookie` FROM `authorizations` WHERE `user_cookie` = ?", cookie)
	err := row.Scan(&authorization.ID, &authorization.AuthCode, &authorization.UseStarted, &authorization.UserCookie)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ok = false
			return
		} else {
			log.Fatal(err)
		}
	}
	ok = true
	return
}

func (r MysqlRepository) SaveDrawing(d Drawing) (int64, error) {
	if d.DateCreated == "" {
		d.DateCreated = time.Now().Format(time.RFC3339)
	}

	res, err := r.db.Exec(
		"INSERT INTO `drawings` (`date_created`, `author`, `data`, `authorization`) VALUES (NOW(), ?, ?, ?)",
		d.Author,
		d.Data,
		d.Authorization,
	)
	if err != nil {
		return -1, err
	}

	return res.LastInsertId()
}

func (r MysqlRepository) GetAllDrawingsRemovedLast() (*[]Drawing, error) {
	rows, err := r.db.Query("SELECT `id`,`date_created`,`author`,`data`,`authorization`,`removed` FROM `drawings` ORDER BY `removed` ASC, `date_created` DESC")
	if err != nil {
		return nil, err
	}

	drawings := []Drawing{}
	for rows.Next() {
		d := Drawing{}
		err := rows.Scan(&d.ID, &d.DateCreated, &d.Author, &d.Data, &d.Authorization, &d.Removed)
		if err != nil {
			return nil, err
		}
		drawings = append(drawings, d)
	}

	return &drawings, nil
}

func (r MysqlRepository) RemoveAuthorization(id int) {
	r.db.Exec("DELETE FROM `authorizations` WHERE id = ?", id)
}
