package server

import (
	"errors"
	"strings"
	"sync"

	"github.com/alexedwards/argon2id"
	"github.com/apeunit/LaunchControlD/pkg/utils"

	normalizer "github.com/dimuska139/go-email-normalizer"
)

var (
	ErrorDuplicatedUser  = errors.New("duplicated user")
	ErrorEmptyEmailOrPwd = errors.New("username and password should not be empty")
)

// User a user in the user database
type User struct {
	Email        string
	PasswordHash string
}

// UserDb keep the users database
type UsersDB struct {
	dbPath    string
	users     map[string]User
	emailNorm *normalizer.Normalizer
	sync.RWMutex
}

func NewUserDB(dbPath string) (db *UsersDB, err error) {
	db = &UsersDB{
		dbPath:    dbPath,
		users:     make(map[string]User),
		emailNorm: normalizer.NewNormalizer(),
	}
	err = db.Load()
	return
}

func (db *UsersDB) RegisterUser(email, pass string) (err error) {
	// basic length check
	if len(strings.TrimSpace(email)) == 0 || len(strings.TrimSpace(pass)) == 0 {
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
	err = db.Store()
	return
}

// IsAuthorized verify if a use is authorized
func (db *UsersDB) IsAuthorized(email, pass string) bool {
	db.RLock()
	defer db.RUnlock()
	// normalize the email
	emailH := utils.Hash(db.emailNorm.Normalize(email))
	// retrieve the user
	u, xs := db.users[emailH]
	// if doesn't exists is not authorized
	if !xs {
		return false
	}
	// verify the password, if there is an error the user is not Authorized anyway
	match, _ := argon2id.ComparePasswordAndHash(pass, u.PasswordHash)
	return match
}

// Store store the db user on file
func (db *UsersDB) Store() (err error) {
	db.Lock()
	defer db.Unlock()
	err = utils.StoreJSON(db.dbPath, db.users)
	return
}

// Load load the user db from file or get an empty db if the
// file does not exists
func (db *UsersDB) Load() (err error) {
	db.Lock()
	defer db.Unlock()
	// if does not exists there is nothing to load
	if !utils.FileExists(db.dbPath) {
		return
	}
	err = utils.LoadJSON(db.dbPath, db.users)
	return
}
