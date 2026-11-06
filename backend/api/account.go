package api

import (
	"os"
	"fmt"
	"log"
	"errors"
	"io/fs"
	"strings"
	"path/filepath"
	"encoding/json"
	"mud/database"
	"golang.org/x/crypto/bcrypt"
)

type Account struct {
	Username string
	PasswordHash string
}

type accountValidationError struct {
	message string
}

const ACCOUNT_FOLDER_NAME = "accounts"
const ACCOUNT_USERNAME_MIN_LENGTH = 3
const ACCOUNT_USERNAME_MAX_LENGTH = 20
const ACCOUNT_PASSWORD_MIN_LENGTH = 8
// bcrypt ignores everything past 72 bytes
const ACCOUNT_PASSWORD_MAX_LENGTH = 72

var errAccountUsernameTaken = errors.New("account username is taken")
var errAccountNotFound = errors.New("account not found")

func (validationError *accountValidationError) Error() string {
	return validationError.message
}

func AccountFolderPath() string {
	return filepath.Join(database.SAVES_FOLDER, ACCOUNT_FOLDER_NAME)
}

// Usernames are lowercased in the filename, so names that differ only by capitalization share a file.
// This is what makes usernames unique regardless of capitalization.
// The username must already have passed validateUsername.
func accountFilePath(username string) string {
	return filepath.Join(AccountFolderPath(), fmt.Sprintf("%s.json", strings.ToLower(username)))
}

// Usernames are used as filenames, so they are limited to ASCII letters, numbers and underscores.
func validateUsername(username string) error {
	if len(username) < ACCOUNT_USERNAME_MIN_LENGTH || len(username) > ACCOUNT_USERNAME_MAX_LENGTH {
		return &accountValidationError {
			message: fmt.Sprintf("Usernames must be between %d and %d characters long.", ACCOUNT_USERNAME_MIN_LENGTH, ACCOUNT_USERNAME_MAX_LENGTH),
		}
	}

	for _, character := range username {
		isLetter := (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z')
		isDigit := character >= '0' && character <= '9'
		if !isLetter && !isDigit && character != '_' {
			return &accountValidationError {
				message: "Usernames may only contain letters, numbers and underscores.",
			}
		}
	}

	return nil
}

func validatePassword(password string) error {
	if len(password) < ACCOUNT_PASSWORD_MIN_LENGTH || len(password) > ACCOUNT_PASSWORD_MAX_LENGTH {
		return &accountValidationError {
			message: fmt.Sprintf("Passwords must be between %d and %d characters long.", ACCOUNT_PASSWORD_MIN_LENGTH, ACCOUNT_PASSWORD_MAX_LENGTH),
		}
	}

	return nil
}

// Creates and saves a new account.
// Returns an *accountValidationError for a bad username or password,
// and errAccountUsernameTaken if an account with that username exists in any capitalization.
func createAccount(username string, password string) (*Account, error) {
	err := validateUsername(username)
	if err != nil {
		return nil, err
	}
	err = validatePassword(password)
	if err != nil {
		return nil, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	account := &Account {
		Username: username,
		PasswordHash: string(passwordHash),
	}

	err = database.WriteJsonAtomically(accountFilePath(username), account, false)
	if errors.Is(err, fs.ErrExist) {
		return nil, errAccountUsernameTaken
	}
	if err != nil {
		return nil, err
	}

	return account, nil
}

// Reads the account with this username in any capitalization.
// Returns errAccountNotFound if there is no such account.
func loadAccount(username string) (*Account, error) {
	// An invalid username can't have an account, and must not be used as a path
	err := validateUsername(username)
	if err != nil {
		return nil, errAccountNotFound
	}

	path := accountFilePath(username)
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, errAccountNotFound
	}
	if err != nil {
		log.Printf("Warn - failed to read account file %s: %s", path, err.Error())
		return nil, err
	}

	var account Account
	err = json.Unmarshal(data, &account)
	if err != nil {
		log.Printf("Warn - failed to parse account file %s: %s", path, err.Error())
		return nil, err
	}

	return &account, nil
}

func checkAccountPassword(account *Account, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(password))
	return err == nil
}
