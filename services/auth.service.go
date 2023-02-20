package services

import "github.com/ivandi1980/my-gomongo/models"

type AuthService interface {
	SignUpUser(*models.SignUpInput) (*models.UserDBResponse, error)
	SignInUser(*models.SignInInput) (*models.UserDBResponse, error)
}
