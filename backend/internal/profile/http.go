package profile

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AccountResolver func(c *gin.Context) (uuid.UUID, bool)

type PublicProfile struct {
	CountryCode    string `json:"country_code"`
	Language       string `json:"language"`
	RegulatorGroup string `json:"regulator_group,omitempty"`
}

type NotificationSettingsResponse struct {
	Settings []Setting `json:"settings"`
}

type profileRequest struct {
	CountryCode string `json:"country_code"`
	Language    string `json:"language"`
}

type notificationSettingRequest struct {
	ProductGroup          string `json:"product_group"`
	AlertThresholdMinutes int    `json:"alert_threshold_minutes"`
}

func RegisterRoutes(router gin.IRouter, service *Service, resolveAccount AccountResolver) {
	router.GET("/v1/profile", getProfile(service, resolveAccount))
	router.PUT("/v1/profile", putProfile(service, resolveAccount))
	router.GET("/v1/notification-settings", getSettings(service, resolveAccount))
	router.PUT("/v1/notification-settings", putSetting(service, resolveAccount))
}

func getProfile(service *Service, resolveAccount AccountResolver) gin.HandlerFunc {
	return func(c *gin.Context) {
		accountID, ok := resolveAccount(c)
		if !ok {
			return
		}
		profile, err := service.Profile(c.Request.Context(), accountID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		c.JSON(http.StatusOK, PublicProfile{CountryCode: profile.CountryCode, Language: profile.Language, RegulatorGroup: service.RegulatorGroup(profile.CountryCode)})
	}
}

func putProfile(service *Service, resolveAccount AccountResolver) gin.HandlerFunc {
	return func(c *gin.Context) {
		accountID, ok := resolveAccount(c)
		if !ok {
			return
		}
		var request profileRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid profile"})
			return
		}
		profile, err := service.SaveProfile(c.Request.Context(), accountID, ProfileInput{CountryCode: request.CountryCode, Language: request.Language})
		if err == ErrInvalidProfile {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid profile"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		c.JSON(http.StatusOK, PublicProfile{CountryCode: profile.CountryCode, Language: profile.Language, RegulatorGroup: service.RegulatorGroup(profile.CountryCode)})
	}
}

func getSettings(service *Service, resolveAccount AccountResolver) gin.HandlerFunc {
	return func(c *gin.Context) {
		accountID, ok := resolveAccount(c)
		if !ok {
			return
		}
		settings, err := service.Settings(c.Request.Context(), accountID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		c.JSON(http.StatusOK, NotificationSettingsResponse{Settings: settings})
	}
}

func putSetting(service *Service, resolveAccount AccountResolver) gin.HandlerFunc {
	return func(c *gin.Context) {
		accountID, ok := resolveAccount(c)
		if !ok {
			return
		}
		var request notificationSettingRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification setting"})
			return
		}
		if err := service.SaveSetting(c.Request.Context(), accountID, request.ProductGroup, request.AlertThresholdMinutes); err == ErrInvalidSetting {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification setting"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		settings, err := service.Settings(c.Request.Context(), accountID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		c.JSON(http.StatusOK, NotificationSettingsResponse{Settings: settings})
	}
}
