#!/bin/bash
# Ждем запуска Grafana
while ! curl -s http://localhost:3000/api/health >/dev/null; do 
  sleep 1
done

# Создаем datasource
curl -X POST -H "Content-Type: application/json" \
  -d '{
    "name": "Prometheus",
    "type": "prometheus",
    "url": "http://prometheus:9090",
    "access": "proxy",
    "isDefault": true
  }' \
  http://localhost:3000/api/datasources

# Импортируем дашборд
curl -X POST -H "Content-Type: application/json" \
  -d "{
    \"dashboard\": $(curl -s https://grafana.com/api/dashboards/1860/revisions/35/download),
    \"overwrite\": true
  }" \
  http://localhost:3000/api/dashboards/import