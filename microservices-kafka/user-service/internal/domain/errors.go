package domain

import "errors"

var (
    ErrUserNotFound = errors.New("user not found")
    ErrEmailExists  = errors.New("email already exists")
    ErrInvalidID    = errors.New("invalid user id")
)