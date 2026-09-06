package models

import (
    "time"
    "github.com/google/uuid"
)

type Ride struct {
    ID              uuid.UUID  `json:"id"`
    UserID          uuid.UUID  `json:"user_id"`
    DriverID        *uuid.UUID `json:"driver_id,omitempty"`
    BusinessID      *uuid.UUID `json:"business_id,omitempty"`
    PickupLat       float64    `json:"pickup_lat"`
    PickupLng       float64    `json:"pickup_lng"`
    PickupAddress   string     `json:"pickup_address,omitempty"`
    DropoffLat      float64    `json:"dropoff_lat"`
    DropoffLng      float64    `json:"dropoff_lng"`
    DropoffAddress  string     `json:"dropoff_address,omitempty"`
    Status          string     `json:"status"`
    VehicleType     string     `json:"vehicle_type,omitempty"`
    PriceEstimate   float64    `json:"price_estimate,omitempty"`
    PriceFinal      float64    `json:"price_final,omitempty"`
    DistanceKm      float64    `json:"distance_km,omitempty"`
    DurationMin     int        `json:"duration_min,omitempty"`
    PaymentMethod   string     `json:"payment_method,omitempty"`
    PaymentStatus   string     `json:"payment_status,omitempty"`
    DriverRating    *float64   `json:"driver_rating,omitempty"`
    PassengerRating *float64   `json:"passenger_rating,omitempty"`
    Feedback        string     `json:"feedback,omitempty"`
    StartedAt       *time.Time `json:"started_at,omitempty"`
    CompletedAt     *time.Time `json:"completed_at,omitempty"`
    CancelledAt     *time.Time `json:"cancelled_at,omitempty"`
    CancelledBy     *uuid.UUID `json:"cancelled_by,omitempty"`
    CancelReason    string     `json:"cancel_reason,omitempty"`
    CreatedAt       time.Time  `json:"created_at"`
    UpdatedAt       time.Time  `json:"updated_at"`
}

type Payment struct {
    ID            uuid.UUID `json:"id"`
    RideID        uuid.UUID `json:"ride_id"`
    UserID        uuid.UUID `json:"user_id"`
    Amount        float64   `json:"amount"`
    Fee           float64   `json:"fee"`
    Total         float64   `json:"total"`
    Method        string    `json:"method"`
    Status        string    `json:"status"`
    TransactionID string    `json:"transaction_id,omitempty"`
    RefundReason  string    `json:"refund_reason,omitempty"`
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
}

type Notification struct {
    ID        uuid.UUID       `json:"id"`
    UserID    uuid.UUID       `json:"user_id"`
    Title     string          `json:"title"`
    Body      string          `json:"body"`
    Type      string          `json:"type,omitempty"`
    Data      interface{}     `json:"data,omitempty"`
    IsRead    bool            `json:"is_read"`
    CreatedAt time.Time       `json:"created_at"`
}

type Review struct {
    ID         uuid.UUID `json:"id"`
    RideID     uuid.UUID `json:"ride_id"`
    ReviewerID uuid.UUID `json:"reviewer_id"`
    RevieweeID uuid.UUID `json:"reviewee_id"`
    Rating     float64   `json:"rating"`
    Comment    string    `json:"comment,omitempty"`
    CreatedAt  time.Time `json:"created_at"`
}
