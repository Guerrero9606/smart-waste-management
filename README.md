# Smart Waste Management API

![Go](https://img.shields.io/badge/Go-1.24-00ADD8?style=for-the-badge&logo=go)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-316192?style=for-the-badge&logo=postgresql)
![Docker](https://img.shields.io/badge/Docker-20.10-2496ED?style=for-the-badge&logo=docker)
![Gin](https://img.shields.io/badge/Gin-v1.9-007F9F?style=for-the-badge)

API backend para el proyecto "Aplicación web para la gestión en tiempo real de contenedores de basuras de una Smart City". Este sistema está diseñado para recibir datos de sensores de nivel de llenado, almacenarlos, y proveer endpoints para la visualización y la optimización de rutas de recogida.

## Descripción del Proyecto

El objetivo de este proyecto es desarrollar una solución de software Open Source para la monitorización de residuos urbanos. La API centraliza la lógica para:

- **Ingesta de Datos**: Recibir lecturas de nivel de llenado de sensores (simulados) en los contenedores.
- **Gestión de Contenedores**: Proveer una interfaz CRUD para administrar los contenedores del sistema.
- **Visualización de Datos**: Exponer endpoints que permitan a un frontend visualizar el estado y la ubicación de los contenedores en un mapa.
- **Optimización de Rutas**: Calcular una ruta de recogida eficiente basada en los contenedores que han alcanzado un nivel de llenado crítico.

La arquitectura sigue los principios de **Arquitectura Limpia (Hexagonal)**, separando el dominio, la aplicación y la infraestructura para un sistema más mantenible, escalable y fácil de probar.

## Stack Tecnológico

| Componente         | Tecnología                                     | Justificación                                                                                                                                |
| ------------------ | ---------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| **Lenguaje Backend** | [Go](https://go.dev/) (v1.24)                  | Alto rendimiento, excelente manejo de la concurrencia para la ingesta de datos, tipado estático y compilación a un binario único.              |
| **Framework API**  | [Gin Gonic](https://gin-gonic.com/)            | Framework web minimalista, rápido y con un robusto sistema de enrutamiento y middlewares.                                                    |
| **Base de Datos**  | [PostgreSQL](https://www.postgresql.org/) (v15) | Sistema de BBDD relacional robusto y fiable.                                                                                                 |
| **Extensión de BBDD** | [PostGIS](https://postgis.net/)                | Provee capacidades geoespaciales avanzadas para almacenar ubicaciones y realizar consultas de proximidad eficientes.                            |
| **Contenerización** | [Docker](https://www.docker.com/) & [Docker Compose](https://docs.docker.com/compose/) | Para crear entornos de desarrollo y producción consistentes, portables y aislados.                                                      |
| **Documentación API** | [Swaggo](https://github.com/swaggo/swag)     | Genera automáticamente una documentación interactiva de la API (Swagger/OpenAPI) a partir de comentarios en el código.                        |
| **Live Reloading**  | [Air](https://github.com/air-verse/air)        | Herramienta de desarrollo que recompila y reinicia la aplicación automáticamente al detectar cambios en el código.                              |

## Estructura del Proyecto
.
├── cmd/api/ # Punto de entrada de la aplicación
├── docs/ # Documentación de Swagger (auto-generada)
├── internal/
│ ├── container/ # Lógica del módulo 'container' (handler, service, repository)
│ ├── domain/ # Entidades y lógica de negocio pura
│ └── platform/ # Adaptadores de infraestructura (ej. conexión a BBDD)
├── simulator/ # Script Python para simular los sensores IoT
├── sql/ # Scripts de inicialización de la BBDD
├── .air.toml # Configuración para la herramienta Air
├── .dockerignore # Ficheros a ignorar por Docker
├── .env.example # Plantilla para variables de entorno
├── Dockerfile # Define cómo construir la imagen Docker de producción
├── docker-compose.yml # Orquesta los servicios para desarrollo local
└── go.mod # Fichero de dependencias de Go


## Guía de Inicio Rápido (Desarrollo Local)

Sigue estos pasos para levantar el entorno de desarrollo en tu máquina local.

### Prerrequisitos

- [Go](https://go.dev/doc/install) (v1.24 o superior)
- [Docker](https://docs.docker.com/get-docker/) & [Docker Compose](https://docs.docker.com/compose/install/)
- [Python](https://www.python.org/downloads/) (v3.8 o superior) para el simulador
- Herramienta [Air](https://github.com/air-verse/air): `go install github.com/air-verse/air@latest`

### Configuración

1.  **Clona el repositorio:**
    ```bash
    git clone <url-del-repositorio>
    cd smart-waste-management
    ```

2.  **Crea tu fichero de entorno local:**
    Copia la plantilla y ajústala si es necesario. Para desarrollo local, `DB_HOST` debe ser `localhost`.
    ```bash
    cp .env.example .env.local
    # Asegúrate de que DB_HOST=localhost en .env.local
    ```

3.  **Configura el simulador de Python:**
    ```bash
    cd simulator
    python3 -m venv .venv
    source .venv/bin/activate
    pip install -r requirements.txt
    cd ..
    ```

### Ejecución

1.  **Levanta la base de datos en Docker:**
    ```bash
    docker-compose up -d db
    ```

2.  **Inicia la API Go con recarga en caliente:**
    En una nueva terminal, desde la raíz del proyecto:
    ```bash
    air -c .air.toml -d
    ```

3.  **Inicia el simulador de sensores:**
    En otra terminal:
    ```bash
    cd simulator
    source .venv/bin/activate
    python main.py
    ```

¡Listo! La API estará corriendo en `http://localhost:8080`.

- **API Endpoints**: `http://localhost:8080/api/v1/...`
- **Documentación Swagger**: `http://localhost:8080/swagger/index.html`

## Despliegue en Producción

El proyecto está preparado para ser desplegado usando contenedores Docker.

1.  **Construye la imagen de la API:**
    Desde la raíz del proyecto, ejecuta:
    ```bash
    # Para arquitectura amd64 (Intel/AMD)
    docker build -t smart-waste-api:1.0 .

    # Para arquitectura arm64 (ej. Raspberry Pi, Orange Pi)
    docker buildx build --platform linux/arm64 -t smart-waste-api-arm:1.0 --load .
    ```

2.  **Transfiere la imagen y los ficheros al servidor:**
    - La imagen generada (usando `docker save` o un registry).
    - El fichero `docker-compose.prod.yml`.
    - El directorio `sql/`.

3.  **Configura el `.env` en el servidor:**
    Crea un fichero `.env` en el servidor con las credenciales y configuraciones de producción. **¡No subas este fichero a Git!**

4.  **Inicia el stack en producción:**
    ```bash
    docker-compose -f docker-compose.prod.yml up -d
    ```

## Endpoints de la API

Para una lista completa y la posibilidad de probar los endpoints, por favor consulta la documentación interactiva de Swagger que se ejecuta junto con la aplicación.

- **URL de Swagger**: `http://<host-de-la-api>/swagger/index.html`

Principales recursos disponibles:
- `POST /api/v1/containers`: Crear un nuevo contenedor.
- `GET /api/v1/containers`: Obtener la lista de todos los contenedores.
- `GET /api/v1/containers/{id}`: Obtener un contenedor específico.
- `POST /api/v1/readings`: Enviar una nueva lectura de sensor.
- `POST /api/v1/routes`: Generar una ruta de recogida.
