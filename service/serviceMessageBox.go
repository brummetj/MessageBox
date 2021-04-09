// Service handler functions. Modify to add new protocol controllers

package service

import (
	"MessageBox/controller"
	"github.com/labstack/echo/v4"
)

type MessageBox struct {
	Base *Base
	httpController controller.MessageBox
}
func NewMessageBox(base *Base) *MessageBox {
	return &MessageBox{Base: base}
}

func (mb *MessageBox) CreateUser(ctx interface{}) error {
	if mb.Base.AppConfig.Protocol == "http" {
		cpy, _ := ctx.(echo.Context)
		return mb.httpController.CreateUser(cpy)
	}
	return nil
}

func (mb *MessageBox) GetUserMessages(ctx interface{}) error {
	if mb.Base.AppConfig.Protocol == "http" {
		cpy, _ := ctx.(echo.Context)
		err := mb.httpController.GetUserMessages(cpy)
		if err != nil {
			return err
		}
	}
	return nil
}

func (mb *MessageBox) CreateGroups(ctx interface{}) error {
	if mb.Base.AppConfig.Protocol == "http" {
		cpy, _ := ctx.(echo.Context)
		err := mb.httpController.CreateGroups(cpy)
		if err != nil {
			return err
		}
	}
	return nil
}

func (mb *MessageBox) GetMessage(ctx interface{}) error {
	if mb.Base.AppConfig.Protocol == "http" {
		cpy, _ := ctx.(echo.Context)
		err := mb.httpController.GetMessage(cpy)
		if err != nil {
			return err
		}
	}
	return nil
}

func (mb *MessageBox) GetReply(ctx interface{}) error {
	if mb.Base.AppConfig.Protocol == "http" {
		cpy, _ := ctx.(echo.Context)
		err := mb.httpController.GetReplies(cpy)
		if err != nil {
			return err
		}
	}
	return nil
}

func (mb *MessageBox) SendMessage(ctx interface{}) error {
	if mb.Base.AppConfig.Protocol == "http" {
		cpy, _ := ctx.(echo.Context)
		err := mb.httpController.SendMessage(cpy)
		if err != nil {
			return err
		}
	}
	return nil
}

func (mb *MessageBox) ReplyMessage(ctx interface{}) error {
	if mb.Base.AppConfig.Protocol == "http" {
		cpy, _ := ctx.(echo.Context)
		err := mb.httpController.ReplyMessage(cpy)
		if err != nil {
			return err
		}
	}
	return nil
}
