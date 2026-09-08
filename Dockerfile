FROM redis:7-alpine

# Copiar .env para dentro do container (opcional)
COPY .env /tmp/.env

# Expor porta
EXPOSE 6379

# Comando para iniciar com senha do .env
CMD ["sh", "-c", "redis-server --appendonly yes --requirepass $(grep REDIS_PASSWORD /tmp/.env | cut -d '=' -f2)"]