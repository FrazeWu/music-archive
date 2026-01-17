# Deployment Guide with Cloudflare

This guide covers deploying the Kanye Archive application with Cloudflare as CDN and SSL provider.

## Prerequisites

- Docker and Docker Compose installed
- Domain pointed to your server via Cloudflare DNS
- Cloudflare account with your domain added

## Cloudflare Configuration

### 1. DNS Setup

In Cloudflare Dashboard → DNS:

```
Type: A
Name: @ (or subdomain)
IPv4: your-server-ip
Proxy status: Proxied (orange cloud)
```

### 2. SSL/TLS Settings

Cloudflare Dashboard → SSL/TLS:

#### Option A: Flexible SSL (Quick Setup)
- **SSL/TLS encryption mode**: Flexible
- Cloudflare ↔ User: HTTPS
- Origin ↔ Cloudflare: HTTP
- **No certificates needed on origin server**
- Current nginx configuration supports this

#### Option B: Full SSL (Recommended)
- **SSL/TLS encryption mode**: Full
- Requires Cloudflare Origin Certificate
- More secure than Flexible

**To set up Full SSL:**

1. Go to SSL/TLS → Origin Server
2. Click "Create Certificate"
3. Select:
   - Key type: RSA (2048)
   - Hostnames: yourdomain.com, *.yourdomain.com
   - Validity: 15 years
4. Click "Create"
5. Copy the certificate and key

Save on your server:
```bash
# On your server
cd /path/to/all_kanye/nginx/ssl
nano cert.pem  # Paste certificate
nano key.pem   # Paste private key
chmod 600 key.pem
```

Update nginx configuration:
```bash
# Use SSL config
cp nginx/nginx-ssl.conf web/nginx.conf
```

Rebuild containers:
```bash
docker-compose down
docker-compose up --build -d
```

#### Option C: Full (Strict)
- Requires valid SSL certificate from trusted CA
- Use Let's Encrypt or other CA

### 3. Cloudflare Settings Optimization

#### Speed Settings (Speed → Optimization)
- Auto Minify: ✓ JavaScript, ✓ CSS, ✓ HTML
- Brotli: ✓ Enabled
- Early Hints: ✓ Enabled (optional)

#### Caching Settings (Caching → Configuration)
- Caching Level: Standard
- Browser Cache TTL: Respect Existing Headers

**Page Rules** (optional):
```
URL: yourdomain.com/api/*
Settings:
- Cache Level: Bypass
- Disable Performance

URL: yourdomain.com/*
Settings:
- Browser Cache TTL: 4 hours
- Cache Level: Standard
```

#### Security Settings
- Security Level: Medium
- Bot Fight Mode: ✓ Enabled
- Challenge Passage: 30 minutes

#### Firewall Rules (optional)
Protect admin endpoints:
```
Expression: (http.request.uri.path contains "/api/admin")
Action: JS Challenge
```

## Server Setup

### 1. Clone Repository
```bash
git clone git@github.com:FrazeWu/musci-archive.git
cd musci-archive
```

### 2. Configure Environment
```bash
cd server
cp .env.example .env.production
nano .env.production
```

Update with your values:
```env
ENV=production
DATABASE_TYPE=mysql
DATABASE_URL=user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local

PORT=8080
GIN_MODE=release

JWT_SECRET=your-generated-secret

# S3 Configuration
AWS_ACCESS_KEY_ID=your-key
AWS_SECRET_ACCESS_KEY=your-secret
S3_BUCKET=your-bucket
S3_REGION=your-region
S3_ENDPOINT=https://your-endpoint
S3_URL_PREFIX=https://your-public-url
```

### 3. Start Services
```bash
cd ..
docker-compose up -d
```

### 4. Verify Deployment
```bash
# Check logs
docker-compose logs -f

# Check health
curl http://localhost/api/songs
```

### 5. Create Admin User
```bash
docker-compose exec backend /app/main create-admin
# Or manually:
docker-compose exec backend sh
cd /app
./main create-admin
```

## Post-Deployment

### Monitor Services
```bash
# View logs
docker-compose logs -f

# View specific service
docker-compose logs -f frontend
docker-compose logs -f backend

# Check status
docker-compose ps
```

### Update Application
```bash
git pull origin main
docker-compose down
docker-compose up --build -d
```

### Backup Database
```bash
# MySQL backup
docker-compose exec backend sh -c 'mysqldump -h $DB_HOST -u $DB_USER -p$DB_PASSWORD $DB_NAME > /app/backup.sql'
docker cp $(docker-compose ps -q backend):/app/backup.sql ./backup-$(date +%Y%m%d).sql
```

## Cloudflare Performance Tips

### Cache Everything with Page Rules
```
URL: yourdomain.com/assets/*
Settings:
- Cache Level: Cache Everything
- Edge Cache TTL: 1 month
```

### Enable Argo Smart Routing (Paid)
- Speeds up dynamic content
- ~$5/month + $0.10/GB

### Enable Polish (Image Optimization)
- Lossy or Lossless compression
- WebP conversion

### Enable Zaraz (Tag Management)
- Load third-party scripts faster

## Troubleshooting

### SSL Issues
```bash
# Check SSL certificate
openssl s_client -connect yourdomain.com:443 -servername yourdomain.com

# Verify Cloudflare is proxying
curl -I https://yourdomain.com
# Look for: cf-ray, cf-cache-status headers
```

### Origin Connection Issues
- Check Cloudflare → SSL/TLS → Overview
- Verify encryption mode matches your setup
- Check firewall allows Cloudflare IPs

### API Not Working
```bash
# Test backend directly
curl http://localhost:8080/api/songs

# Check nginx proxy
docker-compose logs frontend | grep api
```

### Upload Issues
- Check `client_max_body_size` in nginx.conf
- Verify S3 credentials
- Check backend logs

## Security Checklist

- [ ] Change default JWT_SECRET
- [ ] Use Full SSL or Full (Strict) mode
- [ ] Enable Cloudflare firewall rules
- [ ] Set up rate limiting
- [ ] Enable bot protection
- [ ] Regular security updates
- [ ] Database backups scheduled
- [ ] Monitor error logs
- [ ] Use strong database password
- [ ] Restrict admin access by IP (optional)

## Performance Checklist

- [ ] Enable Cloudflare Auto Minify
- [ ] Enable Brotli compression
- [ ] Set up Page Rules for caching
- [ ] Enable WebP conversion
- [ ] Configure CDN caching
- [ ] Monitor response times
- [ ] Optimize database queries
- [ ] Use connection pooling

## Monitoring

### Cloudflare Analytics
- Traffic overview
- Performance metrics
- Security events
- Cache hit rate

### Application Logs
```bash
# Real-time monitoring
docker-compose logs -f --tail=100

# Search for errors
docker-compose logs backend | grep ERROR
```

### Health Checks
```bash
# Backend health
curl http://localhost:8080/api/songs

# Database connection
docker-compose exec backend sh -c 'echo "SELECT 1" | mysql -h $DB_HOST -u $DB_USER -p$DB_PASSWORD $DB_NAME'
```

## Resources

- [Cloudflare SSL Modes](https://developers.cloudflare.com/ssl/origin-configuration/ssl-modes/)
- [Cloudflare Origin Certificates](https://developers.cloudflare.com/ssl/origin-configuration/origin-ca/)
- [Nginx Optimization](https://www.nginx.com/blog/tuning-nginx/)
- [Docker Compose Best Practices](https://docs.docker.com/compose/production/)
