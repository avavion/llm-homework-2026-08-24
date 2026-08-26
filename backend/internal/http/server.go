package http

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"llm-homework/backend/internal/account"
	"llm-homework/backend/internal/auth"
	"llm-homework/backend/internal/product"
	"llm-homework/backend/internal/recipe"
	"llm-homework/backend/internal/recognition"
	"llm-homework/backend/internal/regulation"
)

func NewServer(db *sql.DB) http.Handler {
	router := gin.New()
	router.Use(gin.Recovery())

	router.GET("/healthz", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/json", []byte("{\"status\":\"ok\"}\n"))
	})

	authService := auth.NewService(account.NewRepository(db))
	router.Any("/v1/auth/*action", gin.WrapH(auth.NewHandler(authService)))

	resolveAccount := func(c *gin.Context) (uuid.UUID, bool) {
		current, ok := auth.GinAccount(c)
		if !ok {
			return uuid.UUID{}, false
		}
		return current.ID, true
	}

	authenticated := router.Group("/", auth.GinRequireSession(authService))

	productService := product.NewService(product.NewRepository(db))
	product.RegisterRoutes(authenticated, productService, resolveAccount)

	recipeService := recipe.NewService(regulation.NewRepository())
	recipe.RegisterRoutes(authenticated, recipeService, productService, resolveAccount)

	recognitionService := recognition.NewService(
		recognition.NewRepository(db),
		recognition.UnconfiguredOCRClient{},
		recognition.UnconfiguredLLMClient{},
	)
	recognition.RegisterRoutes(authenticated, recognitionService, resolveAccount)

	return router
}
