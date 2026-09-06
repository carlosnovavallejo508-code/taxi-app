package handlers

import (
    "context"
    "net/http"
	"strconv"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"

    "taxi-app/internal/database"
    "taxi-app/internal/models"
    "taxi-app/pkg/response"
)

type NotificationHandler struct {
    db *database.Database
}

func NewNotificationHandler(db *database.Database) *NotificationHandler {
    return &NotificationHandler{db: db}
}

// GetNotifications godoc
// @Summary Get user notifications
// @Description Get all notifications for current user
// @Tags Notifications
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} models.Notification
// @Router /api/v1/notifications [get]
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
    userID, _ := c.Get("user_id")
    uid := userID.(uuid.UUID)

    limit := 20
    offset := 0
    if l := c.Query("limit"); l != "" {
        if val, err := strconv.Atoi(l); err == nil && val > 0 {
            limit = val
        }
    }
    if o := c.Query("offset"); o != "" {
        if val, err := strconv.Atoi(o); err == nil && val >= 0 {
            offset = val
        }
    }

    query := `
        SELECT id, user_id, title, body, type, data, is_read, created_at
        FROM notifications
        WHERE user_id = $1
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3
    `
    rows, err := h.db.Pool.Query(context.Background(), query, uid, limit, offset)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "Error al obtener notificaciones", "DB_ERROR")
        return
    }
    defer rows.Close()

    notifications := []models.Notification{}
    for rows.Next() {
        var notif models.Notification
        if err := rows.Scan(
            &notif.ID, &notif.UserID, &notif.Title, &notif.Body,
            &notif.Type, &notif.Data, &notif.IsRead, &notif.CreatedAt,
        ); err != nil {
            continue
        }
        notifications = append(notifications, notif)
    }

    response.Success(c, http.StatusOK, notifications)
}

// MarkAsRead godoc
// @Summary Mark notification as read
// @Description Mark a notification as read
// @Tags Notifications
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} response.Response
// @Router /api/v1/notifications/{id}/read [put]
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
    notificationID := c.Param("id")
    nid, _ := uuid.Parse(notificationID)
    if nid == uuid.Nil {
        response.Error(c, http.StatusBadRequest, "ID inválido", "INVALID_ID")
        return
    }

    userID, _ := c.Get("user_id")
    uid := userID.(uuid.UUID)

    query := `UPDATE notifications SET is_read = true WHERE id = $1 AND user_id = $2`
    result, err := h.db.Pool.Exec(context.Background(), query, nid, uid)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "Error al actualizar notificación", "UPDATE_ERROR")
        return
    }

    if result.RowsAffected() == 0 {
        response.Error(c, http.StatusNotFound, "Notificación no encontrada", "NOT_FOUND")
        return
    }

    response.Success(c, http.StatusOK, gin.H{
        "message": "Notificación marcada como leída",
    })
}
