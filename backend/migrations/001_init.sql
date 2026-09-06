-- Extensiones necesarias
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "postgis";

-- ENUMS
CREATE TYPE user_type AS ENUM ('passenger', 'driver', 'business', 'admin');
CREATE TYPE ride_status AS ENUM ('pending', 'accepted', 'arriving', 'in_progress', 'completed', 'cancelled');
CREATE TYPE payment_status AS ENUM ('pending', 'processing', 'completed', 'failed', 'refunded');
CREATE TYPE payment_method AS ENUM ('cash', 'card', 'wallet');
CREATE TYPE vehicle_type AS ENUM ('standard', 'comfort', 'premium', 'xl', 'electric');

-- Tabla de usuarios
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    phone VARCHAR(15) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100) NOT NULL,
    user_type user_type NOT NULL,
    avatar_url TEXT,
    rating DECIMAL(3,2) DEFAULT 5.00,
    total_rides INT DEFAULT 0,
    preferred_language VARCHAR(10) DEFAULT 'es',
    is_verified BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    last_login TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Tabla de conductores
CREATE TABLE drivers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vehicle_type vehicle_type NOT NULL,
    vehicle_make VARCHAR(50),
    vehicle_model VARCHAR(50),
    vehicle_year INT,
    vehicle_color VARCHAR(30),
    license_plate VARCHAR(20) NOT NULL UNIQUE,
    license_number VARCHAR(50) NOT NULL,
    license_expiry DATE,
    insurance_number VARCHAR(50),
    insurance_expiry DATE,
    is_available BOOLEAN DEFAULT true,
    is_online BOOLEAN DEFAULT false,
    current_lat DECIMAL(10,8),
    current_lng DECIMAL(11,8),
    current_heading DECIMAL(5,2),
    last_active TIMESTAMP WITH TIME ZONE,
    total_earnings DECIMAL(12,2) DEFAULT 0,
    total_trips INT DEFAULT 0,
    acceptance_rate DECIMAL(5,2) DEFAULT 100,
    cancellation_rate DECIMAL(5,2) DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de negocios
CREATE TABLE businesses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    business_type VARCHAR(50),
    tax_id VARCHAR(50),
    address TEXT,
    lat DECIMAL(10,8),
    lng DECIMAL(11,8),
    phone VARCHAR(20),
    website VARCHAR(255),
    logo_url TEXT,
    is_premium BOOLEAN DEFAULT false,
    monthly_ride_limit INT DEFAULT 100,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de viajes
CREATE TABLE rides (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    driver_id UUID REFERENCES drivers(id),
    business_id UUID REFERENCES businesses(id),
    pickup_lat DECIMAL(10,8) NOT NULL,
    pickup_lng DECIMAL(11,8) NOT NULL,
    pickup_address TEXT,
    dropoff_lat DECIMAL(10,8) NOT NULL,
    dropoff_lng DECIMAL(11,8) NOT NULL,
    dropoff_address TEXT,
    status ride_status DEFAULT 'pending',
    vehicle_type vehicle_type,
    price_estimate DECIMAL(10,2),
    price_final DECIMAL(10,2),
    distance_km DECIMAL(10,2),
    duration_min INT,
    payment_method payment_method,
    payment_status payment_status DEFAULT 'pending',
    driver_rating DECIMAL(3,2),
    passenger_rating DECIMAL(3,2),
    feedback TEXT,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    cancelled_at TIMESTAMP WITH TIME ZONE,
    cancelled_by UUID,
    cancel_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de pagos
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ride_id UUID NOT NULL REFERENCES rides(id),
    user_id UUID NOT NULL REFERENCES users(id),
    amount DECIMAL(10,2) NOT NULL,
    fee DECIMAL(10,2) DEFAULT 0,
    total DECIMAL(10,2) NOT NULL,
    method payment_method NOT NULL,
    status payment_status DEFAULT 'pending',
    transaction_id VARCHAR(100),
    refund_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de notificaciones
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    title VARCHAR(200) NOT NULL,
    body TEXT NOT NULL,
    type VARCHAR(50),
    data JSONB,
    is_read BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de documentos
CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    driver_id UUID REFERENCES drivers(id),
    document_type VARCHAR(50) NOT NULL,
    document_url TEXT NOT NULL,
    document_number VARCHAR(100),
    expiry_date DATE,
    is_verified BOOLEAN DEFAULT false,
    verified_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de reviews
CREATE TABLE reviews (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ride_id UUID NOT NULL REFERENCES rides(id),
    reviewer_id UUID NOT NULL REFERENCES users(id),
    reviewee_id UUID NOT NULL REFERENCES users(id),
    rating DECIMAL(3,2) NOT NULL,
    comment TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Índices optimizados
CREATE INDEX idx_users_phone ON users(phone);
CREATE INDEX idx_users_email ON users(email) WHERE email IS NOT NULL;
CREATE INDEX idx_users_type ON users(user_type);
CREATE INDEX idx_drivers_user ON drivers(user_id);
CREATE INDEX idx_drivers_available ON drivers(is_available, is_online) WHERE is_available = true AND is_online = true;
CREATE INDEX idx_drivers_location ON drivers(current_lat, current_lng);
CREATE INDEX idx_rides_user ON rides(user_id, status);
CREATE INDEX idx_rides_driver ON rides(driver_id, status);
CREATE INDEX idx_rides_status ON rides(status) WHERE status IN ('pending', 'accepted', 'arriving', 'in_progress');
CREATE INDEX idx_rides_created ON rides(created_at DESC);
CREATE INDEX idx_payments_ride ON payments(ride_id);
CREATE INDEX idx_payments_user ON payments(user_id);
CREATE INDEX idx_notifications_user ON notifications(user_id, is_read) WHERE is_read = false;
CREATE INDEX idx_documents_user ON documents(user_id);
CREATE INDEX idx_documents_driver ON documents(driver_id);
CREATE INDEX idx_reviews_ride ON reviews(ride_id);

-- Trigger para actualizar updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_drivers_updated_at BEFORE UPDATE ON drivers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_rides_updated_at BEFORE UPDATE ON rides
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_payments_updated_at BEFORE UPDATE ON payments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
