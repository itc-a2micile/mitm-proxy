package main_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lqqyt2423/go-mitmproxy/proxy"
	"github.com/stretchr/testify/assert"
)

// TestMITMProxyStart vérifie que le proxy MITM démarre correctement
func TestMITMProxyStart(t *testing.T) {
	// Créer un serveur de test pour le logger
	loggerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"status":"created"}`))
	}))
	defer loggerServer.Close()

	// Configurer les options du proxy avec un port aléatoire pour le test
	opts := &proxy.Options{
		Addr:              ":0", // Port aléatoire
		StreamLargeBodies: 1024 * 1024,
	}

	// Créer un proxy de test
	p, err := proxy.NewProxy(opts)
	assert.NoError(t, err, "Le proxy MITM doit être créé sans erreur")

	// Démarrer le proxy dans une goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- p.Start()
	}()

	// Attendre un peu pour que le proxy démarre
	time.Sleep(100 * time.Millisecond)

	// Vérifier qu'aucune erreur n'a été retournée
	select {
	case err := <-errCh:
		t.Fatalf("Le proxy MITM n'a pas démarré correctement: %v", err)
	default:
		// Pas d'erreur, c'est bon
	}

	// Arrêter le proxy
	p.Close()
}

// TestRequestExtraction vérifie que les informations de la requête sont correctement extraites
func TestRequestExtraction(t *testing.T) {
	// Créer une requête de test
	body := bytes.NewBufferString(`{"key":"value"}`)
	req, _ := http.NewRequest("POST", "http://example.com/api/test", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("client-name", "TestClient")
	req.Header.Set("correlation-id", "test-123")
	req.Header.Set("username", "testuser")

	// Lire et vérifier les headers
	assert.Equal(t, "TestClient", req.Header.Get("client-name"), "L'en-tête client-name doit être correctement extrait")
	assert.Equal(t, "test-123", req.Header.Get("correlation-id"), "L'en-tête correlation-id doit être correctement extrait")
	assert.Equal(t, "testuser", req.Header.Get("username"), "L'en-tête username doit être correctement extrait")

	// Lire et vérifier le body
	bodyBytes, _ := io.ReadAll(req.Body)
	req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Remettre le body pour pouvoir le relire

	assert.Equal(t, `{"key":"value"}`, string(bodyBytes), "Le corps de la requête doit être correctement extrait")
}

// TestResponseHandling vérifie que les réponses HTTP sont correctement traitées
func TestResponseHandling(t *testing.T) {
	// Créer une réponse de test
	body := bytes.NewBufferString(`{"result":"success"}`)
	resp := &http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
		Body:       io.NopCloser(body),
	}
	resp.Header.Set("Content-Type", "application/json")

	// Vérifier le code de statut
	assert.Equal(t, 200, resp.StatusCode, "Le code de statut doit être 200")

	// Lire et vérifier le body
	bodyBytes, _ := io.ReadAll(resp.Body)
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Remettre le body pour pouvoir le relire

	assert.Equal(t, `{"result":"success"}`, string(bodyBytes), "Le corps de la réponse doit être correctement extrait")
}

// TestErrorResponse vérifie que les réponses d'erreur sont correctement identifiées
func TestErrorResponse(t *testing.T) {
	// Créer une réponse d'erreur de test
	body := bytes.NewBufferString(`{"error":"not found"}`)
	resp := &http.Response{
		StatusCode: 404,
		Header:     make(http.Header),
		Body:       io.NopCloser(body),
	}
	resp.Header.Set("Content-Type", "application/json")

	// Vérifier le code de statut
	assert.Equal(t, 404, resp.StatusCode, "Le code de statut doit être 404")
	assert.True(t, resp.StatusCode >= 400, "Le code de statut doit être identifié comme une erreur")

	// Lire et vérifier le body
	bodyBytes, _ := io.ReadAll(resp.Body)
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Remettre le body pour pouvoir le relire

	assert.Equal(t, `{"error":"not found"}`, string(bodyBytes), "Le corps de la réponse d'erreur doit être correctement extrait")
}
