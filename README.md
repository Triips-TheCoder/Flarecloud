## 👤️ Authors 👤

- Nicolas Marsan ([@NicoooM](https://github.com/NicoooM))<br />
- Benoît Favrie ([@benoitfvr](https://github.com/benoitfvr))<br />
- Julian Laballe ([@Triips-TheCoder](https://github.com/Triips-TheCoder))<br />
- Lucas ([@lucasboucher](https://github.com/lucasboucher))<br />
- Paul Mazeau ([@PaulMazeau](https://github.com/PaulMazeau))

# 🚀 GO CDN project 🚀

**Objectif principal :**
Concevoir un Content Delivery Network (CDN) performant en utilisant Go, en appliquant les méthodologies Agile et en réalisant des tests automatisés.

## Run the project

1. Run Docker

```bash
docker-compose up -d
```

2. Install the dependencies

```bash
go mod tidy
```

3. Run the project

```bash
go run .
```

App is running on [`http://localhost:8080/`](http://localhost:8080/) and Minio is running on [`http://localhost:9001/`](http://localhost:9001/)

## Run with Kubernetes

1. **Start Minikube and add Ingress:**
   ```bash
   minikube start --driver=docker
   minikube addons enable ingress
   ```
2. **Build the backend image and load it to Minikube:**
   ```bash
    docker build -t go-backend .
    minikube image load go-backend
   ```
3. **Create secrets based on local env:**
   ```bash
    cd k8s
    kubectl create secret generic go-backend-secrets --from-env-file=.env.backend
    kubectl create secret generic mongodb-secrets --from-env-file=.env.mongo
   ```
4. **Apply the configs:**
   ```bash
    kubectl apply -f deployment.yml
    kubectl apply -f service.yml
    kubectl apply -f ingress.yml
   ```
5. **Access the service:**
   ```bash
    # Get the Minikube IP
    minikube ip
    # Test the app
    curl http://<minikube_ip>/health
   ```

## Fonctionnalités

- **Gestion des utilisateurs**
- **Gestion des fichiers :**
  - Upload
  - Download
  - Delete
  - Update
- **Rate Limiter (captcha)**
- **Gestion des logs**
- **CI (Github Actions): Tests et Linting**
- **CD (Github Actions): Build de l'image et push sur DockerHub**
