package http

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"llm-homework/backend/internal/account"
	"llm-homework/backend/internal/auth"
	"llm-homework/backend/internal/product"
	"llm-homework/backend/internal/recipe"
	"llm-homework/backend/internal/recognition"
	"llm-homework/backend/internal/regulation"
)

func NewServer(db *sql.DB, allowedOrigins []string, recognitionProvider string) http.Handler {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(cors(allowedOrigins))

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
	rules := regulation.NewRepository()
	product.RegisterRoutes(authenticated, productService, resolveAccount, product.DisplayStatusFunc(func(item product.Product) product.DisplayStatus {
		return displayStatusFor(item, rules, time.Now())
	}))

	recipeService := recipe.NewService(rules)
	recipe.RegisterRoutes(authenticated, recipeService, productService, resolveAccount)

	ocrClient, llmClient := recognition.Clients(recognitionProvider)
	recognitionService := recognition.NewService(
		recognition.NewRepository(db),
		ocrClient,
		llmClient,
	)
	recognition.RegisterRoutes(authenticated, recognitionService, resolveAccount)

	return router
}

type regulationRuleLookup interface {
	RuleFor(countryCode string, dateType product.DateType) (regulation.Rule, bool)
}

// displayStatusFor returns a regulation-backed display status without making
// an unsupported food-safety or legal inference. Terminal lifecycle states
// always win; every absent, unconfirmed, or unusable rule is research_required.
func displayStatusFor(item product.Product, rules regulationRuleLookup, now time.Time) product.DisplayStatus {
	switch item.LifecycleStatus {
	case product.LifecycleUsed:
		return product.DisplayStatusUsed
	case product.LifecycleDiscarded:
		return product.DisplayStatusDiscarded
	}

	if item.CountryCode == nil {
		return product.DisplayStatusResearchRequired
	}
	rule, ok := rules.RuleFor(*item.CountryCode, item.DateType)
	if !ok {
		return product.DisplayStatusResearchRequired
	}

	status, err := regulation.Evaluate(item, rule, now)
	if err != nil {
		return product.DisplayStatusResearchRequired
	}
	switch status {
	case regulation.StatusActive:
		return product.DisplayStatusActive
	case regulation.StatusAttention:
		return product.DisplayStatusAttention
	case regulation.StatusExpired:
		return product.DisplayStatusExpired
	default:
		return product.DisplayStatusResearchRequired
	}
}
