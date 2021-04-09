//Echo API api. Used to create HTTP server layer, handlers, and swagger

package api

import (
	"MessageBox/docs"
	"MessageBox/service"
	. "MessageBox/util/logger"
	"MessageBox/util/loggerFactory"
	. "MessageBox/version"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// @contact.name Joshua Brummet
// @contact.url http://github.com/brummetj
// @contact.email Joshua.Brummet@gmail.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
func RegisterEchoServer(service service.MessageBox) error {
	var config = service.Base.AppConfig
	docs.SwaggerInfo.Title = "MessageBox API"
	docs.SwaggerInfo.Version = Version

	// PORT is hardcoded to the proxy server. We don't really want this changing often anyway :)
	docs.SwaggerInfo.Host = fmt.Sprintf("%v:3001", config.Host)
	docs.SwaggerInfo.BasePath = config.Basepath
	docs.SwaggerInfo.Schemes = []string{config.Protocol}
	e := echo.New()
	e.Use(loggerFactory.ZapEchoHandler())
	e.Use(middleware.Recover())
	Log.Info("Creating routing handlers")
	err := CreateHandlers(e, service, config.Basepath)
	if err != nil {
		return errors.Wrap(err, "Creating echo handlers")
	}
	Log.Infof("Starting echo server on: http://%v:%v", config.Host, config.Port)
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%v", config.Port)))
	return nil
}

func CreateHandlers(e *echo.Echo, service service.MessageBox, basePath string) error {
	e.POST(fmt.Sprintf("%v/users", basePath), func(c echo.Context) error {
		return service.CreateUser(c)
	})
	e.GET(fmt.Sprintf("%v/users/:username/mailbox", basePath), func(c echo.Context) error {
		return service.GetUserMessages(c)
	})
	e.POST(fmt.Sprintf("%v/messages", basePath), func(c echo.Context) error {
		return service.SendMessage(c)
	})
	e.GET(fmt.Sprintf("%v/messages/:id", basePath), func(c echo.Context) error {
		return service.GetMessage(c)
	})
	e.GET(fmt.Sprintf("%v/messages/:id/replies", basePath), func(c echo.Context) error {
		return service.GetReply(c)
	})
	e.POST(fmt.Sprintf("%v/messages/:id/replies", basePath), func(c echo.Context) error {
		return service.ReplyMessage(c)
	})
	e.POST(fmt.Sprintf("%v/groups", basePath), func (c echo.Context) error {
		return service.CreateGroups(c)
	} )
	e.GET("/swagger/*", echoSwagger.WrapHandler)
	return nil
}

// used for docs only
type HTTPError struct {
	message interface{}
	code int64
	internal error
}