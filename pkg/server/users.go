package server

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/apeunit/LaunchControlD/pkg/utils"
	cache "github.com/patrickmn/go-cache"

	normalizer "github.com/dimuska139/go-email-normalizer"
	log "github.com/sirupsen/logrus"
)

// error definitions
var (
	ErrorDuplicatedUser  = errors.New("duplicated user")
	ErrorEmptyEmailOrPwd = errors.New("username and password should not be empty")
	ErrorUnauthorized    = errors.New("user not authorized")
	ErrorTokenNotFound   = errors.New("token not found")
)

// User a user in the user database
type User struct {
	Email        string
	PasswordHash string
}

// UsersDB keep the users database
type UsersDB struct {
	dbPath    string
	users     map[string]User
	tokens    *cache.Cache
	emailNorm *normalizer.Normalizer
	sync.RWMutex
}

// NewUserDB create or read an existing database from a path
func NewUserDB(dbPath string) (db *UsersDB, err error) {
	log.Debug("usersDb: initialize new db at: ", dbPath)
	db = &UsersDB{
		dbPath:    dbPath,
		users:     make(map[string]User),
		tokens:    cache.New(12*time.Hour, 1*time.Hour),
		emailNorm: normalizer.NewNormalizer(),
	}
	err = db.load()
	log.Debugln("usersDb: db loaded with", len(db.users), "records")
	return
}

// RegisterUser register a new user into the user database
func (db *UsersDB) RegisterUser(email, pass string) (err error) {
	log.Debugln("usersDb: register user", email)
	// basic length check
	if strings.TrimSpace(email) == "" || strings.TrimSpace(pass) == "" {
		err = ErrorEmptyEmailOrPwd
		return
	}
	// lock for writing
	db.Lock()
	defer db.Unlock()
	// normalize and hash the email
	emailH := utils.Hash(db.emailNorm.Normalize(email))
	// if exists return error
	if _, xs := db.users[emailH]; xs {
		err = ErrorDuplicatedUser
		return
	}
	// otherwise calculate the argon2id hash
	pwdH, err := argon2id.CreateHash(pass, argon2id.DefaultParams)
	if err != nil {
		return
	}
	// store the new user
	db.users[emailH] = User{
		Email:        email,
		PasswordHash: pwdH,
	}
	// save the db to a file
	err = db.store()
	return
}

// IsAuthorized verify if a use is authorized,
// if so, register a user token and returns it
func (db *UsersDB) IsAuthorized(email, pass string) (token string, err error) {
	db.RLock()
	defer db.RUnlock()
	// normalize the email
	emailH := utils.Hash(db.emailNorm.Normalize(email))
	// retrieve the user
	u, xs := db.users[emailH]
	// if doesn't exists is not authorized
	if !xs {
		err = ErrorUnauthorized
		return
	}
	// verify the password, if there is an error the user is not Authorized anyway
	match, err := argon2id.ComparePasswordAndHash(pass, u.PasswordHash)
	if err != nil {
		log.Error("error authorizing: ", err)
		err = ErrorUnauthorized
		return
	}
	if !match {
		err = ErrorUnauthorized
		return
	}
	// if all is good generate a token
	token, err = utils.GenerateRandomHash()
	if err != nil {
		log.Error("randome geneation failed (check os): ", err)
		err = ErrorUnauthorized
		return
	}
	db.tokens.Set(token, emailH, cache.DefaultExpiration)
	return
}

// IsTokenAuthorized check whenever the token exists and returns the associated email
// otherwise returns error
func (db *UsersDB) IsTokenAuthorized(token string) (email string, err error) {
	email, err = db.GetEmailFromToken(token)
	if err != nil {
		err = ErrorUnauthorized
	}
	return
}

// GetEmailFromToken retrieve the email associated to a token
func (db *UsersDB) GetEmailFromToken(token string) (email string, err error) {
	emailHI, found := db.tokens.Get(token)
	if !found {
		err = ErrorTokenNotFound
		return
	}
	user, found := db.users[emailHI.(string)]
	// if it is not found the db is inconsistent
	if !found {
		err = ErrorTokenNotFound
		return
	}
	email = user.Email
	return
}

// DropToken deletes a token
func (db *UsersDB) DropToken(token string) {
	db.tokens.Delete(token)
}

// Store store the db user on file
func (db *UsersDB) store() (err error) {
	err = utils.StoreJSON(db.dbPath, db.users)
	return
}

// Load load the user db from file or get an empty db if the
// file does not exists
func (db *UsersDB) load() (err error) {
	// if does not exists there is nothing to load
	if !utils.FileExists(db.dbPath) {
		return
	}
	err = utils.LoadJSON(db.dbPath, &db.users)
	return
}
