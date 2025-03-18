package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lqqyt2423/go-mitmproxy/proxy"
	"github.com/lqqyt2423/go-mitmproxy/web"
)

// LogModel - Structure pour la journalisation des requêtes et réponses
type LogModel struct {
	ID             string            `json:"id,omitempty"`
	CorrelationID  string            `json:"correlation_id"`
	ClientName     string            `json:"client_name"`
	User           string            `json:"user"`
	OccuredTime    time.Time         `json:"occured_time"`
	HTTPMethod     string            `json:"http_method"`
	HTTPUrl        string            `json:"http_url"`
	HTTPHeaders    map[string]string `json:"http_headers"`
	HTTPBody       string            `json:"http_body"`
	LogTextShort   string            `json:"log_text_short"`
	LogText        string            `json:"log_text"`
	HTTPReturnCode int               `json:"http_return_code,omitempty"`
	HTTPReturnBody string            `json:"http_response_body,omitempty"`
	ExecutionTime  int64             `json:"execution_time,omitempty"`
	LogType        string            `json:"log_type,omitempty"` // "info", "error", "critical"
}

// Config - Configuration du proxy MITM
type Config struct {
	LoggerEndpoint string
	MaxRetries     int
	RetryDelay     time.Duration
	ExcludedRoutes []string
	MaskHeaders    []string
	WebInterface   bool
	WebPort        int
}

// MITMHandler - Gestionnaire pour le proxy MITM
type MITMHandler struct {
	config     Config
	httpClient *http.Client
}

// NewMITMHandler - Créer un nouveau gestionnaire MITM avec la configuration donnée
func NewMITMHandler(config Config) *MITMHandler {
	// Valeurs par défaut si non fournies
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 500 * time.Millisecond
	}
	if len(config.MaskHeaders) == 0 {
		config.MaskHeaders = []string{"authorization", "password", "token", "api-key"}
	}
	if config.WebPort == 0 {
		config.WebPort = 8081
	}

	// Créer le client HTTP
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &MITMHandler{
		config:     config,
		httpClient: httpClient,
	}
}

// Request - Intercepte les requêtes entrantes
func (h *MITMHandler) Request(f *mitmproxy.Flow) {
	// Vérifier si la route doit être exclue
	for _, route := range h.config.ExcludedRoutes {
		if strings.Contains(f.Request.URL.Path, route) {
			return
		}
	}

	// Enregistrer l'heure de début pour calculer le temps d'exécution
	startTime := time.Now()

	// Générer un ID unique pour cette requête
	requestID := uuid.New().String()

	// Extraire les informations de la requête
	clientName := f.Request.Header.Get("client-name")
	if clientName == "" {
		clientName = "Anonyme"
	}

	correlationID := f.Request.Header.Get("correlation-id")
	if correlationID == "" {
		correlationID = requestID
	}

	user := f.Request.Header.Get("username")
	if user == "" {
		user = f.Request.Header.Get("user")
	}
	if user == "" {
		user = "Anonyme"
	}

	// Créer une map pour les en-têtes HTTP
	headers := make(map[string]string)
	for name, values := range f.Request.Header {
		// Masquer les en-têtes sensibles
		headerLower := strings.ToLower(name)
		for _, mask := range h.config.MaskHeaders {
			if headerLower == strings.ToLower(mask) {
				headers[name] = "********"
				continue
			}
		}
		headers[name] = strings.Join(values, ", ")
	}

	// Lire le corps de la requête sans le modifier
	bodyBytes, err := io.ReadAll(f.Request.Body)
	if err != nil {
		log.Printf("Erreur lors de la lecture du corps de la requête: %v", err)
		return
	}

	// Remettre le corps dans la requête pour ne pas interférer avec le flux
	f.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Créer l'entrée de journal initiale
	logEntry := &LogModel{
		ID:            requestID,
		CorrelationID: correlationID,
		ClientName:    clientName,
		User:          user,
		OccuredTime:   time.Now(),
		HTTPMethod:    f.Request.Method,
		HTTPUrl:       f.Request.URL.String(),
		HTTPHeaders:   headers,
		HTTPBody:      string(bodyBytes),
		LogTextShort:  "Requête interceptée",
		LogText:       fmt.Sprintf("Requête interceptée: %s %s", f.Request.Method, f.Request.URL.String()),
		LogType:       "info",
	}

	// Envoyer le journal initial au service de journalisation
	go h.sendLogToLogger(logEntry, "create")

	// Configurer un gestionnaire pour la réponse
	f.OnResponse(func() {
		// Lire le corps de la réponse sans le modifier
		responseBodyBytes, err := io.ReadAll(f.Response.Body)
		if err != nil {
			log.Printf("Erreur lors de la lecture du corps de la réponse: %v", err)
			return
		}

		// Remettre le corps dans la réponse
		f.Response.Body = io.NopCloser(bytes.NewBuffer(responseBodyBytes))

		// Mettre à jour le journal avec les données de réponse
		executionTime := time.Since(startTime).Milliseconds()

		logEntry.HTTPReturnCode = f.Response.StatusCode
		logEntry.HTTPReturnBody = string(responseBodyBytes)
		logEntry.ExecutionTime = executionTime

		// Mettre à jour le texte du journal en fonction du code d'état
		if f.Response.StatusCode >= 400 {
			logEntry.LogTextShort = fmt.Sprintf("Erreur: %d", f.Response.StatusCode)
			logEntry.LogText = fmt.Sprintf("La requête a échoué avec le code d'état %d: %s %s",
				f.Response.StatusCode, logEntry.HTTPMethod, logEntry.HTTPUrl)

			if f.Response.StatusCode >= 500 {
				logEntry.LogType = "critical"
			} else {
				logEntry.LogType = "error"
			}
		} else {
			logEntry.LogTextShort = fmt.Sprintf("Succès: %d", f.Response.StatusCode)
			logEntry.LogText = fmt.Sprintf("La requête s'est terminée avec succès avec le code d'état %d: %s %s",
				f.Response.StatusCode, logEntry.HTTPMethod, logEntry.HTTPUrl)
			logEntry.LogType = "info"
		}

		// Envoyer le journal mis à jour au service de journalisation
		go h.sendLogToLogger(logEntry, "update")
	})
}

// Response - Intercepte les réponses (non utilisé car nous utilisons OnResponse dans Request)
func (h *MITMHandler) Response(f *mitmproxy.Flow) {
	// Nous utilisons OnResponse dans la méthode Request
}

// sendLogToLogger - Envoyer une entrée de journal au service de journalisation
func (h *MITMHandler) sendLogToLogger(logEntry *LogModel, action string) {
	// Convertir l'entrée de journal en JSON
	jsonData, err := json.Marshal(logEntry)
	if err != nil {
		log.Printf("Erreur lors de la sérialisation de l'entrée de journal: %v", err)
		return
	}

	// Déterminer le point de terminaison en fonction de l'action
	endpoint := h.config.LoggerEndpoint
	if action == "update" {
		endpoint = fmt.Sprintf("%s/update", endpoint)
	}

	// Envoyer le journal au logger avec des tentatives
	var resp *http.Response
	for i := 0; i <= h.config.MaxRetries; i++ {
		req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("Erreur lors de la création de la requête vers le logger: %v", err)
			time.Sleep(h.config.RetryDelay)
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		resp, err = h.httpClient.Do(req)
		if err == nil && resp.StatusCode < 500 {
			resp.Body.Close()
			break
		}

		if resp != nil {
			resp.Body.Close()
		}

		if i < h.config.MaxRetries {
			// Backoff exponentiel
			backoff := h.config.RetryDelay * time.Duration(1<<uint(i))
			time.Sleep(backoff)
		}
	}

	if resp == nil || resp.StatusCode >= 400 {
		log.Printf("Échec de l'envoi du journal au logger après %d tentatives", h.config.MaxRetries)
	}
}

func main() {
	// Obtenir la configuration à partir des variables d'environnement ou utiliser les valeurs par défaut
	loggerEndpoint := getEnv("LOGGER_ENDPOINT", "http://logger-service/api/logs")
	maxRetries := getEnvInt("MAX_RETRIES", 3)
	retryDelay := getEnvDuration("RETRY_DELAY", 500*time.Millisecond)
	webInterface := getEnvBool("WEB_INTERFACE", true)
	webPort := getEnvInt("WEB_PORT", 8081)

	// Analyser les routes exclues
	excludedRoutesStr := getEnv("EXCLUDED_ROUTES", "")
	var excludedRoutes []string
	if excludedRoutesStr != "" {
		excludedRoutes = strings.Split(excludedRoutesStr, ",")
	}

	// Analyser les en-têtes à masquer
	maskHeadersStr := getEnv("MASK_HEADERS", "authorization,password,token,api-key")
	maskHeaders := strings.Split(maskHeadersStr, ",")

	// Créer la configuration
	config := Config{
		LoggerEndpoint: loggerEndpoint,
		MaxRetries:     maxRetries,
		RetryDelay:     retryDelay,
		ExcludedRoutes: excludedRoutes,
		MaskHeaders:    maskHeaders,
		WebInterface:   webInterface,
		WebPort:        webPort,
	}

	// Créer le gestionnaire MITM
	handler := NewMITMHandler(config)

	// Configurer les options du proxy
	opts := proxy.Options{
		Addr:              getEnv("LISTEN_ADDR", ":8080"),
		StreamLargeBodies: 1024 * 1024 * 5, // 5 MB
	}

	// Configurer l'interface web si activée
	if webInterface {
		webOpts := web.Options{
			Addr: fmt.Sprintf(":%d", webPort),
		}

		log.Printf("Démarrage du proxy MITM avec interface web sur :%d", webPort)
		log.Printf("Point de terminaison du logger: %s", loggerEndpoint)

		// Démarrer le proxy avec l'interface web
		if err := mitmproxy.StartWithWeb(handler, opts, webOpts); err != nil {
			log.Fatalf("Erreur lors du démarrage du proxy MITM: %v", err)
		}
	} else {
		log.Printf("Démarrage du proxy MITM sur %s", opts.Addr)
		log.Printf("Point de terminaison du logger: %s", loggerEndpoint)

		// Démarrer le proxy sans interface web
		if err := mitmproxy.Start(handler, opts); err != nil {
			log.Fatalf("Erreur lors du démarrage du proxy MITM: %v", err)
		}
	}
}

// Fonctions utilitaires pour les variables d'environnement
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value == "true" || value == "1" || value == "yes"
}

func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue := defaultValue
	fmt.Sscanf(value, "%d", &intValue)
	return intValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return duration
}
