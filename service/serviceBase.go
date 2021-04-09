package service

import (
	. "MessageBox/config"
	"MessageBox/util/loggerFactory"
	"github.com/pkg/errors"
)

type Base struct {
	AppConfig *AppConfig
}

func (sb *Base) Init(filename string) error {
	var err error
	cfg, err := ReadConfig(filename)
	if err != nil {
		return errors.Wrap(err, "Unable to load config")
	}
	sb.AppConfig = cfg
	err = loggerFactory.RegisterLogger(sb.AppConfig.AppLogger)
	if err != nil {
		return errors.Wrap(err, "loading logger")
	}
	return nil
}