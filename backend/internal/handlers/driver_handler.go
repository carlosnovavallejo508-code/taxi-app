package handlers

import (
    "context"
    "encoding/json"
    "fmt"
    "math"
    "net/http"
    "strconv"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"

    "taxi-app/internal/database"
    "taxi-app/pkg/response"
)

type DriverHandler struct {
    db    *database.Database
    cache *database.Cache
}

func NewDriverHandler(db *database.Database, cache *database.Cache) *DriverHandler {
    return &DriverHandler{db: db, cache: cache}
}

type NearbyDriver struct {
    ID          uuid.UUID `json:"id"`
    Name        string    `json:"name"`
    VehicleType string    `json:"vehicle_type"`
    VehicleModel string   `json:"vehicle_model,omitempty"`
    LicensePlate string   `json:"license_plate"`
    Rating      float64   `json:"rating"`
    DistanceKm  float64   `json:"distance_km"`
    ETA         int       `json:"eta_minutes"`
    CurrentLat  float64   `json:"current_lat"`
    CurrentLng  float64   `json:"current_lng"`
}

// GetNearbyDrivers godoc
// @Summary Get nearby drivers
// @Description Get available drivers near a location
// @Tags Drivers
// @Produce json
// @Param lat query number true "Latitude"
// @Param lng query number true "Longitude"
// @Param radius query number false "Radius in km" default(5)
// @Success 200 {array} NearbyDriver
// @Router /api/v1/drivers/nearby [get]
func (h *DriverHandler) GetNearbyDrivers(c *gin.Context) {
    lat, err := strconv.ParseFloat(c.Query("lat"), 64)
    if err != nil {
        response.Error(c, http.StatusBadRequest, "Latitud inválida", "INVALID_LAT")
        return
    }

    lng, err := strconv.ParseFloat(c.Query("lng"), 64)
    if err != nil {
        response.Error(c, http.StatusBadRequest, "Longitud inválida", "INVALID_LNG")
        return
    }

    radius, _ := strconv.ParseFloat(c.DefaultQuery("radius", "5"), 64)
    if radius <= 0 {
        radius = 5
    }

    // Intentar obtener de cache
    cacheKey := fmt.Sprintf("drivers:nearby:%.4f:%.4f:%.1f", lat, lng, radius)
    if h.cache != nil {
        if cached, err := h.cache.Client.Get(context.Background(), cacheKey).Result(); err == nil {
            var drivers []NearbyDriver
            if json.Unmarshal([]byte(cached), &drivers) == nil {
                response.Success(c, http.StatusOK, drivers)
                return
            }
        }
    }

    // Query para encontrar conductores cercanos
    query := `
        SELECT d.id, u.full_name, d.vehicle_type, d.vehicle_model, d.license_plate,
               u.rating, d.current_lat, d.current_lng,
               (6371 * acos(cos(radians($1)) * cos(radians(d.current_lat)) * 
                cos(radians(d.current_lng) - radians($2)) + 
                sin(radians($1)) * sin(radians(d.current_lat)))) AS distance
        FROM drivers d
        JOIN users u ON d.user_id = u.id
        WHERE d.is_available = true 
          AND d.is_online = true
          AND d.current_lat IS NOT NULL 
          AND d.current_lng IS NOT NULL
          AND (6371 * acos(cos(radians($1)) * cos(radians(d.current_lat)) * 
               cos(radians(d.current_lng) - radians($2)) + 
               sin(radians($1)) * sin(radians(d.current_lat)))) <= $3
        ORDER BY distance
        LIMIT 20
    `

    rows, err := h.db.Pool.Query(context.Background(), query, lat, lng, radius)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "Error al buscar conductores", "DB_ERROR")
        return
    }
    defer rows.Close()

    var drivers []NearbyDriver
    for rows.Next() {
        var driver NearbyDriver
        var distance float64
        if err := rows.Scan(
            &driver.ID, &driver.Name, &driver.VehicleType, &driver.VehicleModel,
            &driver.LicensePlate, &driver.Rating, &driver.CurrentLat, &driver.CurrentLng,
            &distance,
        ); err != nil {
            continue
        }
        driver.DistanceKm = math.Round(distance*100) / 100
        driver.ETA = int(distance / 30 * 60) // Asumiendo 30 km/h promedio
        if driver.ETA < 1 {
            driver.ETA = 1
        }
        drivers = append(drivers, driver)
    }

    // Guardar en cache por 30 segundos
    if h.cache != nil {
        if data, err := json.Marshal(drivers); err == nil {
            h.cache.Client.Set(context.Background(), cacheKey, data, 30*time.Second)
        }
    }

    response.Success(c, http.StatusOK, drivers)
}

// UpdateLocation godoc
// @Summary Update driver location
// @Description Update current driver location
// @Tags Drivers
// @Accept json
// @Produce json
// @Param request body object true "Location data"
// @Success 200 {object} response.Response
// @Router /api/v1/drivers/location [put]
func (h *DriverHandler) UpdateLocation(c *gin.Context) {
    userID, _ := c.Get("user_id")
    uid := userID.(uuid.UUID)

    var req struct {
        Lat     float64 `json:"lat" binding:"required"`
        Lng     float64 `json:"lng" binding:"required"`
        Heading float64 `json:"heading"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "Datos inválidos", "INVALID_REQUEST")
        return
    }

    // Validar coordenadas
    if req.Lat < -90 || req.Lat > 90 || req.Lng < -180 || req.Lng > 180 {
        response.Error(c, http.StatusBadRequest, "Coordenadas inválidas", "INVALID_COORDINATES")
        return
    }

    // Actualizar ubicación
    query := `
        UPDATE drivers 
        SET current_lat = $1, current_lng = $2, current_heading = $3, 
            last_active = $4, updated_at = $5
        WHERE user_id = $6
    `
    _, err := h.db.Pool.Exec(context.Background(), query,
        req.Lat, req.Lng, req.Heading, time.Now(), time.Now(), uid,
    )
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "Error al actualizar ubicación", "UPDATE_ERROR")
        return
    }

    response.Success(c, http.StatusOK, gin.H{
        "message": "Ubicación actualizada",
    })
}

// UpdateAvailability godoc
// @Summary Update driver availability
// @Description Update driver availability status
// @Tags Drivers
// @Accept json
// @Produce json
// @Param request body object true "Availability data"
// @Success 200 {object} response.Response
// @Router /api/v1/drivers/availability [put]
func (h *DriverHandler) UpdateAvailability(c *gin.Context) {
    userID, _ := c.Get("user_id")
    uid := userID.(uuid.UUID)

    var req struct {
        IsAvailable bool `json:"is_available"`
        IsOnline    bool `json:"is_online"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "Datos inválidos", "INVALID_REQUEST")
        return
    }

    query := `
        UPDATE drivers 
        SET is_available = $1, is_online = $2, 
            last_active = $3, updated_at = $4
        WHERE user_id = $5
    `
    _, err := h.db.Pool.Exec(context.Background(), query,
        req.IsAvailable, req.IsOnline, time.Now(), time.Now(), uid,
    )
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "Error al actualizar disponibilidad", "UPDATE_ERROR")
        return
    }

    response.Success(c, http.StatusOK, gin.H{
        "message": "Disponibilidad actualizada",
        "is_available": req.IsAvailable,
        "is_online": req.IsOnline,
    })
}
