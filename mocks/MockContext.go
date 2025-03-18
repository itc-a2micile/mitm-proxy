package mocks

import (
	"net/http"
)

// MockContext est une implémentation simulée du contexte de proxy pour les tests
type MockContext struct {
	Req  *http.Request
	Resp *http.Response
}

// GetRequest retourne la requête HTTP
func (m *MockContext) GetRequest() *http.Request {
	return m.Req
}

// GetResponse retourne la réponse HTTP
func (m *MockContext) GetResponse() *http.Response {
	return m.Resp
}

// SetResponse définit la réponse HTTP
func (m *MockContext) SetResponse(resp *http.Response) {
	m.Resp = resp
}
