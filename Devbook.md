# 📖 DevBook – Microservice MITM Logging en Go  
**Version :** 1.0  
**Date :** 17/03/2025   
**Auteur :** William
---

## **1️⃣ Introduction**
Ce **DevBook** détaille les fonctionnalités du **microservice MITM Logging**, ainsi que les tâches et sous-tâches nécessaires pour le développement.  

Le microservice intercepte et journalise les **requêtes HTTP et HTTPS** transitant via **mitmproxy** et **go-mitmproxy**, puis envoie ces logs au **microservice Logger**.

---

## **2️⃣ Fonctionnalités Principales**
### **2.1 Interception et capture des requêtes HTTP/HTTPS**
#### **Tâches principales**
- [ ] Mettre en place **go-mitmproxy** pour capturer le trafic réseau.
- [ ] Intercepter les requêtes **HTTP** et **HTTPS**.
- [ ] Extraire les **headers**, le **body** et les **métadonnées** des requêtes.

#### **Sous-tâches**
- [ ] Configurer le proxy MITM pour intercepter tout le trafic.
- [ ] Lire et stocker les headers de la requête.
- [ ] Lire le **corps de la requête** sans l’altérer pour la transmission.
- [ ] Générer un **timestamp** (`OccuredTime`).
- [ ] Extraire les headers spécifiques :  
  - [ ] `client-name`
  - [ ] `correlation-id`
  - [ ] `user` ou `username` (sinon `"Anonyme"`)

---

### **2.2 Journalisation et envoi des logs**
#### **Tâches principales**
- [ ] Convertir la requête interceptée en **objet JSON** (`LogModel`).
- [ ] Envoyer les logs au **microservice Logger** via une **API REST**.

#### **Sous-tâches**
- [ ] Définir le **format JSON** du log.
- [ ] Sérialiser l’objet `LogModel` en JSON.
- [ ] Envoyer une requête `POST` au microservice Logger.
- [ ] Gérer les erreurs de connexion avec des **retries**.

---

### **2.3 Capture et mise à jour des réponses serveur**
#### **Tâches principales**
- [ ] Attendre la réponse du serveur cible.
- [ ] Mettre à jour le log avec les **données de réponse**.

#### **Sous-tâches**
- [ ] Capturer le **code HTTP de retour** (`200`, `500`, etc.).
- [ ] Lire les **headers de réponse**.
- [ ] Lire et enregistrer le **corps de la réponse**.
- [ ] Mesurer et enregistrer le **temps d’exécution** (`ExecutionTime`).
- [ ] Mettre à jour le log existant avec la réponse.

---

### **2.4 Gestion des erreurs et cas spéciaux**
#### **Tâches principales**
- [ ] Gérer les **erreurs réseau et timeouts**.
- [ ] Gérer les **erreurs applicatives** (ex : serveur renvoie `500`).

#### **Sous-tâches**
- [ ] Logger les requêtes échouées avec un code `500` par défaut.
- [ ] Ajouter une gestion des **timeouts** (`408 Request Timeout`).
- [ ] Implémenter **un retry avec backoff** pour les erreurs temporaires.
- [ ] Logger les exceptions Go (`panic`, `nil`, etc.).
- [ ] Différencier les logs en **Info**, **Erreur**, **Critical**.

---

### **2.5 Sécurité et protection des données**
#### **Tâches principales**
- [ ] Chiffrer les communications avec le Logger en **HTTPS**.
- [ ] Implémenter l’authentification avec **JWT**.

#### **Sous-tâches**
- [ ] **Masquer les données sensibles** (`Authorization`, `password`).
- [ ] Vérifier et respecter **RGPD / conformité interne**.
- [ ] Ajouter des **règles de filtrage** pour exclure certaines routes.
- [ ] Implémenter un **système de permissions** (JWT + rôles).
- [ ] Éviter les **fuites de données** en anonymisant les logs sensibles.

---

### **2.6 Déploiement et infrastructure**
#### **Tâches principales**
- [ ] Conteneuriser le service avec **Docker**.
- [ ] Automatiser le déploiement avec **CI/CD**.

#### **Sous-tâches**
- [ ] Créer un **Dockerfile** optimisé.
- [ ] Configurer un **docker-compose** pour l’environnement local.
- [ ] Intégrer **GitHub Actions** ou **GitLab CI/CD**.
- [ ] Surveiller avec **Prometheus + Grafana**.
- [ ] Ajouter un **système de logs centralisés** (`Fluentd`, `ELK`).

---

### **2.7 Documentation et maintenance**
#### **Tâches principales**
- [ ] Écrire un **README** détaillé.
- [ ] Ajouter des **commentaires et documentation GoDoc**.

#### **Sous-tâches**
- [ ] Expliquer comment **installer et configurer le service**.
- [ ] Documenter **l’API Logger** (`POST /logs`, `PUT /logs/{id}`).
- [ ] Rédiger une **checklist de mise en production**.
- [ ] Mettre en place **une veille technologique**.


## **4️⃣ Conclusion**
Ce DevBook fournit une **vue complète** des fonctionnalités et tâches pour **développer, tester et déployer** ce microservice MITM Logging.  

🚀 **Prêt à coder ? Let's Go(lang) !** 🔥