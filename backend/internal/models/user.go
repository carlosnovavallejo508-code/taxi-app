package models

import (
    "time"
    "github.com/google/uuid"
)

type User struct {
    ID          uuid.UUID  `json:"id"`
    Phone       string     `json:"phone"`
    Email       *string    `json:"email,omitempty"`
    PasswordHash string    `json:"-"`
    FullName    string     `json:"full_name"`
    UserType    string     `json:"user_type"`
    AvatarURL   *string    `json:"avatar_url,omitempty"`
    Rating      float64    `json:"rating"`
    TotalRides  int        `json:"total_rides"`
    IsVerified  bool       `json:"is_verified"`
    IsActive    bool       `json:"is_active"`
    LastLogin   *time.Time `json:"last_login,omitempty"`
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`
    DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type Driver struct {
    ID              uuid.UUID  `json:"id"`
    UserID          uuid.UUID  `json:"user_id"`
    VehicleType     string     `json:"vehicle_type"`
    VehicleMake     string     `json:"vehicle_make"`
    VehicleModel    string     `json:"vehicle_model"`
    VehicleYear     int        `json:"vehicle_year"`
    VehicleColor    string     `json:"vehicle_color"`
    LicensePlate    string     `json:"license_plate"`
    LicenseNumber   string     `json:"license_number"`
    IsAvailable     bool       `json:"is_available"`
    IsOnline        bool       `json:"is_online"`
    CurrentLat      *float64   `json:"current_lat,omitempty"`
    CurrentLng      *float64   `json:"current_lng,omitempty"`
    CurrentHeading  *float64   `json:"current_heading,omitempty"`
    LastActive      *time.Time `json:"last_active,omitempty"`
    TotalEarnings   float64    `json:"total_earnings"`
    TotalTrips      int        `json:"total_trips"`
    AcceptanceRate  float64    `json:"acceptance_rate"`
    CancellationRate float64   `json:"cancellation_rate"`
    CreatedAt       time.Time  `json:"created_at"`
    UpdatedAt       time.Time  `json:"updated_at"`
}

type Business struct {
    ID             uuid.UUID `json:"id"`
    UserID         uuid.UUID `json:"user_id"`
    Name           string    `json:"name"`
    BusinessType   string    `json:"business_type"`
    TaxID          string    `json:"tax_id,omitempty"`
    Address        string    `json:"address,omitempty"`
    Lat            *float64  `json:"lat,omitempty"`
    Lng            *float64  `json:"lng,omitempty"`
    Phone          string    `json:"phone,omitempty"`
    Website        string    `json:"website,omitempty"`
    LogoURL        string    `json:"logo_url,omitempty"`
    IsPremium      bool      `json:"is_premium"`
    MonthlyRideLimit int     `json:"monthly_ride_limit"`
    CreatedAt      time.Time `json:"created_at"`
    UpdatedAt      time.Time `json:"updated_at"`
}
