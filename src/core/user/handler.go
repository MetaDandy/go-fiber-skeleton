package user

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/MetaDandy/go-fiber-skeleton/helper"
	"github.com/MetaDandy/go-fiber-skeleton/src/auth"
	"github.com/MetaDandy/go-fiber-skeleton/src/enum"
	"github.com/gofiber/fiber/v3"
)

type Handler interface {
	RegisterRoutes(router fiber.Router)
	Create(c fiber.Ctx) error
	FindAll(c fiber.Ctx) error
	FindByID(c fiber.Ctx) error
	Update(c fiber.Ctx) error
	Delete(c fiber.Ctx) error
	OAuthLogin(c fiber.Ctx) error
	OAuthCallback(c fiber.Ctx) error
}

type handler struct {
	service Service
}

func NewHandler(service Service) Handler {
	return &handler{
		service: service,
	}
}

// generateState genera un estado aleatorio para CSRF protection
func generateState() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	state := make([]byte, 32)
	for i := range state {
		state[i] = charset[rand.Intn(len(charset))]
	}
	return string(state)
}

func (h *handler) RegisterRoutes(router fiber.Router) {
	users := router.Group("/users")
	users.Post("/", h.Create)
	users.Get("/", h.FindAll)
	users.Get("/:id", h.FindByID)
	users.Patch("/:id", h.Update)
	users.Delete("/:id", h.Delete)

	auth := router.Group("/auth")
	auth.Get("/login/:provider", h.OAuthLogin)
	auth.Get("/callback", h.OAuthCallback)
}

func (h *handler) Create(c fiber.Ctx) error {
	input := new(Create)
	if err := c.Bind().Body(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid input")
	}

	err := h.service.Create(*input)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Could not create user")
	}

	return c.SendStatus(fiber.StatusCreated)
}

func (h *handler) FindAll(c fiber.Ctx) error {
	opts := helper.NewFindAllOptionsFromQuery(c)
	finded, err := h.service.FindAll(opts)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(finded)
}

func (h *handler) FindByID(c fiber.Ctx) error {
	id := c.Params("id")
	finded, err := h.service.FindByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(finded)
}

func (h *handler) Update(c fiber.Ctx) error {
	var input Update

	if err := c.Bind().Body(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	id := c.Params("id")

	err := h.service.Update(id, input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusOK)
}

func (h *handler) Delete(c fiber.Ctx) error {
	id := c.Params("id")

	if err := h.service.Delete(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusNoContent).Send(nil)
}

// OAuthLogin inicia el flujo de login OAuth
func (h *handler) OAuthLogin(c fiber.Ctx) error {
	provider := c.Params("provider")

	fmt.Printf("\n========================================\n")
	fmt.Printf("🔐 [OAuth Login] RAW Params:\n")
	fmt.Printf("   Provider recibido: %s\n", provider)
	fmt.Printf("   Bytes: %v\n", []byte(provider))
	fmt.Printf("   Length: %d\n", len(provider))
	fmt.Printf("   URL actual: %s\n", c.OriginalURL())
	fmt.Printf("========================================\n\n")

	// Validar proveedor
	if !enum.IsValidAuthProvider(provider) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "provider no soportado",
		})
	}

	// Cargar credenciales
	creds, err := auth.LoadCredentials(provider)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Generar estado para CSRF
	state := generateState()

	// Guardar estado + provider en cookie con SameSite=Lax para desarrollo
	c.Cookie(&fiber.Cookie{
		Name:     "oauth_state",
		Value:    fmt.Sprintf("%s:%s", state, provider), // Format: state:provider
		Expires:  time.Now().Add(15 * time.Minute),
		Path:     "/",
		SameSite: "Lax",
		HTTPOnly: true,
	})

	// Construir redirect URL
	redirectURL := fmt.Sprintf("%s/api/v1/auth/callback", c.BaseURL())

	// Obtener URL de autorización
	authURL, err := auth.GetAuthURL(provider, creds, redirectURL, state)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	fmt.Printf("🔐 [OAuth Login] Provider: %s\n", provider)
	fmt.Printf("📍 Auth URL: %s\n", authURL)
	fmt.Printf("🔒 State: %s\n\n", state)

	return c.Redirect().To(authURL)
}

// OAuthCallback maneja el callback de OAuth
func (h *handler) OAuthCallback(c fiber.Ctx) error {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "missing parameters: code or state",
		})
	}

	// Obtener cookie con state:provider
	cookieValue := c.Cookies("oauth_state")
	if cookieValue == "" {
		fmt.Printf("❌ [OAuth Callback] Cookie no encontrada\n\n")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "cookie not found",
		})
	}

	// Parse: state:provider de la cookie
	parts := strings.Split(cookieValue, ":")
	if len(parts) != 2 {
		fmt.Printf("❌ [OAuth Callback] Cookie formato inválido\n\n")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid cookie format",
		})
	}

	cookieState := parts[0]
	provider := parts[1]

	// Validar state
	if cookieState != state {
		fmt.Printf("❌ [OAuth Callback] CSRF validation failed\n")
		fmt.Printf("   Expected: %s\n", cookieState)
		fmt.Printf("   Got: %s\n\n", state)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid state",
		})
	}

	fmt.Printf("✓ [OAuth Callback] State validado - Provider: %s\n", provider)

	// Cargar credenciales
	creds, err := auth.LoadCredentials(provider)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Construir redirect URL
	redirectURL := fmt.Sprintf("%s/api/v1/auth/callback", c.BaseURL())

	// Intercambiar código por token
	token, err := auth.ExchangeCode(provider, creds, redirectURL, code)
	if err != nil {
		fmt.Printf("❌ [OAuth Callback] Failed to exchange code: %v\n\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Obtener información del usuario
	userInfo, err := auth.GetUserInfo(context.Background(), provider, token)
	if err != nil {
		fmt.Printf("❌ [OAuth Callback] Failed to get user info: %v\n\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Mostrar información completa en consola
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("✅ [OAuth Callback] Autenticación exitosa\n")
	fmt.Printf("🔐 Provider: %s\n", provider)
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("👤 Información del Usuario:\n")
	fmt.Printf("   ID: %s\n", userInfo.ID)
	fmt.Printf("   Email: %s\n", userInfo.Email)
	fmt.Printf("   Nombre: %s\n", userInfo.Name)
	fmt.Printf("   Imagen: %s\n", userInfo.Image)
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("🔑 Token:\n")
	fmt.Printf("   Access Token: %s\n", token.AccessToken[:20]+"...")
	fmt.Printf("   Token Type: %s\n", token.TokenType)
	fmt.Printf("   Expiry: %v\n", token.Expiry)
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	// Retornar información del usuario
	return c.JSON(fiber.Map{
		"success":  true,
		"provider": provider,
		"user": fiber.Map{
			"id":    userInfo.ID,
			"email": userInfo.Email,
			"name":  userInfo.Name,
			"image": userInfo.Image,
		},
		"token": fiber.Map{
			"access_token": token.AccessToken,
			"token_type":   token.TokenType,
			"expiry":       token.Expiry,
		},
	})
}
