package handlers

import (
    "context"
    "encoding/json"
    "math"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"

    "taxi-app/internal/database"
    "taxi-app/internal/models"
    "taxi-app/pkg/response"
)

type RideHandler struct {
    db    *database.Database
    cache *database.Cache
}

func NewRideHandler(db *database.Database, cache *database.Cache) *RideHandler {
    return &RideHandler{db: db, cache: cache}
}

type CreateRideRequest struct {
    PickupLat      float64 `json:"pickup_lat" binding:"required"`
    PickupLng      float64 `json:"pickup_lng" binding:"required"`
    PickupAddress  string  `json:"pickup_address"`
    DropoffLat     float64 `json:"dropoff_lat" binding:"required"`
    DropoffLng     float64 `json:"dropoff_lng" binding:"required"`
    DropoffAddress string  `json:"dropoff_address"`
    VehicleType    string  `json:"vehicle_type" binding:"required,oneof=standard comfort premium xl electric"`
    PaymentMethod  string  `json:"payment_method" binding:"required,oneof=cash card wallet"`
}

// CreateRide godoc
// @Summary Create new ride
// @Description Create a new ride request
// @Tags Rides
// @Accept json
// @Produce json
// @Param request body CreateRideRequest true "Ride data"
// @Success 201 {object} models.Ride
// @Router /api/v1/rides [post]
func (h *RideHandler) CreateRide(c *gin.Context) {
    userID, _ := c.Get("user_id")
    uid := userID.(uuid.UUID)

    var req CreateRideRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "Datos inválidos", "INVALID_REQUEST")
        return
    }

    // Validar coordenadas
    if !isValidCoordinates(req.PickupLat, req.PickupLng) || !isValidCoordinates(req.DropoffLat, req.DropoffLng) {
        response.Error(c, http.StatusBadRequest, "Coordenadas inválidas", "INVALID_COORDINATES")
        return
    }

    // Calcular distancia y precio estimado
    distance := calculateDistance(req.PickupLat, req.PickupLng, req.DropoffLat, req.DropoffLng)
    priceEstimate := calculatePrice(distance, req.VehicleType)

    // Crear ride
    ride := models.Ride{
        ID:             uuid.New(),
        UserID:         uid,
        PickupLat:      req.PickupLat,
        PickupLng:      req.PickupLng,
        PickupAddress:  req.PickupAddress,
        DropoffLat:     req.DropoffLat,
        DropoffLng:     req.DropoffLng,
        DropoffAddress: req.DropoffAddress,
        Status:         "pending",
        VehicleType:    req.VehicleType,
        PriceEstimate:  priceEstimate,
        DistanceKm:     distance,
        DurationMin:    int(distance / 30 * 60), // 30 km/h promedio
        PaymentMethod:  req.PaymentMethod,
        PaymentStatus:  "pending",
        CreatedAt:      time.Now(),
        UpdatedAt:      time.Now(),
    }

    // Insertar en base de datos
    query := `
        INSERT INTO rides (
            id, user_id, pickup_lat, pickup_lng, pickup_address,
            dropoff_lat, dropoff_lng, dropoff_address,
            status, vehicle_type, price_estimate, distance_km,
            duration_min, payment_method, payment_status,
            created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
    `
    _, err := h.db.Pool.Exec(context.Background(), query,
        ride.ID, ride.UserID, ride.PickupLat, ride.PickupLng, ride.PickupAddress,
        ride.DropoffLat, ride.DropoffLng, ride.DropoffAddress,
        ride.Status, ride.VehicleType, ride.PriceEstimate, ride.DistanceKm,
        ride.DurationMin, ride.PaymentMethod, ride.PaymentStatus,
        ride.CreatedAt, ride.UpdatedAt,
    )
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "Error al crear viaje", "DB_ERROR")
        return
    }

    // Publicar en Redis para que los conductores cercanos lo vean
    if h.cache != nil {
        rideJSON, _ := json.Marshal(ride)
        h.cache.Client.Publish(context.Background(), "ride_requests", rideJSON)
    }

    response.Created(c, ride)
}

// GetRide godoc
// @Summary Get ride by ID
// @Description Get ride details by ID
// @Tags Rides
// @Produce json
// @Param id path string true "Ride ID"
// @Success 200 {object} models.Ride
// @Router /api/v1/rides/{id} [get]
func (h *RideHandler) GetRide(c *gin.Context) {
    rideID := c.Param("id")
    uid, _ := uuid.Parse(rideID)
    if uid == uuid.Nil {
        response.Error(c, http.StatusBadRequest, "ID inválido", "INVALID_ID")
        return
    }

    userID, _ := c.Get("user_id")
    userType, _ := c.Get("user_type")

    var ride models.Ride
    query := `SELECT * FROM rides WHERE id = $1`
    err := h.db.Pool.QueryRow(context.Background(), query, uid).Scan(
        &ride.ID, &ride.UserID, &ride.DriverID, &ride.BusinessID,
        &ride.PickupLat, &ride.PickupLng, &ride.PickupAddress,
        &ride.DropoffLat, &ride.DropoffLng, &ride.DropoffAddress,
        &ride.Status, &ride.VehicleType, &ride.PriceEstimate, &ride.PriceFinal,
        &ride.DistanceKm, &ride.DurationMin, &ride.PaymentMethod, &ride.PaymentStatus,
        &ride.DriverRating, &ride.PassengerRating, &ride.Feedback,
        &ride.StartedAt, &ride.CompletedAt, &ride.CancelledAt,
        &ride.CancelledBy, &ride.CancelReason, &ride.CreatedAt, &ride.UpdatedAt,
    )
    if err != nil {
        response.Error(c, http.StatusNotFound, "Viaje no encontrado", "RIDE_NOT_FOUND")
        return
    }

    // Verificar permisos
    if userType != "admin" && ride.UserID != userID {
        if userType != "driver" || (ride.DriverID != nil && *ride.DriverID != userID) {
            response.Error(c, http.StatusForbidden, "No autorizado", "FORBIDDEN")
            return
        }
    }

    response.Success(c, http.StatusOK, ride)
}

// GetActiveRide godoc
// @Summary Get active ride
// @Description Get current active ride for user
// @Tags Rides
// @Produce json
// @Success 200 {object} models.Ride
// @Router /api/v1/rides/active [get]
func (h *RideHandler) GetActiveRide(c *gin.Context) {
    userID, _ := c.Get("user_id")
    uid := userID.(uuid.UUID)

    query := `
        SELECT * FROM rides 
        WHERE (user_id = $1 OR driver_id = $1)
          AND status IN ('pending', 'accepted', 'arriving', 'in_progress')
        ORDER BY created_at DESC
        LIMIT 1
    `
    var ride models.Ride
    err := h.db.Pool.QueryRow(context.Background(), query, uid).Scan(
        &ride.ID, &ride.UserID, &ride.DriverID, &ride.BusinessID,
        &ride.PickupLat, &ride.PickupLng, &ride.PickupAddress,
        &ride.DropoffLat, &ride.DropoffLng, &ride.DropoffAddress,
        &ride.Status, &ride.VehicleType, &ride.PriceEstimate, &ride.PriceFinal,
        &ride.DistanceKm, &ride.DurationMin, &ride.PaymentMethod, &ride.PaymentStatus,
        &ride.DriverRating, &ride.PassengerRating, &ride.Feedback,
        &ride.StartedAt, &ride.CompletedAt, &ride.CancelledAt,
        &ride.CancelledBy, &ride.CancelReason, &ride.CreatedAt, &ride.UpdatedAt,
    )
    if err != nil {
        response.Error(c, http.StatusNotFound, "No hay viajes activos", "NO_ACTIVE_RIDE")
        return
    }

    response.Success(c, http.StatusOK, ride)
}

// UpdateRideStatus godoc
// @Summary Update ride status
// @Description Update ride status (accepted, arriving, in_progress, completed)
// @Tags Rides
// @Accept json
// @Produce json
// @Param id path string true "Ride ID"
// @Param request body object true "Status data"
// @Success 200 {object} models.Ride
// @Router /api/v1/rides/{id}/status [put]
func (h *RideHandler) UpdateRideStatus(c *gin.Context) {
    rideID := c.Param("id")
    uid, _ := uuid.Parse(rideID)
    if uid == uuid.Nil {
        response.Error(c, http.StatusBadRequest, "ID inválido", "INVALID_ID")
        return
    }

    var req struct {
        Status string `json:"status" binding:"required,oneof=accepted arriving in_progress completed"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "Estado inválido", "INVALID_STATUS")
        return
    }

    // Para conductor, verificar que sea el asignado
    userType, _ := c.Get("user_type")
    userID, _ := c.Get("user_id")

    if userType == "driver" {
        var driverID uuid.UUID
        query := `SELECT driver_id FROM rides WHERE id = $1`
        err := h.db.Pool.QueryRow(context.Background(), query, uid).Scan(&driverID)
        if err != nil || driverID != userID {
            response.Error(c, http.StatusForbidden, "No autorizado", "FORBIDDEN")
            return
        }
    }

    // Actualizar estado
    var updateQuery string
    var args []interface{}

    switch req.Status {
    case "accepted":
        updateQuery = `UPDATE rides SET status = 'accepted', driver_id = $1, updated_at = $2 WHERE id = $3`
        args = append(args, userID, time.Now(), uid)
    case "arriving":
        updateQuery = `UPDATE rides SET status = 'arriving', updated_at = $1 WHERE id = $2`
        args = append(args, time.Now(), uid)
    case "in_progress":
        updateQuery = `UPDATE rides SET status = 'in_progress', started_at = $1, updated_at = $2 WHERE id = $3`
        args = append(args, time.Now(), time.Now(), uid)
    case "completed":
        updateQuery = `UPDATE rides SET status = 'completed', completed_at = $1, updated_at = $2 WHERE id = $3`
        args = append(args, time.Now(), time.Now(), uid)
    }

    _, err := h.db.Pool.Exec(context.Background(), updateQuery, args...)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "Error al actualizar estado", "UPDATE_ERROR")
        return
    }

    response.Success(c, http.StatusOK, gin.H{
        "message": "Estado actualizado",
        "status":  req.Status,
    })
}

// CancelRide godoc
// @Summary Cancel ride
// @Description Cancel a ride request
// @Tags Rides
// @Accept json
// @Produce json
// @Param id path string true "Ride ID"
// @Param request body object true "Cancel data"
// @Success 200 {object} response.Response
// @Router /api/v1/rides/{id}/cancel [put]
func (h *RideHandler) CancelRide(c *gin.Context) {
    rideID := c.Param("id")
    uid, _ := uuid.Parse(rideID)
    if uid == uuid.Nil {
        response.Error(c, http.StatusBadRequest, "ID inválido", "INVALID_ID")
        return
    }

    userID, _ := c.Get("user_id")

    var req struct {
        Reason string `json:"reason"`
    }
    c.ShouldBindJSON(&req)

    query := `
        UPDATE rides 
        SET status = 'cancelled', cancelled_at = $1, cancelled_by = $2, 
            cancel_reason = $3, updated_at = $4
        WHERE id = $5
    `
    _, err := h.db.Pool.Exec(context.Background(), query,
        time.Now(), userID, req.Reason, time.Now(), uid,
    )
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "Error al cancelar viaje", "CANCEL_ERROR")
        return
    }

    response.Success(c, http.StatusOK, gin.H{
        "message": "Viaje cancelado",
    })
}

// RateRide godoc
// @Summary Rate ride
// @Description Rate a completed ride
// @Tags Rides
// @Accept json
// @Produce json
// @Param id path string true "Ride ID"
// @Param request body object true "Rating data"
// @Success 200 {object} response.Response
// @Router /api/v1/rides/{id}/rate [post]
func (h *RideHandler) RateRide(c *gin.Context) {
    rideID := c.Param("id")
    uid, _ := uuid.Parse(rideID)
    if uid == uuid.Nil {
        response.Error(c, http.StatusBadRequest, "ID inválido", "INVALID_ID")
        return
    }

    var req struct {
        Rating  float64 `json:"rating" binding:"required,min=1,max=5"`
        Comment string  `json:"comment"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "Rating inválido", "INVALID_RATING")
        return
    }

    // Actualizar rating según quién califica
    userType, _ := c.Get("user_type")
    var query string
    if userType == "passenger" {
        query = `UPDATE rides SET driver_rating = $1, feedback = $2, updated_at = $3 WHERE id = $4`
    } else {
        query = `UPDATE rides SET passenger_rating = $1, feedback = $2, updated_at = $3 WHERE id = $4`
    }

    _, err := h.db.Pool.Exec(context.Background(), query,
        req.Rating, req.Comment, time.Now(), uid,
    )
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "Error al calificar", "RATE_ERROR")
        return
    }

    response.Success(c, http.StatusOK, gin.H{
        "message": "Calificación registrada",
    })
}

// Funciones auxiliares
func isValidCoordinates(lat, lng float64) bool {
    return lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180
}

func calculateDistance(lat1, lng1, lat2, lng2 float64) float64 {
    const earthRadius = 6371.0 // km

    lat1Rad := lat1 * math.Pi / 180
    lat2Rad := lat2 * math.Pi / 180
    dLat := (lat2 - lat1) * math.Pi / 180
    dLng := (lng2 - lng1) * math.Pi / 180

    a := math.Sin(dLat/2)*math.Sin(dLat/2) +
        math.Cos(lat1Rad)*math.Cos(lat2Rad)*
            math.Sin(dLng/2)*math.Sin(dLng/2)

    c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

    return math.Round(earthRadius*c*100) / 100
}

func calculatePrice(distanceKm float64, vehicleType string) float64 {
    // Tarifas base
    baseFare := map[string]float64{
        "standard": 2.50,
        "comfort":  4.00,
        "premium":  6.00,
        "xl":       5.00,
        "electric": 3.50,
    }

    // Tarifa por km
    perKm := map[string]float64{
        "standard": 1.20,
        "comfort":  1.80,
        "premium":  2.50,
        "xl":       2.00,
        "electric": 1.50,
    }

    base := baseFare[vehicleType]
    kmRate := perKm[vehicleType]

    if base == 0 {
        base = 2.50
    }
    if kmRate == 0 {
        kmRate = 1.20
    }

    total := base + (distanceKm * kmRate)
    return math.Round(total*100) / 100
}
