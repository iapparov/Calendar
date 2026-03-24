package routers

import (
	_ "calendar/docs"
	"calendar/internal/web/handlers"
	"time"

	"calendar/internal/logger"

	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(engine *gin.Engine, hEvent *handlers.CalendarHandler, hUser *handlers.UserHandler, l *logger.Service) {

	engine.Use(gin.Recovery())
	engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	engine.Use(LoggerMiddleware(l))
	// Register routes
	api := engine.Group("/api/v1")

	api.GET("/swagger/*any", gin.WrapH(httpSwagger.WrapHandler))
	eventApi := api.Group("/events")
	authApi := api.Group("/auth")

	authApi.POST("/login", hUser.Login)
	authApi.POST("/register", hUser.Register)
	authApi.POST("/refresh-token", hUser.RefreshToken)

	eventApi.Use(AuthMiddleware(hUser.Auth()))
	eventApi.POST("", hEvent.CreateEvent)
	eventApi.PUT("", hEvent.UpdateEvent)
	eventApi.DELETE("", hEvent.DeleteEvent)
	eventApi.GET("/day", hEvent.EventsForDay)
	eventApi.GET("/week", hEvent.EventsForWeek)
	eventApi.GET("/month", hEvent.EventsForMonth)
}
