package handler

type HTTPError struct {
	Error   string `json:"error,omitempty"`
	Comment string `json:"comment,omitempty"`
}

// 401
const (
	typeAuthUsernameRequired      = "AUTH_USERNAME_REQUIRED"
	typeAuthPasswordRequired      = "AUTH_PASSWORD_REQUIRED"
	typeAuthUsernameAlreadyExists = "AUTH_USERNAME_ALREADY_EXISTS"

	typeTaskNotFound = "TASK_NOT_FOUND"

	typeParametersRequired = "PARAMETERS_REQUIRED"

	typeParameterRequired = "PARAMETER_REQUIRED"
	typeParameterTooLong  = "PARAMETER_TOO_LONG"
)

// 500
const (
	typeInternalError = "INTERNAL"
)
