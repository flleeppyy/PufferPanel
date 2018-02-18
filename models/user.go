package models

import (
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"time"
	"github.com/markbates/pop"
	"github.com/markbates/validate"
	"github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

type User struct {
	ID             uuid.UUID `db:id`
	Username       string    `db:username`
	Email          string    `db:email`
	HashedPassword string    `db:string`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

type Users []User

func (u *User) ChangePassword(newPw string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPw), bcrypt.DefaultCost)
	if err != nil {
		return errors.WithStack(err)
	}

	u.HashedPassword = string(hash)
	return nil
}

func (u *User) ValidatePassword(testPw string) bool {
	res := bcrypt.CompareHashAndPassword([]byte(u.HashedPassword), []byte(testPw))
	//lib returns an error if they are not the same, so we check to see if it's null
	return res == nil
}

func (u *User) Validate(tx *pop.Connection) (*validate.Errors, error) {
	//validate id, username, email, and password are set
	validationErrors := validate.NewErrors()

	err := validation.ValidateStruct(&u,
		validation.Field(&u.Email, validation.Required, is.Email),
		validation.Field(&u.Username, validation.Required),
		validation.Field(&u.HashedPassword, validation.Required),
	)

	errs, ok := err.(validation.Errors)

	if ok && (err != nil && errs.Filter() != nil) {
		for k, v := range errs {
			validationErrors.Add(k, v.Error())
		}
	}

	return validationErrors, nil
}

func (u *User) BeforeCreate(tx *pop.Connection) error {
	validateEmailUser := &User {
		Email: u.Email,
	}

	count, err := tx.Count(validateEmailUser)

	if err != nil {
		return err
	}

	if count > 0 {
		return errors.New("email already in use")
	}

	validateUsernameUser := &User {
		Username: u.Username,
	}

	count, err = tx.Count(validateUsernameUser)

	if err != nil {
		return err
	}

	if count > 0 {
		return errors.New("username already in use")
	}

	return nil
}