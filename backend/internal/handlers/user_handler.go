package handlers

import (
    "context"
    "net/http"
	"strconv"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"

    "taxi-app/internal/database"
    "taxi-app/internal/models"
    "taxi-app/pkg/response"
)

type UserHandler struct {
    db *database.Database
}

func NewUserHandler(db *database.Database) *UserHandler {
    return &UserHandler{db: db}
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get current user profile
// @Tags Users
// @Produce json
// @Success 200 {object} models.User
// @Router /api/v1/users/me [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
    userID, _ := c.Get("user_id")
    uid := userID.(uuid.UUID)

    var user models.User
    query := `
        SELECT id, phone, email, full_name, user_type, avatar_url, 
               rating, total_rides, is_verified, is_active, created_at, updated_at
        FROM users 
        WHERE id = $1 AND deleted_at IS NULL
    `
    err := h.db.Pool.QueryRow(context.Background(), query, uid).Scan(
        &user.ID, &user.Phone, &user.Email, &user.FullName, &user.UserType,
        &user.AvatarURL, &user.Rating, &user.TotalRides, &user.IsVerified,
        &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
    )
    if err != nil {
        response.Error(c, http.StatusNotFound, "Usuario no encontrado", "USER_NOT_FOUND")
        return
    }

    response.Success(c, http.StatusOK, user)
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update current user profile
// @Tags Users
// @Accept json
// @Produce json
// @Param request body object true "User data to update"
// @Success 200 {object} models.User
// @Router /api/v1/users/me [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
    userID, _ := c.Get("user_id")
    uid := userID.(uuid.UUID)

    var req struct {
        FullName  string `json:"full_name"`
        Email     string `json:"email"`
        AvatarURL string `json:"avatar_url"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "Datos inválidos", "INVALID_REQUEST")
        return
    }

    // Construir query dinámica
    updates := []string{}
    args := []interface{}{}
    argCount := 1

    if req.FullName != "" {
        updates = append(updates, "full_name = $"+itoa(argCount))
        args = append(args, req.FullName)
        argCount++
    }
    if req.Email != "" {
        updates = append(updates, "email = $"+itoa(argCount))
        args = append(args, req.Email)
        argCount++
    }
    if req.AvatarURL != "" {
        updates = append(updates, "avatar_url = $"+itoa(argCount))
        args = append(args, req.AvatarURL)
        argCount++
    }

    if len(updates) == 0 {
        response.Error(c, http.StatusBadRequest, "No hay datos para actualizar", "NO_DATA")
        return
    }

    // Agregar updated_at
    updates = append(updates, "updated_at = $"+itoa(argCount))
    args = append(args, time.Now())
    argCount++

    // Agregar WHERE
    args = append(args, uid)

    query := "UPDATE users SET " + join(updates, ", ") + " WHERE id = $" + itoa(argCount) + 
        " RETURNING id, phone, email, full_name, user_type, avatar_url, rating, total_rides, is_verified, is_active, created_at, updated_at"

    var user models.User
    err := h.db.Pool.QueryRow(context.Background(), query, args...).Scan(
        &user.ID, &user.Phone, &user.Email, &user.FullName, &user.UserType,
        &user.AvatarURL, &user.Rating, &user.TotalRides, &user.IsVerified,
        &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
    )
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "Error al actualizar", "UPDATE_ERROR")
        return
    }

    response.Success(c, http.StatusOK, user)
}

// Funciones auxiliares
func itoa(i int) string {
    return strconv.Itoa(i)
}

func join(strs []string, sep string) string {
    result := ""
    for i, s := range strs {
        if i > 0 {
            result += sep
        }
        result += s
    }
    return result
}
