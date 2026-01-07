package controladores

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
)

// ControladorAuth gerencia as operações de autenticação
type ControladorAuth struct {
	authClient *auth.Client
	apiKey     string
}

// NovoControladorAuth cria uma nova instância do controlador de autenticação
func NovoControladorAuth(ctx context.Context, credentialsPath string) (*ControladorAuth, error) {
	var app *firebase.App
	var err error

	// Se credentialsPath for fornecido, usa o arquivo de credenciais
	// Caso contrário, usa as credenciais padrão do ambiente (ADC)
	if credentialsPath != "" {
		opt := option.WithCredentialsFile(credentialsPath)
		app, err = firebase.NewApp(ctx, nil, opt)
	} else {
		app, err = firebase.NewApp(ctx, nil)
	}

	if err != nil {
		return nil, err
	}

	client, err := app.Auth(ctx)
	if err != nil {
		return nil, err
	}

	// API Key do Firebase (necessária para trocar custom token por ID token)
	apiKey := os.Getenv("FIREBASE_API_KEY")

	return &ControladorAuth{
		authClient: client,
		apiKey:     apiKey,
	}, nil
}

// LoginRequest representa a requisição de login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"usuario@exemplo.com"`
	Password string `json:"password" binding:"required,min=6" example:"senha123"`
}

// LoginResponse representa a resposta de login
type LoginResponse struct {
	IDToken      string `json:"idToken" example:"eyJhbGciOiJSUzI1NiIsImtpZCI6IjEyMzQ1Njc4OTAifQ..."`
	RefreshToken string `json:"refreshToken,omitempty" example:"..."`
	UID          string `json:"uid" example:"abc123def456"`
	Email        string `json:"email" example:"usuario@exemplo.com"`
	ExpiresIn    string `json:"expiresIn" example:"3600"`
}

// FirebaseTokenResponse representa a resposta da API do Firebase
type FirebaseTokenResponse struct {
	IDToken      string `json:"idToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    string `json:"expiresIn"`
}

// GerarToken gera um ID token para autenticação
// @Summary      Gerar token de autenticação
// @Description  Gera um ID token JWT do Firebase para um usuário específico
// @Tags         Autenticação
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "Credenciais de login"
// @Success      200 {object} LoginResponse "Token gerado com sucesso"
// @Failure      400 {object} map[string]string "Requisição inválida"
// @Failure      401 {object} map[string]string "Credenciais inválidas"
// @Failure      500 {object} map[string]string "Erro interno do servidor"
// @Router       /api/v2/auth/login [post]
func (ctrl *ControladorAuth) GerarToken(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Dados inválidos: " + err.Error()})
		return
	}

	// 1. Buscar o usuário pelo email para validar que existe
	user, err := ctrl.authClient.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		slog.Warn("Usuário não encontrado", "email", req.Email, "erro", err)
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Credenciais inválidas"})
		return
	}

	// 2. Gerar um custom token
	customToken, err := ctrl.authClient.CustomToken(c.Request.Context(), user.UID)
	if err != nil {
		slog.Error("Erro ao gerar custom token", "erro", err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao gerar token"})
		return
	}

	// 3. Trocar custom token por ID token usando Firebase Auth REST API
	if ctrl.apiKey != "" {
		idToken, refreshToken, expiresIn, err := ctrl.trocarCustomTokenPorIDToken(customToken)
		if err != nil {
			slog.Error("Erro ao trocar custom token por ID token", "erro", err)
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao gerar ID token"})
			return
		}

		slog.Info("ID Token gerado com sucesso", "uid", user.UID, "email", user.Email)

		c.JSON(http.StatusOK, LoginResponse{
			IDToken:      idToken,
			RefreshToken: refreshToken,
			UID:          user.UID,
			Email:        user.Email,
			ExpiresIn:    expiresIn,
		})
		return
	}

	// Fallback: retorna custom token com aviso
	slog.Warn("FIREBASE_API_KEY não configurada, retornando custom token. Configure FIREBASE_API_KEY para obter ID tokens.")
	c.JSON(http.StatusOK, gin.H{
		"customToken": customToken,
		"uid":         user.UID,
		"email":       user.Email,
		"aviso":       "Este é um custom token. Para obter um ID token, configure FIREBASE_API_KEY ou use o Firebase Client SDK no frontend.",
		"instrucoes":  "Use este custom token com Firebase Auth Client SDK: firebase.auth().signInWithCustomToken(customToken)",
	})
}

// trocarCustomTokenPorIDToken troca um custom token por um ID token usando Firebase Auth REST API
func (ctrl *ControladorAuth) trocarCustomTokenPorIDToken(customToken string) (string, string, string, error) {
	url := fmt.Sprintf("https://identitytoolkit.googleapis.com/v1/accounts:signInWithCustomToken?key=%s", ctrl.apiKey)

	payload := map[string]interface{}{
		"token":             customToken,
		"returnSecureToken": true,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", "", "", err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", "", "", fmt.Errorf("firebase API error: %s", string(body))
	}

	var tokenResp FirebaseTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", "", "", err
	}

	return tokenResp.IDToken, tokenResp.RefreshToken, tokenResp.ExpiresIn, nil
}

// VerificarToken verifica se o token é válido
// @Summary      Verificar token
// @Description  Verifica se o token JWT fornecido é válido
// @Tags         Autenticação
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} map[string]interface{} "Token válido"
// @Failure      401 {object} map[string]string "Token inválido"
// @Router       /api/v2/auth/verify [get]
func (ctrl *ControladorAuth) VerificarToken(c *gin.Context) {
	// Se chegou aqui, o token já foi validado pelo middleware
	uid, _ := c.Get("uid")
	email, _ := c.Get("email")
	claims, _ := c.Get("firebase_claims")

	c.JSON(http.StatusOK, gin.H{
		"valido": true,
		"uid":    uid,
		"email":  email,
		"claims": claims,
	})
}
