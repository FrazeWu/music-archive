# Kanye West Archive - Deployment Guide

This guide helps you deploy the Go + Vue.js application on ARM Debian using Docker Compose, with Cloudflare for CDN and SSL.

## Architecture
- Frontend: Vue.js SPA served by nginx
- Backend: Go API server
- Database: External MySQL (production) or SQLite (development)
- Reverse Proxy: nginx for API proxying
- CDN/SSL: Cloudflare

## Prerequisites
- ARM Debian server (e.g., Raspberry Pi 4, AWS Graviton, etc.)
- Docker and Docker Compose installed
- Domain name
- MySQL database (external, for production)

## Quick Start

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd all_kanye
   ```

2. **Configure environment**
   Edit `server/.env` with your secrets:
   ```
   ENV=production
   DATABASE_TYPE=mysql
   DATABASE_URL=user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
   JWT_SECRET=your-very-secure-secret-key
   PORT=8080
   ```

   For development (SQLite):
   ```
   ENV=development
   DATABASE_TYPE=sqlite
   DATABASE_URL=./database.sqlite
   ```

   **Note:** For production, set up an external MySQL database separately and update the DATABASE_URL accordingly.

3. **Build and run**
   ```bash
   docker-compose up -d --build
   ```

4. **Check status**
   ```bash
   docker-compose ps
   docker-compose logs
   ```

## Cloudflare Configuration

1. **Sign up** at [cloudflare.com](https://cloudflare.com)

2. **Add your domain**
   - Enter your domain name
   - Cloudflare will scan DNS records

3. **Update DNS**
   - Create A record pointing to your server IP
   - Enable the orange cloud (CDN enabled)

4. **SSL/TLS Settings**
   - SSL/TLS > Overview > Full (strict)
   - Always Use HTTPS > On

5. **Performance**
   - Speed > Optimization > Auto Minify > On for HTML, CSS, JS
   - Caching > Browser Cache TTL > Respect Existing Headers

## Production Considerations

- **Security**: Change default JWT secret
- **Backups**: Regularly backup `server/database.sqlite`
- **Updates**: Use `docker-compose pull && docker-compose up -d` for updates
- **Monitoring**: Check logs with `docker-compose logs -f`
- **Scaling**: For high traffic, consider separate database server

## Troubleshooting

- **Port conflicts**: Ensure ports 80 and 8080 are free
- **Database issues**: Check file permissions for database.sqlite
- **Build failures**: Ensure ARM64 architecture is supported

## File Structure

- `web/`: Vue.js frontend
- `server/`: Go backend
- `docker-compose.yml`: Container orchestration
- `DEPLOYMENT.md`: This file

Enjoy your Kanye West music archive!