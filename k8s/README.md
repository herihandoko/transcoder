# Kubernetes Deployment for Linier Channel

This directory contains Kubernetes manifests for deploying the Linier Channel video transcoding service.

## Prerequisites

- Kubernetes cluster (v1.20+)
- kubectl configured
- NFS or compatible storage class
- NGINX Ingress Controller (optional)

## Deployment Order

Deploy the manifests in the following order:

```bash
# 1. Create namespace
kubectl apply -f namespace.yaml

# 2. Create secrets and configmaps
kubectl apply -f secret.yaml
kubectl apply -f configmap.yaml

# 3. Create persistent volume claims
kubectl apply -f pvc.yaml
kubectl apply -f mysql-deployment.yaml
kubectl apply -f redis-deployment.yaml

# 4. Wait for MySQL and Redis to be ready
kubectl wait --for=condition=ready pod -l app=mysql -n linier-channel
kubectl wait --for=condition=ready pod -l app=redis -n linier-channel

# 5. Deploy the main application
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml

# 6. Deploy ingress (optional)
kubectl apply -f ingress.yaml

# 7. Deploy HPA (optional)
kubectl apply -f hpa.yaml
```

## Configuration

### Environment Variables

The application is configured through environment variables:

- **Database**: MySQL connection settings
- **Redis**: Redis connection settings  
- **Storage**: File upload and transcoded video paths
- **FFmpeg**: FFmpeg binary paths
- **Transcoding**: Worker count and HLS settings

### Secrets

Update the secrets in `secret.yaml`:

```bash
# Encode your password
echo -n "your-password" | base64

# Update the secret
kubectl apply -f secret.yaml
```

### Storage

The application requires persistent storage for:
- Uploaded videos (`/uploads`)
- Transcoded videos (`/transcoded-videos`)
- MySQL data
- Redis data

Adjust the `storageClassName` in the PVC manifests based on your cluster.

## Monitoring

### Health Checks

The application provides health check endpoints:
- `/health` - Application health status

### Logs

View application logs:

```bash
kubectl logs -f deployment/linier-channel -n linier-channel
```

### Scaling

The HPA will automatically scale based on CPU and memory usage:
- Min replicas: 2
- Max replicas: 10
- CPU target: 70%
- Memory target: 80%

## API Endpoints

Once deployed, the service will be available at:

- **Local**: `http://linier-channel.local` (if using ingress)
- **Cluster**: `http://linier-channel-service.linier-channel.svc.cluster.local:8080`

### Key Endpoints

- `POST /api/v1/videos/upload` - Upload video
- `GET /api/v1/videos/:id/status` - Get video status
- `GET /api/v1/stream/:videoId/master.m3u8` - HLS master playlist
- `GET /api/v1/stream/:videoId/:resolution/playlist.m3u8` - HLS playlist
- `GET /api/v1/stream/:videoId/:resolution/:segment` - Video segment

## Troubleshooting

### Common Issues

1. **Storage Issues**
   - Check PVC status: `kubectl get pvc -n linier-channel`
   - Verify storage class: `kubectl get storageclass`

2. **Database Connection**
   - Check MySQL pod: `kubectl get pods -l app=mysql -n linier-channel`
   - Check MySQL logs: `kubectl logs -l app=mysql -n linier-channel`

3. **Redis Connection**
   - Check Redis pod: `kubectl get pods -l app=redis -n linier-channel`
   - Check Redis logs: `kubectl logs -l app=redis -n linier-channel`

4. **Application Issues**
   - Check pod status: `kubectl get pods -l app=linier-channel -n linier-channel`
   - Check logs: `kubectl logs -l app=linier-channel -n linier-channel`

### Cleanup

To remove the deployment:

```bash
kubectl delete namespace linier-channel
```

## Customization

### Resource Limits

Adjust resource requests and limits in `deployment.yaml`:

```yaml
resources:
  requests:
    memory: "512Mi"
    cpu: "500m"
  limits:
    memory: "2Gi"
    cpu: "2000m"
```

### Scaling

Modify HPA settings in `hpa.yaml`:

```yaml
minReplicas: 2
maxReplicas: 10
```

### Storage

Adjust storage sizes in PVC manifests:

```yaml
resources:
  requests:
    storage: 100Gi
```
