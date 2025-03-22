package web

import (
	"net/http"
	"time"

	"github.com/bilbothegreedy/HNS/internal/models"
	"github.com/bilbothegreedy/HNS/internal/repository"
	"github.com/gin-gonic/gin"
)

func GetProfilePage(userRepo repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		usernameRaw, exists := c.Get("username")
		if !exists {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		username, _ := usernameRaw.(string)
		user, err := userRepo.GetByUsername(c.Request.Context(), username)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.HTML(http.StatusOK, "base.html", gin.H{
			"Title":       "User Profile",
			"Username":    user.Username,
			"Email":       user.Email,
			"FirstName":   user.FirstName,
			"LastName":    user.LastName,
			"Role":        string(user.Role),
			"IsAdmin":     user.Role == models.RoleAdmin,
			"LoggedIn":    true,
			"CurrentYear": time.Now().Year(),
		})
	}
}
