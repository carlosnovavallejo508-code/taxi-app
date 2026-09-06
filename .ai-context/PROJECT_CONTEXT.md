# 📋 CONTEXTO DEL PROYECTO - TAXI APP

## ESTADO ACTUAL
- Fase: Inicial
- Progreso: 0%
- Última actualización: $(date)

## ARQUITECTURA
- Backend: Go con microservicios
- Frontend: React + TypeScript
- Base de datos: PostgreSQL
- Cache: Redis
- API: REST + WebSocket

## ESTRUCTURA
taxi-app/
├── backend/
│ ├── cmd/
│ ├── internal/
│ ├── pkg/
│ └── migrations/
├── frontend/
│ ├── passenger-app/
│ ├── driver-app/
│ ├── admin-dashboard/
│ └── business-portal/
├── docs/
├── scripts/
└── docker/

## PRÓXIMOS PASOS
1. Crear go.mod
2. Configurar API Gateway
3. Crear modelos de datos
4. Implementar endpoints básicos

## CONVENCIONES
- Go: CamelCase
- TypeScript: PascalCase para interfaces
- SQL: snake_case
- API: RESTful con /api/v1/
