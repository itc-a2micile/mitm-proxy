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
	ProxyPort      int // Renommé de WebPort à ProxyPort pour plus de clarté
	WebPort        int
}

// MITMHandler - Gestionnaire pour le proxy MITM
type MITMHandler struct {
	config     Config
	httpClient *http.Client
	flowData   map[string]*LogModel
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
	if config.ProxyPort == 0 {
		config.ProxyPort = 9080
	}
	if config.WebPort == 0 {
		config.WebPort = 9081
	}

	// Créer le client HTTP
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &MITMHandler{
		config:     config,
		httpClient: httpClient,
		flowData:   make(map[string]*LogModel),
	}
}

// Request - Intercepte les requêtes entrantes
func (h *MITMHandler) Request(f *proxy.Flow) {
	req := f.Request

	// Vérifier si la route doit être exclue
	for _, route := range h.config.ExcludedRoutes {
		if strings.Contains(req.URL.String(), route) {
			return
		}
	}

	// Enregistrer l'heure de début pour calculer le temps d'exécution
	startTime := time.Now()

	// Générer un ID unique pour cette requête
	requestID := uuid.New().String()

	// Extraire les informations de la requête
	clientName := req.Header.Get("client-name")
	if clientName == "" {
		clientName = "Anonyme"
	}

	correlationID := req.Header.Get("correlation-id")
	if correlationID == "" {
		correlationID = requestID
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

	// Lire le corps de la requête
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes = req.Body
	}

	// Créer l'entrée de journal initiale
	logEntry := &LogModel{
		ID:            requestID,
		CorrelationID: correlationID,
		ClientName:    clientName,
		User:          user,
		OccuredTime:   time.Now(),
		HTTPMethod:    req.Method,
		HTTPUrl:       req.URL.String(),
		HTTPHeaders:   headers,
		HTTPBody:      string(bodyBytes),
		LogTextShort:  "Requête interceptée",
		LogText:       fmt.Sprintf("Requête interceptée: %s %s", req.Method, req.URL.String()),
		LogType:       "info",
	}

	// Envoyer le journal initial au service de journalisation
	go h.sendLogToLogger(logEntry, "create")

	// Stocker les données pour les récupérer dans Response
	h.flowData[f.Id.String()] = logEntry

	// Ecrire en console le temps d'exécution
	log.Printf("Temps d'exécution: %d ms", time.Since(startTime).Milliseconds())
}

// Response - Intercepte les réponses
func (h *MITMHandler) Response(f *proxy.Flow) {
	// Récupérer les données stockées
	logEntry, ok := h.flowData[f.Id.String()]
	if !ok {
		return
	}

	resp := f.Response
	startTime := logEntry.OccuredTime

	// Lire le corps de la réponse
	var responseBodyBytes []byte
	if resp.Body != nil {
		responseBodyBytes = resp.Body
	}

	// Mettre à jour le journal avec les données de réponse
	executionTime := time.Since(startTime).Milliseconds()

	logEntry.HTTPReturnCode = resp.StatusCode
	logEntry.HTTPReturnBody = string(responseBodyBytes)
	logEntry.ExecutionTime = executionTime

	// Mettre à jour le texte du journal en fonction du code d'état
	if resp.StatusCode >= 400 {
		logEntry.LogTextShort = fmt.Sprintf("Erreur: %d", resp.StatusCode)
		logEntry.LogText = fmt.Sprintf("La requête a échoué avec le code d'état %d: %s %s",
			resp.StatusCode, logEntry.HTTPMethod, logEntry.HTTPUrl)

		if resp.StatusCode >= 500 {
			logEntry.LogType = "critical"
		} else {
			logEntry.LogType = "error"
		}
	}

	// Envoyer le journal mis à jour au service de journalisation
	go h.sendLogToLogger(logEntry, "update")

	// Nettoyer les données stockées
	delete(h.flowData, f.Id.String())
}

// Done - Appelé lorsque le flux est terminé
func (h *MITMHandler) Done(f *proxy.Flow) {
	// Nettoyer les données stockées si ce n'est pas déjà fait
	delete(h.flowData, f.Id.String())
}

// Requis par l'interface proxy.Addon
func (h *MITMHandler) ResponseHeader(f *proxy.Flow)                                 {}
func (h *MITMHandler) RequestHeader(f *proxy.Flow)                                  {}
func (h *MITMHandler) Connect(f *proxy.Flow)                                        {}
func (h *MITMHandler) Connected(f *proxy.Flow)                                      {}
func (h *MITMHandler) Error(f *proxy.Flow)                                          {}
func (h *MITMHandler) HTTPError(f *proxy.Flow)                                      {}
func (h *MITMHandler) ParentProxy(*proxy.Flow) string                               { return "" }
func (h *MITMHandler) AccessProxyServer(req *http.Request, res http.ResponseWriter) {}
func (h *MITMHandler) StreamRequestModifier(f *proxy.Flow, in io.Reader) io.Reader  { return in }
func (h *MITMHandler) StreamResponseModifier(f *proxy.Flow, in io.Reader) io.Reader { return in }
func (h *MITMHandler) ClientConnected(client *proxy.ClientConn)                     {}
func (h *MITMHandler) ClientDisconnected(client *proxy.ClientConn)                  {}
func (h *MITMHandler) ServerConnected(ctx *proxy.ConnContext)                       {}
func (h *MITMHandler) ServerDisconnected(ctx *proxy.ConnContext)                    {}
func (h *MITMHandler) TlsEstablishedServer(ctx *proxy.ConnContext)                  {}
func (h *MITMHandler) Requestheaders(f *proxy.Flow)                                 {}
func (h *MITMHandler) Responseheaders(f *proxy.Flow)                                {}

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
	// Charger la configuration depuis les variables d'environnement
	config := Config{
		LoggerEndpoint: getEnv("LOGGER_ENDPOINT", "http://localhost:8080/api/logs"),
		MaxRetries:     getEnvInt("MAX_RETRIES", 3),
		RetryDelay:     getEnvDuration("RETRY_DELAY", 500*time.Millisecond),
		ExcludedRoutes: strings.Split(getEnv("EXCLUDED_ROUTES", ""), ","),
		MaskHeaders:    strings.Split(getEnv("MASK_HEADERS", "authorization,password,token,api-key"), ","),
		WebInterface:   getEnvBool("WEB_INTERFACE", true),
		ProxyPort:      getEnvInt("PROXY_PORT", 9080),
		WebPort:        getEnvInt("WEB_PORT", 9081),
	}

	// Créer le gestionnaire MITM
	handler := NewMITMHandler(config)

	// Configurer les options du proxy
	opts := &proxy.Options{
		Addr:              fmt.Sprintf(":%d", config.ProxyPort),
		StreamLargeBodies: 1024 * 1024, // 1MB
	}

	// Créer et démarrer le proxy
	p, err := proxy.NewProxy(opts)
	if err != nil {
		log.Fatal(err)
	}

	// Ajouter le gestionnaire MITM comme addon
	p.AddAddon(handler)

	// Configurer et démarrer l'interface web si activée
	if config.WebInterface {
		webAddr := fmt.Sprintf(":%d", config.WebPort)
		webAddon := web.NewWebAddon(webAddr)
		p.AddAddon(webAddon)
		fmt.Printf("Interface web disponible sur http://localhost:%d\n", config.WebPort)
	}

	fmt.Printf("Proxy MITM démarré sur le port %d\n", config.ProxyPort)
	log.Fatal(p.Start())
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
