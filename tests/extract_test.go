package main_test

import (
	"bytes"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// RequestData - Structure pour stocker les données extraites d'une requête
type RequestData struct {
	ClientName    string
	CorrelationID string
	User          string
	HTTPMethod    string
	HTTPUrl       string
	HTTPHeaders   map[string]string
	HTTPBody      string
	OccuredTime   time.Time
}

// extractRequestData extrait les données importantes d'une requête HTTP
func extractRequestData(req *http.Request) RequestData {
	// Extraire les informations de la requête
	clientName := req.Header.Get("client-name")
	if clientName == "" {
		clientName = "Anonyme"
	}

	correlationID := req.Header.Get("correlation-id")
	if correlationID == "" {
		correlationID = "generated-id" // Dans une implémentation réelle, on générerait un UUID
	}

	user := req.Header.Get("username")
	if user == "" {
		user = req.Header.Get("user")
	}
	if user == "" {
		user = "Anonyme"
	}

	// Créer une map pour les en-têtes HTTP
	headers := make(map[string]string)
	for name, values := range req.Header {
		headers[name] = values[0]
	}

	// Lire le corps de la requête
	var bodyString string
	if req.Body != nil {
		bodyBytes, _ := io.ReadAll(req.Body)
		bodyString = string(bodyBytes)
		// Remettre le corps dans la requête
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	return RequestData{
		ClientName:    clientName,
		CorrelationID: correlationID,
		User:          user,
		HTTPMethod:    req.Method,
		HTTPUrl:       req.URL.String(),
		HTTPHeaders:   headers,
		HTTPBody:      bodyString,
		OccuredTime:   time.Now(),
	}
}

// TestExtractRequestData vérifie que les données sont correctement extraites d'une requête
func TestExtractRequestData(t *testing.T) {
	// Créer une requête de test
	body := bytes.NewBufferString(`{"key":"value"}`)
	req, _ := http.NewRequest("POST", "http://test.com/api/resource", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("client-name", "ServiceA")
	req.Header.Set("correlation-id", "abcd-1234")
	req.Header.Set("username", "jdoe")

	// Extraire les données
	data := extractRequestData(req)

	// Vérifier les données extraites
	assert.Equal(t, "ServiceA", data.ClientName, "Le nom du client doit être correctement extrait")
	assert.Equal(t, "abcd-1234", data.CorrelationID, "L'ID de corrélation doit être correctement extrait")
	assert.Equal(t, "jdoe", data.User, "L'utilisateur doit être correctement extrait")
	assert.Equal(t, "POST", data.HTTPMethod, "La méthode HTTP doit être correctement extraite")
	assert.Equal(t, "http://test.com/api/resource", data.HTTPUrl, "L'URL doit être correctement extraite")
	assert.Equal(t, "application/json", data.HTTPHeaders["Content-Type"], "Les en-têtes doivent être correctement extraits")
	assert.Equal(t, `{"key":"value"}`, data.HTTPBody, "Le corps de la requête doit être correctement extrait")
}

// TestExtractRequestDataWithDefaults vérifie que les valeurs par défaut sont utilisées lorsque les en-têtes sont absents
func TestExtractRequestDataWithDefaults(t *testing.T) {
	// Créer une requête de test sans en-têtes spécifiques
	req, _ := http.NewRequest("GET", "http://test.com", nil)

	// Extraire les données
	data := extractRequestData(req)

	// Vérifier les valeurs par défaut
	assert.Equal(t, "Anonyme", data.ClientName, "Le nom du client par défaut doit être 'Anonyme'")
	assert.NotEmpty(t, data.CorrelationID, "Un ID de corrélation doit être généré")
	assert.Equal(t, "Anonyme", data.User, "L'utilisateur par défaut doit être 'Anonyme'")
}

// TestExtractRequestDataWithUserHeader vérifie que l'en-tête 'user' est utilisé si 'username' est absent
func TestExtractRequestDataWithUserHeader(t *testing.T) {
	// Créer une requête de test avec l'en-tête 'user' au lieu de 'username'
	req, _ := http.NewRequest("GET", "http://test.com", nil)
	req.Header.Set("user", "john.doe")

	// Extraire les données
	data := extractRequestData(req)

	// Vérifier que l'utilisateur est extrait de l'en-tête 'user'
	assert.Equal(t, "john.doe", data.User, "L'utilisateur doit être extrait de l'en-tête 'user'")
}
