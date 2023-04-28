package config

import (
	"errors"
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/go-playground/validator/v10"
)

var errInvalidField = errors.New("invalid field")

type Conf struct {
	Server   server
	Postgres postgres
	Logger   logger
}

type server struct {
	ManagementUsername  string `toml:"ManagementUsername" validate:"min=5"`
	ManagementPassword  string `toml:"ManagementPassword" validate:"min=5"`
	Port                string `toml:"Port" validate:"numeric"`
	ReadTimeoutSeconds  int    `toml:"ReadTimeoutSeconds" validate:"gte=1,lte=300"`
	WriteTimeoutSeconds int    `toml:"WriteTimeoutSeconds" validate:"gte=1,lte=300"`
}

type postgres struct {
	ConnAddress  string `toml:"ConnAddress" validate:"min=10"`
	MaxOpenConns int    `toml:"MaxOpenConns" validate:"gte=1,lte=100"`
	MaxIdleConns int    `toml:"MaxIdleConns" validate:"gte=1,lte=100"`
	QueryTimeout int    `toml:"QueryTimeout" validate:"gte=2,lte=60"`
}

type logger struct {
	FileName   string `toml:"FileName" validate:"min=2"`
	MaxSizeMb  int    `toml:"MaxSizeMb" validate:"gte=1,lte=1000"`
	MaxBackups int    `toml:"MaxBackups" validate:"gte=1,lte=50"`
	MaxAgeDays int    `toml:"MaxAgeDays" validate:"gte=1,lte=720"`
}

func Get(fileName string) (*Conf, error) {
	var conf *Conf
	if _, err := toml.DecodeFile(fileName, &conf); err != nil {
		return nil, fmt.Errorf("decode file: %w", err)
	}

	if err := validator.New().Struct(*conf); err != nil {
		var vErrors validator.ValidationErrors
		if errors.As(err, &vErrors) {
			if err := checkValidatorErr(vErrors); err != nil {
				return nil, fmt.Errorf("validator: check err: %w", err)
			}

			return nil, fmt.Errorf("validator: %w", err)
		}
	}

	return conf, nil
}

func checkValidatorErr(errs validator.ValidationErrors) error {
	for _, err := range errs {
		return fmt.Errorf("%w: %s(%s): see it <%v> want <%s=%s>",
			errInvalidField,
			err.StructNamespace(),
			err.Type(),
			err.Value(),
			err.ActualTag(),
			err.Param())
	}

	return nil
}