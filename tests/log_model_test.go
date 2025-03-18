package main_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
	LogType        string            `json:"log_type,omitempty"`
}

// convertToLogModel convertit les données de requête en un modèle de log
func convertToLogModel(reqData RequestData) LogModel {
	return LogModel{
		CorrelationID: reqData.CorrelationID,
		ClientName:    reqData.ClientName,
		User:          reqData.User,
		OccuredTime:   reqData.OccuredTime,
		HTTPMethod:    reqData.HTTPMethod,
		HTTPUrl:       reqData.HTTPUrl,
		HTTPHeaders:   reqData.HTTPHeaders,
		HTTPBody:      reqData.HTTPBody,
		LogTextShort:  "Requête interceptée",
		LogText:       "Requête interceptée: " + reqData.HTTPMethod + " " + reqData.HTTPUrl,
		LogType:       "info",
	}
}

// TestConvertToLogModel vérifie que les données de requête sont correctement converties en modèle de log
func TestConvertToLogModel(t *testing.T) {
	// Créer des données de requête de test
	now := time.Now()
	reqData := RequestData{
		ClientName:    "ServiceA",
		CorrelationID: "abcd-1234",
		User:          "jdoe",
		HTTPMethod:    "POST",
		HTTPUrl:       "http://test.com",
		HTTPHeaders:   map[string]string{"Content-Type": "application/json"},
		HTTPBody:      `{"key":"value"}`,
		OccuredTime:   now,
	}

	// Convertir en modèle de log
	logEntry := convertToLogModel(reqData)

	// Vérifier les champs du modèle de log
	assert.Equal(t, "ServiceA", logEntry.ClientName, "Le nom du client doit être correctement transféré")
	assert.Equal(t, "abcd-1234", logEntry.CorrelationID, "L'ID de corrélation doit être correctement transféré")
	assert.Equal(t, "jdoe", logEntry.User, "L'utilisateur doit être correctement transféré")
	assert.Equal(t, "POST", logEntry.HTTPMethod, "La méthode HTTP doit être correctement transférée")
	assert.Equal(t, "http://test.com", logEntry.HTTPUrl, "L'URL doit être correctement transférée")
	assert.Equal(t, "application/json", logEntry.HTTPHeaders["Content-Type"], "Les en-têtes doivent être correctement transférés")
	assert.Equal(t, `{"key":"value"}`, logEntry.HTTPBody, "Le corps de la requête doit être correctement transféré")
	assert.Equal(t, now, logEntry.OccuredTime, "L'heure d'occurrence doit être correctement transférée")
	assert.Equal(t, "Requête interceptée", logEntry.LogTextShort, "Le texte court du log doit être correctement défini")
	assert.Equal(t, "Requête interceptée: POST http://test.com", logEntry.LogText, "Le texte du log doit être correctement défini")
	assert.Equal(t, "info", logEntry.LogType, "Le type de log doit être 'info' par défaut")
}
