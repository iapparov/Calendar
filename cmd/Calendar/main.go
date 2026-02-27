// @title           Calendar API
// @version         1.0
// @description     API для управления календарем событий.
// @termsOfService  http://swagger.io/terms/
// @BasePath        /

package main

import (
	"calendar/internal/app"
)

func main() {
	app.NewApp().Run()
}
