# Dockerfile

# --- Etapa 1: Builder ---
# Usamos una imagen oficial de Go con Alpine Linux, que es ligera.
# La nombramos 'builder' para poder referirnos a ella más tarde.
FROM golang:1.21-alpine AS builder

# Establecemos el directorio de trabajo dentro del contenedor.
WORKDIR /app

# Instalamos las herramientas necesarias para la compilación.
# 'git' es necesario para que 'go mod download' pueda clonar repositorios.
# 'ca-certificates' es necesario para realizar conexiones HTTPS seguras.
RUN apk add --no-cache git ca-certificates

# Copiamos primero los ficheros de módulos.
# Esta es una optimización clave de cacheado de Docker. Si estos ficheros no cambian,
# Docker reutilizará la capa cacheada del 'go mod download', haciendo las
# construcciones posteriores mucho más rápidas.
COPY go.mod go.sum ./
RUN go mod download

# Ahora, copia el resto del código fuente de la aplicación.
# El .dockerignore se asegurará de que no copiemos ficheros innecesarios.
COPY . .

# Generamos la documentación de Swagger. Este paso asegura que la documentación
# esté "horneada" dentro de la imagen final.
# La herramienta 'swag' debe estar disponible a través de 'go install'.
# Nos aseguramos de que el $GOPATH/bin está en el PATH del shell.
RUN go install github.com/air-verse/air@latest
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN /go/bin/swag init -g cmd/api/main.go

# Compilamos la aplicación Go.
# CGO_ENABLED=0: Crea un binario estático que no depende de librerías C del sistema. Crucial para distroless.
# -ldflags="-w -s": Optimización que reduce el tamaño del binario eliminando información de depuración.
# -o /app/main: Especifica que el binario de salida se llamará 'main' y estará en /app.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/main ./cmd/api/main.go


# --- Etapa 2: Final ---
# Usamos una imagen 'distroless' de Google. Estas imágenes son ultra-minimalistas.
# Contienen únicamente la aplicación y sus dependencias de tiempo de ejecución.
# No tienen shell, gestor de paquetes (apt, apk), ni utilidades comunes (ls, cat).
# Esto reduce drásticamente la superficie de ataque de seguridad.
FROM gcr.io/distroless/static-debian11

# Establecemos el directorio de trabajo.
WORKDIR /app

# Copiamos los certificados de CA desde la etapa 'builder'.
# Esto es necesario para que nuestra aplicación pueda hacer peticiones HTTPS si lo necesitara.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copiamos el directorio 'docs' generado por Swagger desde la etapa 'builder'.
# Esto es necesario para que el handler de gin-swagger pueda servir la UI.
COPY --from=builder /app/docs ./docs

# Copiamos ÚNICAMENTE el binario compilado desde la etapa 'builder'.
COPY --from=builder /app/main .

# Exponemos el puerto en el que nuestra aplicación Gin está escuchando dentro del contenedor.
EXPOSE 8080

# Definimos el usuario no-root que ejecutará la aplicación para mayor seguridad.
# El usuario 'nonroot' con ID 65532 es un estándar en imágenes distroless.
USER nonroot:nonroot

# El comando que se ejecutará cuando el contenedor se inicie.
# Ejecuta nuestro binario compilado.
CMD ["/app/main"]