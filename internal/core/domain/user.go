package domain

import (
	"fmt"
	"regexp"

	core_errors "github.com/Rics69/x-chat/internal/core/errors"
)

// компилим один раз на старте, а не на каждый Validate()
var phoneNumberRegexp = regexp.MustCompile(`^\+[0-9]{9,14}$`)

type User struct {
	ID      int
	Version int

	FullName    string
	PhoneNumber *string
}

func NewUser(id int, version int, fullname string, phoneNumber *string) User {
	return User{
		ID:          id,
		Version:     version,
		FullName:    fullname,
		PhoneNumber: phoneNumber,
	}
}

func NewUserUnitialized(fullName string, phoneNumber *string) User {
	return NewUser(UnitializedID, UnitializedVersion, fullName, phoneNumber)
}

func (u *User) Validate() error {
	fullNameLength := len([]rune(u.FullName))
	if fullNameLength < 3 || fullNameLength > 100 {
		return fmt.Errorf("invalid `FullName` len: %d: %w", fullNameLength, core_errors.ErrInvalidArgument)
	}

	if u.PhoneNumber != nil {
		phoneNumberLength := len([]rune(*u.PhoneNumber))
		if phoneNumberLength < 10 || phoneNumberLength > 15 {
			return fmt.Errorf("invalid `PhoneNumber` len: %d: %w", phoneNumberLength, core_errors.ErrInvalidArgument)
		}

		if !phoneNumberRegexp.MatchString(*u.PhoneNumber) {
			return fmt.Errorf("invalid `PhoneNumber` format: %w", core_errors.ErrInvalidArgument)
		}
	}

	return nil
}

type UserPatch struct {
	Fullname    Nullable[string]
	PhoneNumber Nullable[string]
}

func NewUserPatch(fullName Nullable[string], phoneNumber Nullable[string]) UserPatch {
	return UserPatch{
		Fullname:    fullName,
		PhoneNumber: phoneNumber,
	}
}

func (p *UserPatch) Validate() error {
	if p.Fullname.Set && p.Fullname.Value == nil {
		return fmt.Errorf("'Fullname' can't be patched to NULL: %w", core_errors.ErrInvalidArgument)
	}

	return nil
}

func (u *User) ApplyPatch(patch UserPatch) error {
	if err := patch.Validate(); err != nil {
		return fmt.Errorf("validate user patch: %w", err)
	}

	tmp := *u

	if patch.Fullname.Set {
		tmp.FullName = *patch.Fullname.Value
	}

	if patch.PhoneNumber.Set {
		tmp.PhoneNumber = patch.PhoneNumber.Value
	}

	if err := tmp.Validate(); err != nil {
		return fmt.Errorf("valite  patched user: %w", err)
	}

	*u = tmp

	return nil
}
