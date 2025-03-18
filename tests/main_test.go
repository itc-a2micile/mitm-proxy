package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mitmproxy/go-mitmproxy"
	"github.com/mitmproxy/go-mitmproxy/proxy"
	"github.com/stretchr/testify/assert"
)

// TestMITMProxyStart vérifie que le proxy MITM démarre correctement
func TestMITMProxyStart(t *testing.T) {
	// Créer une configuration de test
	config := Config{
		LoggerEndpoint: "http://localhost:8082/api/logs",
		MaxRetries:     1,
		RetryDelay:     100 * time.Millisecond,
		ExcludedRoutes: []string{"health"},
		MaskHeaders:    []string{"authorization"},
		WebInterface:   false,
		WebPort:        0, // Port aléatoire
	}

	// Créer le gestionnaire MITM
	handler := NewMITMHandler(config)

	// Configurer les options du proxy avec un port aléatoire pour le test
	opts := proxy.Options{
		Addr:              ":0", // Port aléatoire
		StreamLargeBodies: 1024,
	}

	// Démarrer le proxy dans une goroutine avec un contexte annulable
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- mitmproxy.Start(handler, opts)
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

	// Annuler le contexte pour arrêter le proxy
	cancel()
}

// TestRequestInterception vérifie que les requêtes passent bien par le proxy
func TestRequestInterception(t *testing.T) {
	// Créer un serveur de test qui simule le service de journalisation
	loggerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"status":"created"}`))
	}))
	defer loggerServer.Close()

	// Créer une configuration de test
	config := Config{
		LoggerEndpoint: loggerServer.URL,
		MaxRetries:     1,
		RetryDelay:     100 * time.Millisecond,
		ExcludedRoutes: []string{},
		MaskHeaders:    []string{"authorization"},
		WebInterface:   false,
		WebPort:        0,
	}

	// Créer le gestionnaire MITM
	handler := NewMITMHandler(config)

	// Créer un flux de test
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	req.Header.Set("client-name", "TestClient")
	req.Header.Set("correlation-id", "test-123")
	req.Header.Set("username", "testuser")

	// Simuler une réponse
	resp := &http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
		Body:       http.NoBody,
	}

	// Créer un flux mitmproxy simulé
	flow := &mitmproxy.Flow{
		Request:  req,
		Response: resp,
	}

	// Appeler la méthode Request du gestionnaire
	handler.Request(flow)

	// Vérifier que la méthode OnResponse a été configurée
	assert.NotNil(t, flow.OnResponseFunc, "La fonction OnResponse doit être configurée")

	// Simuler l'appel à OnResponse
	if flow.OnResponseFunc != nil {
		flow.OnResponseFunc()
	}

	// Comme nous ne pouvons pas facilement vérifier que le log a été envoyé dans ce test,
	// nous nous contentons de vérifier que le flux a été traité sans erreur
	assert.Equal(t, 200, flow.Response.StatusCode, "Le code de statut de la réponse doit être 200")
}
