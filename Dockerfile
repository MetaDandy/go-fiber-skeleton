# ---------- build stage ----------
FROM golang:1.25.5-alpine AS builder

# Directorio de trabajo
WORKDIR /src
    
# Dependencias del sistema necesarias para compilación
RUN apk add --no-cache git gcc musl-dev 
    
# Dependencias de Go
COPY go.mod go.sum ./
RUN go mod download
    
# Copiar el código fuente
COPY . .
    
# Reconocer arquitectura dinámica de BuildKit y compilar para ella
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -o app ./cmd
    
# ---------- runtime stage ----------
FROM alpine:3.20
    
# Instalar certificados de CA
RUN apk add --no-cache ca-certificates ffmpeg
    
# Directorio de trabajo para la app
WORKDIR /app
    
# Copiar binario compilado
COPY --from=builder /src/app .
    
# Permisos de ejecución
RUN chmod +x app
    
# Puerto inyectado por Render
ENV PORT 8000
EXPOSE ${PORT}
    
# Comando de arranque: la app debe leer os.Getenv("PORT")
CMD ["./app"]
    