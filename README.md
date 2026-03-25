# Gestor de Turnos

Sistema de gestión de turnos para profesionales independientes. Permite administrar pacientes, agenda semanal, turnos recurrentes, bloqueos de horario y reservas online por parte de los pacientes.

Diseñado para cualquier tipo de profesional: psicólogos, médicos, nutricionistas, odontólogos, etc.

---

## Funcionalidades

### Para el profesional

- **Calendario semanal** con vista por día en mobile y semana en desktop
- **Gestión de pacientes** con historial de turnos y notas de sesión
- **Turnos recurrentes** (semanal o quincenal, hasta 24 sesiones)
- **Bloqueo de agenda** para marcar horarios no disponibles
- **Estadísticas** de asistencia, turnos de la semana y pacientes activos
- **Recordatorios automáticos** por email y WhatsApp 24hs antes del turno
- **Perfil editable** (nombre, especialidad, contraseña)
- **Multi-profesional**: cada cuenta es independiente con sus propios pacientes y turnos

### Para el paciente

- **Reserva online** sin necesidad de crear cuenta
- El paciente accede a un link único del profesional, elige fecha y horario disponible, y confirma el turno
- Confirmación por WhatsApp y/o email

### Técnico

- PWA instalable en Android/iOS
- Responsive (mobile y desktop)
- JWT con expiración de 30 días
- Notificaciones por email via [Resend](https://resend.com)
- Notificaciones por WhatsApp via [Evolution API](https://evolution-api.com)

---

## Stack

| Capa | Tecnología |
|------|-----------|
| Backend | Go + Gin + GORM |
| Frontend | React + TypeScript + Vite + Tailwind CSS |
| Base de datos | PostgreSQL |
| Email | Resend |
| WhatsApp | Evolution API |
| Deploy | Railway |

---

## Cómo usar el sistema

### 1. Registro

Cada profesional crea su propia cuenta en `/register` con nombre, email, contraseña y especialidad.

### 2. Configurar disponibilidad

En **Disponibilidad** (sidebar), el profesional configura sus horarios de atención:
- Día de la semana
- Hora de inicio y fin
- Duración de cada turno (30, 45, 60, 90 o 120 min)

Ejemplo: Lunes a viernes de 09:00 a 18:00, turnos de 60 minutos.

### 3. Compartir link de reserva

Una vez configurada la disponibilidad, el profesional copia su link único desde la misma página:

```
https://tuapp.com/book/[id]
```

El paciente abre ese link, elige fecha y horario, y completa nombre y teléfono.

### 4. Gestión de turnos

- Los turnos aparecen en el **calendario** (semana en desktop, día en mobile)
- Se pueden marcar como **completados** o **cancelados** desde el popover
- Se pueden **reprogramar** con una fecha y hora nuevas
- Se pueden crear **turnos recurrentes** desde el modal de nuevo turno

### 5. Pacientes

- Alta de paciente con nombre, teléfono, email y notas generales
- Perfil del paciente con historial de turnos y notas de sesión por turno
- El teléfono debe ingresarse con código de país sin `+` (ej: `5491112345678`)

### 6. Recordatorios

El sistema envía automáticamente un mensaje 24hs antes de cada turno programado.

---

## Variables de entorno

### Backend

| Variable | Descripción | Requerida |
|----------|-------------|-----------|
| `DATABASE_URL` | URL de PostgreSQL | Sí |
| `JWT_SECRET` | Clave para firmar tokens | Sí |
| `PORT` | Puerto del servidor (default: 8080) | No |
| `RESEND_API_KEY` | API key de Resend para emails | No |
| `FROM_EMAIL` | Email remitente (ej: `turnos@tudominio.com`) | No |
| `WHATSAPP_MODE` | `mock` (desarrollo) o `evolution` (producción) | No |
| `EVOLUTION_API_URL` | URL de Evolution API | Solo si `evolution` |
| `EVOLUTION_API_KEY` | API key de Evolution | Solo si `evolution` |
| `EVOLUTION_INSTANCE_NAME` | Nombre de instancia Evolution | Solo si `evolution` |

### Frontend

| Variable | Descripción |
|----------|-------------|
| `VITE_API_BASE_URL` | URL del backend (ej: `https://api.tudominio.com/api`) |

---

## Deploy en Railway

### Backend

1. Crear servicio desde el repositorio del backend
2. Agregar PostgreSQL como addon
3. Configurar variables de entorno (Railway provee `DATABASE_URL` automáticamente)
4. El deploy se activa automáticamente en cada push a `main`

### Frontend

1. Crear servicio desde el repositorio del frontend
2. Build command: `npm run build`
3. Start command: `npm run preview`
4. Configurar `VITE_API_BASE_URL` con la URL del backend

---

## API

### Pública (sin autenticación)

| Método | Endpoint | Descripción |
|--------|----------|-------------|
| `GET` | `/api/public/professional/:id` | Info pública del profesional |
| `GET` | `/api/public/slots?professional_id=&date=` | Slots disponibles para una fecha |
| `POST` | `/api/public/appointments` | Crear turno como paciente |

### Autenticada

| Método | Endpoint | Descripción |
|--------|----------|-------------|
| `POST` | `/api/auth/register` | Crear cuenta |
| `POST` | `/api/auth/login` | Login |
| `GET` | `/api/profile` | Perfil del profesional |
| `PUT` | `/api/profile` | Actualizar perfil |
| `PUT` | `/api/profile/password` | Cambiar contraseña |
| `GET` | `/api/appointments` | Listar turnos |
| `POST` | `/api/appointments` | Crear turno |
| `POST` | `/api/appointments/recurring` | Crear turnos recurrentes |
| `PUT` | `/api/appointments/:id` | Actualizar turno |
| `DELETE` | `/api/appointments/:id` | Cancelar turno |
| `GET` | `/api/patients` | Listar pacientes |
| `POST` | `/api/patients` | Crear paciente |
| `GET` | `/api/patients/:id` | Detalle de paciente |
| `PUT` | `/api/patients/:id` | Actualizar paciente |
| `GET` | `/api/notes?appointment_id=` | Notas de un turno |
| `POST` | `/api/notes` | Crear nota |
| `GET` | `/api/availability` | Horarios configurados |
| `POST` | `/api/availability` | Agregar horario |
| `DELETE` | `/api/availability/:id` | Eliminar horario |
| `GET` | `/api/blocks` | Bloqueos de agenda |
| `POST` | `/api/blocks` | Crear bloqueo |
| `DELETE` | `/api/blocks/:id` | Eliminar bloqueo |

---

## Desarrollo local

### Requisitos

- Go 1.21+
- Node.js 18+
- Docker

### Setup

```bash
# Base de datos
docker-compose up -d

# Backend
cd backend
go mod download
go run main.go

# Frontend (en otra terminal)
cd frontend
npm install
npm run dev
```

Backend: `http://localhost:8080`
Frontend: `http://localhost:5173`

---

## Arquitectura

```
Router → Controllers → Services → Repositories → PostgreSQL
```

- **Controllers**: Manejo de request/response HTTP
- **Services**: Lógica de negocio
- **Repositories**: Acceso a datos (GORM)
- **Scheduler**: Corre cada hora, envía recordatorios 24hs antes de cada turno
