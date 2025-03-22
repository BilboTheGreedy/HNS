package handlers

import (
	"strconv"

	"github.com/bilbothegreedy/HNS/internal/models"
	"github.com/bilbothegreedy/HNS/internal/repository"
	"github.com/bilbothegreedy/HNS/internal/web/helpers"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

// UserHandler handles user-related requests
type UserHandler struct {
	BaseHandler
	userRepo     repository.UserRepository
	hostnameRepo repository.HostnameRepository
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(
	userRepo repository.UserRepository,
	hostnameRepo repository.HostnameRepository,
) *UserHandler {
	return &UserHandler{
		BaseHandler:  BaseHandler{},
		userRepo:     userRepo,
		hostnameRepo: hostnameRepo,
	}
}

// ViewUser displays user details (admin only)
func (h *UserHandler) ViewUser(c *gin.Context) {
	// Check admin access
	isAdmin, exists := c.Get("isAdmin")
	if !exists || !isAdmin.(bool) {
		h.Forbidden(c)
		return
	}

	// Get user ID
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.RedirectWithAlert(c, "/admin/users", "danger", "Invalid user ID")
		return
	}

	// Get user
	ctx := c.Request.Context()
	user, err := h.userRepo.GetByID(ctx, id)
	if err != nil {
		h.RedirectWithAlert(c, "/admin/users", "danger", "User not found")
		return
	}

	// Get user's hostnames
	filters := map[string]interface{}{
		"reserved_by": user.Username,
	}
	hostnames, total, err := h.hostnameRepo.List(ctx, 10, 0, filters)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user's hostnames")
	}

	// Render template
	h.RenderTemplate(c, "admin_user_detail", gin.H{
		"Title":      "User Details",
		"ActivePage": "admin",
		"User":       user,
		"Hostnames":  hostnames,
		"Total":      total,
	})
}

// UserProfile displays the user profile page
func (h *UserHandler) UserProfile(c *gin.Context) {
	// Get user from context
	userInterface, exists := c.Get("user")
	if !exists {
		c.Redirect(302, "/login")
		return
	}

	user := userInterface.(*models.User)

	// Get user's recent hostnames
	ctx := c.Request.Context()
	filters := map[string]interface{}{
		"reserved_by": user.Username,
	}
	recentHostnames, _, err := h.hostnameRepo.List(ctx, 5, 0, filters)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user's recent hostnames")
	}

	// Get statistics
	reservedCount, _ := h.hostnameRepo.CountByUser(ctx, user.Username, models.StatusReserved)
	committedCount, _ := h.hostnameRepo.CountByUser(ctx, user.Username, models.StatusCommitted)
	releasedCount, _ := h.hostnameRepo.CountByUser(ctx, user.Username, models.StatusReleased)

	stats := gin.H{
		"Reserved":  reservedCount,
		"Committed": committedCount,
		"Released":  releasedCount,
		"Total":     reservedCount + committedCount + releasedCount,
	}

	// Format recent activities
	type Activity struct {
		ID        int64
		Name      string
		Action    string
		Status    string
		Timestamp interface{}
	}

	activities := make([]Activity, 0)
	for _, hostname := range recentHostnames {
		// Add reservation activity
		activities = append(activities, Activity{
			ID:        hostname.ID,
			Name:      hostname.Name,
			Action:    "Reserved",
			Status:    string(hostname.Status),
			Timestamp: hostname.ReservedAt,
		})

		// Add commit activity if committed
		if hostname.CommittedAt != nil {
			activities = append(activities, Activity{
				ID:        hostname.ID,
				Name:      hostname.Name,
				Action:    "Committed",
				Status:    string(hostname.Status),
				Timestamp: hostname.CommittedAt,
			})
		}

		// Add release activity if released
		if hostname.ReleasedAt != nil {
			activities = append(activities, Activity{
				ID:        hostname.ID,
				Name:      hostname.Name,
				Action:    "Released",
				Status:    string(hostname.Status),
				Timestamp: hostname.ReleasedAt,
			})
		}
	}

	// Render template
	h.RenderTemplate(c, "user_profile", gin.H{
		"Title":           "My Profile",
		"ActivePage":      "profile",
		"User":            user,
		"RecentHostnames": recentHostnames,
		"Activities":      activities,
		"Stats":           stats,
	})
}

// DeleteUser handles user deletion (admin only)
func (h *UserHandler) DeleteUser(c *gin.Context) {
	// Check admin access
	isAdmin, exists := c.Get("isAdmin")
	if !exists || !isAdmin.(bool) {
		h.Forbidden(c)
		return
	}

	// Get user ID
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.RedirectWithAlert(c, "/admin/users", "danger", "Invalid user ID")
		return
	}

	// Check if trying to delete self
	userID, _ := c.Get("userID")
	if userID.(int64) == id {
		h.RedirectWithAlert(c, "/admin/users", "danger", "Cannot delete your own account")
		return
	}

	// Delete user
	err = h.userRepo.Delete(c.Request.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete user")
		h.RedirectWithAlert(c, "/admin/users", "danger", "Failed to delete user")
		return
	}

	// Success
	h.RedirectWithAlert(c, "/admin/users", "success", "User deleted successfully")
}

// UpdateProfile handles profile update form submission
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	// Get user
	userID, exists := c.Get("userID")
	if !exists {
		c.Redirect(302, "/login")
		return
	}

	// Get form data
	firstName := c.PostForm("first_name")
	lastName := c.PostForm("last_name")
	email := c.PostForm("email")

	// Get user from database
	ctx := c.Request.Context()
	user, err := h.userRepo.GetByID(ctx, userID.(int64))
	if err != nil {
		h.RedirectWithAlert(c, "/profile", "danger", "User not found")
		return
	}

	// Update user fields
	user.FirstName = firstName
	user.LastName = lastName
	user.Email = email

	// Save changes
	err = h.userRepo.Update(ctx, user)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update user profile")
		h.RedirectWithAlert(c, "/profile", "danger", "Failed to update profile")
		return
	}

	// Success
	h.RedirectWithAlert(c, "/profile", "success", "Profile updated successfully")
}

// ChangePassword handles password change form submission
func (h *UserHandler) ChangePassword(c *gin.Context) {
	// Get user
	userID, exists := c.Get("userID")
	if !exists {
		c.Redirect(302, "/login")
		return
	}

	// Get form data
	currentPassword := c.PostForm("current_password")
	newPassword := c.PostForm("new_password")
	confirmPassword := c.PostForm("confirm_password")

	// Validate passwords match
	if newPassword != confirmPassword {
		h.RedirectWithAlert(c, "/profile", "danger", "New passwords do not match")
		return
	}

	// Get user from database
	ctx := c.Request.Context()
	user, err := h.userRepo.GetByID(ctx, userID.(int64))
	if err != nil {
		h.RedirectWithAlert(c, "/profile", "danger", "User not found")
		return
	}

	// Verify current password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword))
	if err != nil {
		h.RedirectWithAlert(c, "/profile", "danger", "Current password is incorrect")
		return
	}

	// Update password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password")
		h.RedirectWithAlert(c, "/profile", "danger", "Failed to update password")
		return
	}

	user.PasswordHash = string(hashedPassword)

	// Save changes
	err = h.userRepo.Update(ctx, user)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update password")
		h.RedirectWithAlert(c, "/profile", "danger", "Failed to update password")
		return
	}

	// Success
	h.RedirectWithAlert(c, "/profile", "success", "Password updated successfully")
}

// UsersList displays the list of users (admin only)
func (h *UserHandler) UsersList(c *gin.Context) {
	// Check admin access
	isAdmin, exists := c.Get("isAdmin")
	if !exists || !isAdmin.(bool) {
		h.Forbidden(c)
		return
	}

	// Get pagination parameters
	limit, offset := helpers.GetPaginationParams(c)

	// Get users
	ctx := c.Request.Context()
	users, total, err := h.userRepo.List(ctx, limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get users")
		h.RenderTemplate(c, "admin_users", gin.H{
			"Title": "User Management",
			"Users": []*models.User{},
		})
		return
	}

	// Get current username
	username, _ := c.Get("username")
	if username == nil {
		username = ""
	}

	// Combine pagination data with template data
	templateData := gin.H{
		"Title":       "User Management",
		"ActivePage":  "admin",
		"Users":       users,
		"CurrentUser": username.(string),
	}

	// Add pagination data
	paginationData := helpers.GetPaginationData(total, limit, offset)
	for k, v := range paginationData {
		templateData[k] = v
	}

	h.RenderTemplate(c, "admin_users", templateData)
}

// CreateUser handles user creation form submission (admin only)
func (h *UserHandler) CreateUser(c *gin.Context) {
	// Check admin access
	isAdmin, exists := c.Get("isAdmin")
	if !exists || !isAdmin.(bool) {
		h.Forbidden(c)
		return
	}

	// Get form data
	username := c.PostForm("username")
	email := c.PostForm("email")
	password := c.PostForm("password")
	firstName := c.PostForm("first_name")
	lastName := c.PostForm("last_name")
	role := c.PostForm("role")
	isActive := c.PostForm("is_active") == "on"

	// Create user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password")
		h.RedirectWithAlert(c, "/admin/users", "danger", "Failed to create user: Error hashing password")
		return
	}

	user := &models.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hashedPassword),
		FirstName:    firstName,
		LastName:     lastName,
		Role:         models.Role(role),
		IsActive:     isActive,
	}

	// Save user
	err = h.userRepo.Create(c.Request.Context(), user)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create user")
		h.RedirectWithAlert(c, "/admin/users", "danger", "Failed to create user: "+err.Error())
		return
	}

	// Success
	h.RedirectWithAlert(c, "/admin/users", "success", "User created successfully")
}

// UpdateUser handles user update form submission (admin only)
func (h *UserHandler) UpdateUser(c *gin.Context) {
	// Check admin access
	isAdmin, exists := c.Get("isAdmin")
	if !exists || !isAdmin.(bool) {
		h.Forbidden(c)
		return
	}

	// Get user ID
	idStr := c.PostForm("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.RedirectWithAlert(c, "/admin/users", "danger", "Invalid user ID")
		return
	}

	// Get user from database
	ctx := c.Request.Context()
	user, err := h.userRepo.GetByID(ctx, id)
	if err != nil {
		h.RedirectWithAlert(c, "/admin/users", "danger", "User not found")
		return
	}

	// Get form data
	email := c.PostForm("email")
	firstName := c.PostForm("first_name")
	lastName := c.PostForm("last_name")
	role := c.PostForm("role")
	password := c.PostForm("password")
	isActive := c.PostForm("is_active") == "on"

	// Update user fields
	user.Email = email
	user.FirstName = firstName
	user.LastName = lastName
	user.Role = models.Role(role)
	user.IsActive = isActive

	// Update password if provided
	if password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Error().Err(err).Msg("Failed to hash password")
			h.RedirectWithAlert(c, "/admin/users", "danger", "Failed to update password")
			return
		}
		user.PasswordHash = string(hashedPassword)
	}

	// Save changes
	err = h.userRepo.Update(ctx, user)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update user")
		h.RedirectWithAlert(c, "/admin/users", "danger", "Failed to update user")
		return
	}

	// Success
	h.RedirectWithAlert(c, "/admin/users", "success", "User updated successfully")
}

// EditUser displays the user edit form (admin only)
func (h *UserHandler) EditUser(c *gin.Context) {
	// Check admin access
	isAdmin, exists := c.Get("isAdmin")
	if !exists || !isAdmin.(bool) {
		h.Forbidden(c)
		return
	}

	// Get user ID
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.RedirectWithAlert(c, "/admin/users", "danger", "Invalid user ID")
		return
	}

	// Get user
	ctx := c.Request.Context()
	user, err := h.userRepo.GetByID(ctx, id)
	if err != nil {
		h.RedirectWithAlert(c, "/admin/users", "danger", "User not found")
		return
	}

	// Define available roles directly
	roles := []string{"admin", "user"}

	// Render template
	h.RenderTemplate(c, "admin_user_edit", gin.H{
		"Title":      "Edit User",
		"ActivePage": "admin",
		"User":       user,
		"Roles":      roles,
	})
}
