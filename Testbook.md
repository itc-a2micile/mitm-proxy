# 📖 TestBook – Microservice MITM Logging en Go (TDD)
**Version :** 1.0  
**Date :** 17/03/2025   
**Auteur :** William
---

## **📌 Introduction**
Ce **TestBook** décrit la méthodologie de développement en **Test-Driven Development (TDD)** pour le **microservice MITM Logging en Go**.  

L'objectif est de coder les fonctionnalités dans un **ordre stratégique** pour éviter d'avoir à **réécrire** de grandes parties du code en fin de projet.

Nous utiliserons **Go Testing** (`testing`), **Testify** (`github.com/stretchr/testify/assert`), et **httptest** pour **mock les requêtes HTTP**.

---

## **🛠️ 1. Plan de Test & Développement**
### **1️⃣ Configuration du proxy MITM (`go-mitmproxy`)**
> 📌 **Objectif** : Configurer et tester l’interception des requêtes HTTP/HTTPS.

#### **🔹 Test : Initialisation du MITM Proxy**
✅ **Vérifier que le proxy MITM démarre correctement**  
✅ **Vérifier que les requêtes passent bien par le proxy**

```go
func TestMITMProxyStart(t *testing.T) {
	mitm := MITMProxy{}
	err := mitmproxy.Start(mitm)
	assert.NoError(t, err, "Le proxy MITM doit démarrer sans erreur")
}
```
🔹 Développement associé
- Créer un handler de proxy
- Implémenter la capture des requêtes via go-mitmproxy


### **2️⃣ Extraction des données des requêtes**
> 📌 **Objectif** : Intercepter une requête et en extraire les données.

#### 🔹** Test : Extraction des headers et du body**
✅ **Vérifier que les headers et le body sont bien récupérés**  
✅ **Vérifier que les headers client-name, correlation-id, user ou username sont bien extraits**

```go
func TestExtractRequestData(t *testing.T) {
	req := httptest.NewRequest("POST", "http://test.com", bytes.NewBuffer([]byte(`{"key":"value"}`)))
	req.Header.Set("client-name", "ServiceA")
	req.Header.Set("correlation-id", "abcd-1234")
	req.Header.Set("username", "jdoe")

	data := extractRequestData(req)

	assert.Equal(t, "ServiceA", data.ClientName)
	assert.Equal(t, "abcd-1234", data.CorrelationID)
	assert.Equal(t, "jdoe", data.User)
	assert.Equal(t, `{"key":"value"}`, data.HTTPBody)
}
```
🔹 Développement associé   
- Implémenter une fonction extractRequestData()
- Extraire les headers client-name, correlation-id, user
- Lire et stocker le body de la requête

### **3️⃣ Journalisation initiale de la requête**
> 📌 **Objectif** : Convertir la requête interceptée en un log JSON et l’envoyer au microservice Logger.

#### 🔹 **Test : Conversion de la requête en JSON**
✅ **Vérifier que la struct LogModel est bien remplie**   
✅ **Vérifier que la requête est correctement sérialisée en JSON**

```go
func TestConvertRequestToLogModel(t *testing.T) {
	reqData := RequestData{
		ClientName:    "ServiceA",
		CorrelationID: "abcd-1234",
		User:          "jdoe",
		HTTPMethod:    "POST",
		HTTPUrl:       "http://test.com",
		HTTPHeaders:   map[string]string{"Content-Type": "application/json"},
		HTTPBody:      `{"key":"value"}`,
		OccuredTime:   time.Now(),
	}

	logEntry := convertToLogModel(reqData)

	assert.Equal(t, "ServiceA", logEntry.ClientName)
	assert.Equal(t, "abcd-1234", logEntry.CorrelationID)
	assert.Equal(t, "jdoe", logEntry.User)
	assert.Equal(t, `{"key":"value"}`, logEntry.HTTPBody)
}
```

🔹 Développement associé
- Implémenter convertToLogModel()
- Vérifier la sérialisation JSON
- Assurer que tous les champs du LogModel sont bien remplis

### **4️⃣ Envoi du log au microservice Logger**
> 📌 **Objectif** : Envoyer le log en POST /logs au Logger.

#### 🔹 **Test : Envoi du log et gestion des erreurs** 
✅ **Vérifier que le logger reçoit bien la requête**   
✅ **Vérifier la gestion des erreurs réseau (retry)**

```go
func TestSendLogToLogger(t *testing.T) {
	loggerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer loggerServer.Close()

	logEntry := LogModel{
		ClientName: "ServiceA",
		CorrelationID: "abcd-1234",
		User: "jdoe",
	}

	err := sendLogToLogger(loggerServer.URL, logEntry)
	assert.NoError(t, err)
}
```
🔹 Développement associé
- Implémenter sendLogToLogger()
- Gérer les timeouts et retries

### **5️⃣ Capture et journalisation de la réponse**
> 📌 **Objectif** : Mettre à jour le log avec la réponse HTTP.

#### 🔹 **Test : Mise à jour du log avec la réponse**
✅ **Vérifier que le code HTTP et le body sont bien enregistrés**
✅ **Vérifier que le log est mis à jour avec PUT /logs/{id}**

```go
func TestUpdateLogWithResponse(t *testing.T) {
	logEntry := LogModel{
		ClientName: "ServiceA",
		CorrelationID: "abcd-1234",
		User: "jdoe",
	}

	response := http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
		Body:       ioutil.NopCloser(bytes.NewBufferString(`{"message":"success"}`)),
	}

	updateLogWithResponse(&logEntry, &response)

	assert.Equal(t, 200, logEntry.HTTPReturnCode)
	assert.Equal(t, `{"message":"success"}`, logEntry.HTTPReturnBody)
}
```
🔹 Développement associé
- Implémenter updateLogWithResponse()
- Ajouter la gestion des temps d’exécution

## 📌 Plan de Développement

### 🔹 Étapes stratégiques en TDD
- Configurer go-mitmproxy et tester la capture des requêtes
- Extraire les headers et le body des requêtes
- Convertir les requêtes en logs JSON
- Envoyer les logs au Logger
- Capturer et enregistrer la réponse serveur
- Gérer les erreurs, timeouts, et sécurité
- Déployer avec Docker et CI/CD
