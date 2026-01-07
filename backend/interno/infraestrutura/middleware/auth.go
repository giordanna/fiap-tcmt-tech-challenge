package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
)

// FirebaseAuth é o middleware de autenticação JWT usando Firebase
type FirebaseAuth struct {
	client *auth.Client
}

// NovoFirebaseAuth cria uma nova instância do middleware de autenticação
func NovoFirebaseAuth(ctx context.Context, credentialsPath string) (*FirebaseAuth, error) {
	var app *firebase.App
	var err error

	// Se credentialsPath for fornecido, usa o arquivo de credenciais
	// Caso contrário, usa as credenciais padrão do ambiente (ADC - Application Default Credentials)
	if credentialsPath != "" {
		opt := option.WithCredentialsFile(credentialsPath)
		app, err = firebase.NewApp(ctx, nil, opt)
	} else {
		// Usa Application Default Credentials (útil para Cloud Run)
		app, err = firebase.NewApp(ctx, nil)
	}

	if err != nil {
		return nil, err
	}

	client, err := app.Auth(ctx)
	if err != nil {
		return nil, err
	}

	slog.Info("Firebase Auth inicializado com sucesso")
	return &FirebaseAuth{client: client}, nil
}

// Middleware retorna o middleware do Gin para autenticação JWT
func (f *FirebaseAuth) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extrai o token do header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"erro": "Token de autenticação não fornecido",
			})
			c.Abort()
			return
		}

		// Verifica se o header está no formato "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"erro": "Formato de token inválido. Use: Bearer <token>",
			})
			c.Abort()
			return
		}

		token := parts[1]

		// Verifica o token com Firebase
		decodedToken, err := f.client.VerifyIDToken(c.Request.Context(), token)
		if err != nil {
			slog.Warn("Token inválido", "erro", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"erro": "Token inválido ou expirado",
			})
			c.Abort()
			return
		}

		// Armazena informações do usuário no contexto
		c.Set("uid", decodedToken.UID)
		c.Set("email", decodedToken.Claims["email"])
		c.Set("firebase_claims", decodedToken.Claims)

		slog.Info("Usuário autenticado", "uid", decodedToken.UID, "email", decodedToken.Claims["email"])

		c.Next()
	}
}

// GetUID retorna o UID do usuário autenticado do contexto
func GetUID(c *gin.Context) string {
	uid, exists := c.Get("uid")
	if !exists {
		return ""
	}
	return uid.(string)
}

// GetEmail retorna o email do usuário autenticado do contexto
func GetEmail(c *gin.Context) string {
	email, exists := c.Get("email")
	if !exists {
		return ""
	}
	if email == nil {
		return ""
	}
	return email.(string)
}
