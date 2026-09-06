package handlers

import (
    "context"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/jackc/pgx/v5/pgtype"

    "taxi-app/internal/database"
    "taxi-app/pkg/response"
)

type AdminHandler struct {
    db *database.Database
}

func NewAdminHandler(db *database.Database) *AdminHandler {
    return &AdminHandler{db: db}
}

// GetStats godoc
// @Summary Get system stats
// @Description Get general system statistics
// @Tags Admin
// @Produce json
// @Success 200 {object} object
// @Router /api/v1/admin/stats [get]
func (h *AdminHandler) GetStats(c *gin.Context) {
    ctx := context.Background()
    stats := gin.H{}

    // Total usuarios
    var totalUsers int64
    if err := h.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`).Scan(&totalUsers); err == nil {
        stats["total_users"] = totalUsers
    }

    // Total conductores
    var totalDrivers int64
    if err := h.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM drivers`).Scan(&totalDrivers); err == nil {
        stats["total_drivers"] = totalDrivers
    }

    // Total viajes
    var totalRides int64
    if err := h.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM rides`).Scan(&totalRides); err == nil {
        stats["total_rides"] = totalRides
    }

    // Viajes de hoy
    var todayRides int64
    if err := h.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM rides WHERE created_at >= CURRENT_DATE`).Scan(&todayRides); err == nil {
        stats["today_rides"] = todayRides
    }

    // Ingresos totales
    var totalRevenue pgtype.Numeric
    if err := h.db.Pool.QueryRow(ctx, `SELECT COALESCE(SUM(total), 0) FROM payments WHERE status = 'completed'`).Scan(&totalRevenue); err == nil {
        if totalRevenue.Valid {
            if val, err := totalRevenue.Float64Value(); err == nil {
                stats["total_revenue"] = val.Float64
            }
        } else {
            stats["total_revenue"] = 0
        }
    }

    // Usuarios activos (últimas 24h)
    var activeUsers int64
    if err := h.db.Pool.QueryRow(ctx, `SELECT COUNT(DISTINCT user_id) FROM rides WHERE created_at >= NOW() - INTERVAL '24 hours'`).Scan(&activeUsers); err == nil {
        stats["active_users_24h"] = activeUsers
    }

    response.Success(c, http.StatusOK, stats)
}

// GetUsers godoc
// @Summary Get all users
// @Description Get list of all users
// @Tags Admin
// @Produce json
// @Success 200 {array} object
// @Router /api/v1/admin/users [get]
func (h *AdminHandler) GetUsers(c *gin.Context) {
    query := `
        SELECT id::text, phone, email, full_name, user_type, rating::float8, 
               total_rides, is_verified, is_active, created_at
        FROM users
        WHERE deleted_at IS NULL
        ORDER BY created_at DESC
        LIMIT 50
    `
    rows, err := h.db.Pool.Query(context.Background(), query)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "Error al obtener usuarios", "DB_ERROR")
        return
    }
    defer rows.Close()

    users := []gin.H{}
    for rows.Next() {
        var (
            id         string
            phone      string
            email      *string
            fullName   string
            userType   string
            rating     float64
            totalRides int
            isVerified bool
            isActive   bool
            createdAt  time.Time
        )
        if err := rows.Scan(&id, &phone, &email, &fullName, &userType,
            &rating, &totalRides, &isVerified, &isActive, &createdAt); err != nil {
            continue
        }
        users = append(users, gin.H{
            "id":          id,
            "phone":       phone,
            "email":       email,
            "full_name":   fullName,
            "user_type":   userType,
            "rating":      rating,
            "total_rides": totalRides,
            "is_verified": isVerified,
            "is_active":   isActive,
            "created_at":  createdAt,
        })
    }

    response.Success(c, http.StatusOK, users)
}

// GetDrivers godoc
// @Summary Get all drivers
// @Description Get list of all drivers
// @Tags Admin
// @Produce json
// @Success 200 {array} object
// @Router /api/v1/admin/drivers [get]
func (h *AdminHandler) GetDrivers(c *gin.Context) {
    query := `
        SELECT d.id::text, u.full_name, u.phone, d.vehicle_type::text, 
               d.vehicle_model, d.license_plate, d.is_available, d.is_online,
               d.current_lat::float8, d.current_lng::float8, 
               d.total_earnings::float8, d.total_trips, 
               d.acceptance_rate::float8, d.cancellation_rate::float8, 
               d.created_at
        FROM drivers d
        JOIN users u ON d.user_id = u.id
        ORDER BY d.created_at DESC
        LIMIT 100
    `
    rows, err := h.db.Pool.Query(context.Background(), query)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "Error al obtener conductores", "DB_ERROR")
        return
    }
    defer rows.Close()

    drivers := []gin.H{}
    for rows.Next() {
        var (
            id                string
            fullName          string
            phone             string
            vehicleType       string
            vehicleModel      *string
            licensePlate      string
            isAvailable       bool
            isOnline          bool
            currentLat        *float64
            currentLng        *float64
            totalEarnings     float64
            totalTrips        int
            acceptanceRate    float64
            cancellationRate  float64
            createdAt         time.Time
        )
        if err := rows.Scan(&id, &fullName, &phone, &vehicleType,
            &vehicleModel, &licensePlate, &isAvailable, &isOnline,
            &currentLat, &currentLng, &totalEarnings, &totalTrips,
            &acceptanceRate, &cancellationRate, &createdAt); err != nil {
            continue
        }
        drivers = append(drivers, gin.H{
            "id":                id,
            "full_name":         fullName,
            "phone":             phone,
            "vehicle_type":      vehicleType,
            "vehicle_model":     vehicleModel,
            "license_plate":     licensePlate,
            "is_available":      isAvailable,
            "is_online":         isOnline,
            "current_lat":       currentLat,
            "current_lng":       currentLng,
            "total_earnings":    totalEarnings,
            "total_trips":       totalTrips,
            "acceptance_rate":   acceptanceRate,
            "cancellation_rate": cancellationRate,
            "created_at":        createdAt,
        })
    }

    response.Success(c, http.StatusOK, drivers)
}

// GetRides godoc
// @Summary Get all rides
// @Description Get list of all rides
// @Tags Admin
// @Produce json
// @Success 200 {array} object
// @Router /api/v1/admin/rides [get]
func (h *AdminHandler) GetRides(c *gin.Context) {
    query := `
        SELECT r.id::text, r.user_id::text, r.driver_id::text,
               r.pickup_lat::float8, r.pickup_lng::float8,
               r.dropoff_lat::float8, r.dropoff_lng::float8,
               r.status::text, r.vehicle_type::text,
               r.price_estimate::float8, r.price_final::float8,
               r.distance_km::float8, r.duration_min,
               r.payment_method::text, r.payment_status::text,
               r.created_at, r.completed_at
        FROM rides r
        ORDER BY r.created_at DESC
        LIMIT 100
    `
    rows, err := h.db.Pool.Query(context.Background(), query)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "Error al obtener viajes", "DB_ERROR")
        return
    }
    defer rows.Close()

    rides := []gin.H{}
    for rows.Next() {
        var (
            id            string
            userID        string
            driverID      *string
            pickupLat     float64
            pickupLng     float64
            dropoffLat    float64
            dropoffLng    float64
            status        string
            vehicleType   string
            priceEstimate *float64
            priceFinal    *float64
            distanceKm    *float64
            durationMin   *int
            paymentMethod *string
            paymentStatus *string
            createdAt     time.Time
            completedAt   *time.Time
        )
        if err := rows.Scan(&id, &userID, &driverID,
            &pickupLat, &pickupLng, &dropoffLat, &dropoffLng,
            &status, &vehicleType, &priceEstimate, &priceFinal,
            &distanceKm, &durationMin, &paymentMethod, &paymentStatus,
            &createdAt, &completedAt); err != nil {
            continue
        }
        rides = append(rides, gin.H{
            "id":             id,
            "user_id":        userID,
            "driver_id":      driverID,
            "pickup_lat":     pickupLat,
            "pickup_lng":     pickupLng,
            "dropoff_lat":    dropoffLat,
            "dropoff_lng":    dropoffLng,
            "status":         status,
            "vehicle_type":   vehicleType,
            "price_estimate": priceEstimate,
            "price_final":    priceFinal,
            "distance_km":    distanceKm,
            "duration_min":   durationMin,
            "payment_method": paymentMethod,
            "payment_status": paymentStatus,
            "created_at":     createdAt,
            "completed_at":   completedAt,
        })
    }

    response.Success(c, http.StatusOK, rides)
}
