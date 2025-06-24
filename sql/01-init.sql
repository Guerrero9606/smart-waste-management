-- sql/01-init.sql

-- Establecemos la zona horaria a UTC para consistencia en todas las operaciones de tiempo.
SET TIME ZONE 'UTC';

-- Habilitamos las extensiones necesarias.
-- 'pgcrypto' para la generación de UUIDs.
CREATE EXTENSION IF NOT EXISTS pgcrypto;
-- 'postgis' para el soporte de datos geoespaciales.
CREATE EXTENSION IF NOT EXISTS postgis;

-- Creamos un tipo ENUM para el estado del contenedor.
-- Esto asegura la integridad de los datos, solo se permiten estos tres valores.
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'container_status') THEN
        CREATE TYPE container_status AS ENUM ('low', 'medium', 'high');
    END IF;
END$$;


-- Creamos la tabla 'containers' que almacenará la información estática de cada contenedor.
CREATE TABLE IF NOT EXISTS containers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- GEOGRAPHY es mejor que GEOMETRY para coordenadas lat/lon, ya que los cálculos (distancia, etc.) son más precisos.
    -- SRID 4326 es el estándar para WGS 84 (GPS).
    location GEOGRAPHY(POINT, 4326) NOT NULL,
    capacity_liters INT NOT NULL CHECK (capacity_liters > 0),

    -- Campos denormalizados para un acceso rápido al estado actual sin tener que consultar la tabla de lecturas.
    current_status container_status NOT NULL DEFAULT 'low',
    last_fill_level INT NOT NULL DEFAULT 0 CHECK (last_fill_level >= 0 AND last_fill_level <= 100),
    last_updated_at TIMESTAMPTZ, -- Timestamp con zona horaria de la última actualización de estado.

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Creamos un índice geoespacial GIST para acelerar las consultas de ubicación (ej. "contenedores en el área visible del mapa").
CREATE INDEX IF NOT EXISTS containers_location_idx ON containers USING GIST (location);
-- Un índice en el estado actual puede ser útil para filtrar rápidamente los contenedores llenos.
CREATE INDEX IF NOT EXISTS containers_current_status_idx ON containers (current_status);


-- Creamos la tabla 'readings' para almacenar el historial de lecturas de los sensores.
CREATE TABLE IF NOT EXISTS readings (
    id BIGSERIAL PRIMARY KEY,
    -- Referencia al contenedor. Si un contenedor se elimina, sus lecturas también se eliminan (ON DELETE CASCADE).
    container_id UUID NOT NULL REFERENCES containers(id) ON DELETE CASCADE,
    fill_level INT NOT NULL CHECK (fill_level >= 0 AND fill_level <= 100),
    recorded_at TIMESTAMPTZ NOT NULL
);

-- Creamos un índice compuesto para buscar eficientemente las lecturas de un contenedor específico, ordenadas por fecha.
CREATE INDEX IF NOT EXISTS readings_container_id_recorded_at_idx ON readings (container_id, recorded_at DESC);


-- === DATOS DE PRUEBA (SEED DATA) ===
-- Insertamos algunos contenedores de ejemplo para tener datos desde el principio.
-- Esto es increíblemente útil para el desarrollo del frontend y del backend.
DO $$
BEGIN
    -- Solo inserta si la tabla está vacía
    IF NOT EXISTS (SELECT 1 FROM containers) THEN
        INSERT INTO containers (id, location, capacity_liters, current_status, last_fill_level, last_updated_at) VALUES
        -- Contenedor lleno cerca de Sol, Madrid
        ('c7a1c7d6-3d2c-4e8d-9a6a-0b1e4c7b8e1a', ST_SetSRID(ST_MakePoint(-3.703790, 40.416775), 4326), 2400, 'high', 95, NOW() - INTERVAL '5 minutes'),
        -- Contenedor a nivel medio cerca del Palacio Real
        ('b8b2d8e7-4e3d-5f9e-a0b0-1c2f5d8e9f2b', ST_SetSRID(ST_MakePoint(-3.714141, 40.417953), 4326), 1100, 'medium', 60, NOW() - INTERVAL '1 hour'),
        -- Contenedor casi vacío cerca del Retiro
        ('a9c3e9f8-5f4e-6a0f-b1c1-2d3a6e9f0a3c', ST_SetSRID(ST_MakePoint(-3.684439, 40.414436), 4326), 2400, 'low', 15, NOW() - INTERVAL '2 hours');

        -- Insertamos algunas lecturas históricas para el primer contenedor
        INSERT INTO readings (container_id, fill_level, recorded_at) VALUES
        ('c7a1c7d6-3d2c-4e8d-9a6a-0b1e4c7b8e1a', 95, NOW() - INTERVAL '5 minutes'),
        ('c7a1c7d6-3d2c-4e8d-9a6a-0b1e4c7b8e1a', 80, NOW() - INTERVAL '3 hours'),
        ('c7a1c7d6-3d2c-4e8d-9a6a-0b1e4c7b8e1a', 50, NOW() - INTERVAL '1 day');
    END IF;
END$$;