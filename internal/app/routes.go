package app

import (
	"time"

	"Worker/internal/auth"
	"Worker/internal/cache"
	"Worker/internal/config"
	"Worker/internal/handlers"
	"Worker/internal/repo"
	"Worker/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/swag"
)

// Setup registers all routes on the given engine.
func Setup(r *gin.Engine, cfg config.Config, db *pgxpool.Pool, rdb *redis.Client) {
	r.GET("/", rootHandler(cfg))
	r.GET("/health", healthHandler(cfg))
	r.GET("/version", versionHandler(cfg))
	r.GET("/swagger-doc.json", swaggerDocHandler())
	r.GET("/swagger", func(c *gin.Context) { c.Redirect(302, "/swagger/index.html") })
	r.GET("/swagger/*any", ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.URL("/swagger-doc.json"),
		ginSwagger.DefaultModelsExpandDepth(-1),
		ginSwagger.PersistAuthorization(true),
	))

	api := r.Group("/api/v1")

	sessionStore := auth.NewStore(rdb, 24*time.Hour)
	userRepo := repo.NewPGUserRepo(db)
	userSvc := service.NewUserService(userRepo)
	authHandler := handlers.NewAuthHandler(sessionStore, userSvc)
	registerAuthRoutes(api, authHandler)

	protected := api.Group("", auth.RequireSession(sessionStore))
	todoRepo := repo.NewPGTodoRepo(db)
	todoCache := cache.NewTodoCache(rdb, cfg.Redis.DefaultTTL)
	todoSvc := service.NewTodoService(todoRepo, todoCache)
	todoHandler := handlers.NewTodoHandler(todoSvc)
	registerTodoRoutes(protected, todoHandler)

}

func rootHandler(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service": "Todo API",
			"version": cfg.App.Version,
			"env":     cfg.App.Env,
			"docs":    "/swagger/index.html",
			"spec":    "/swagger-doc.json",
			"health": "/health",
			"api":    "/api/v1",
		})
	}
}

func healthHandler(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true, "env": cfg.App.Env})
	}
}

func versionHandler(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{"version": cfg.App.Version})
	}
}

func swaggerDocHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		doc, err := swag.ReadDoc("swagger")
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.Data(200, "application/json; charset=utf-8", []byte(doc))
	}
}

func registerTodoRoutes(api *gin.RouterGroup, h *handlers.TodoHandler) {
	api.POST("/todos", h.Create)
	api.GET("/todos", h.List)
	api.GET("/todos/search", h.Search)
	api.GET("/todos/overdue", h.Overdue)
	api.GET("/todos/:id", h.GetByID)
	api.PATCH("/todos/:id", h.Update)
	api.DELETE("/todos/:id", h.Delete)
	api.POST("/todos/:id/complete", h.Complete)
}

func registerAuthRoutes(api *gin.RouterGroup, h *handlers.AuthHandler) {
	api.POST("/auth/login", h.Login)
	api.POST("/auth/register", h.Register)
	api.POST("/auth/logout", h.Logout)
}
