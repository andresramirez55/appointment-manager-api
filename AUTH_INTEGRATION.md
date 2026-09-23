# Integración con Auth Service

Contrato verificado contra `andresramirez55/auth-service`, commit
`e0c81fed2bbd212d907b0c17501de1a570ab72d4`. Los endpoints públicos `/health` y
`/.well-known/jwks.json` del deployment respondieron correctamente.

## Responsabilidades

- Auth Service guarda identidad, contraseña y sesiones. Firma access tokens RS256.
- `identity/client.go` implementa el cliente HTTP y valida firma, issuer, expiración,
  tipo de token y subject. Solo acepta RS256; cachea JWKS cinco minutos, limita
  consultas ante claves desconocidas y no acepta claves vencidas si Auth no responde.
- `services/auth_service.go` coordina Auth con el perfil profesional local mediante
  el contrato de repositorio. No accede a GORM.
- `professionals.auth_user_id` es único y corresponde a `sub`. Los pacientes,
  turnos, consultorios y permisos siguen usando el `professional_id` local.
- El middleware exige un perfil vinculado. Un JWT de otro usuario no da acceso
  a los datos de un profesional por compartir email ni por contener su ID numérico.
- El frontend recibe un access token y lo guarda en `sessionStorage` (sigue siendo
  accesible a JavaScript; prevenir XSS continúa siendo necesario). El refresh token
  solo se entrega en una cookie HttpOnly. El backend no firma tokens ni usa JWT_SECRET.
- La renovación rota la cookie y reintenta una vez la petición. Las peticiones
  simultáneas comparten una renovación; Web Locks serializa la rotación entre pestañas
  en navegadores compatibles. Cerrar sesión revoca el refresh token en Auth.
  El access token emitido sigue vigente hasta su expiración, según el contrato de Auth.
- La PWA no cachea respuestas de API ni peticiones autenticadas. El backend marca
  respuestas protegidas y de autenticación con `Cache-Control: no-store`.

## Railway: configuración necesaria antes del deploy

En el **backend de turnos**:

```dotenv
AUTH_SERVICE_URL=https://auth-service-production-8529.up.railway.app
AUTH_ISSUER=https://auth-service-production-8529.up.railway.app
FRONTEND_URL=https://appointment-manager-web-production.up.railway.app
AUTH_COOKIE_SECURE=true
```

Mantener `DATABASE_URL` y las variables de notificaciones. `JWT_SECRET` ya no se usa.
El usuario confirmó que el issuer desplegado es el indicado arriba.
No copiar la clave privada RSA a esta aplicación.

En el **frontend**:

```dotenv
VITE_API_BASE_URL=/api
BACKEND_URL=https://appointment-manager-api-production.up.railway.app
```

`BACKEND_URL` es el origen del backend de turnos, **no** Auth Service, sin `/api`.
La URL se verificó en el bundle público del frontend desplegado.
El build debe regenerarse al cambiar `VITE_API_BASE_URL`. Usar `node server.js`
(como ya indica `railway.toml`), ya que ese servidor envía `/api` al backend.
Este proxy conserva las cookies en el origen del frontend y evita bloqueos de
cookies de terceros entre los dominios Railway. No usar `vite preview` en producción.

Configurar variables, desplegar backend y frontend y comprobar:
registro con contraseña de al menos 12 caracteres, login, perfil, datos propios,
renovación al expirar el access token y logout. Los tests locales usan proveedores
simulados; no se crearon cuentas ni se modificaron variables o datos en producción.

## Desarrollo local

Copiar el `.env.example` raíz a `backend/.env`, con `FRONTEND_URL=http://localhost:5173`
y `AUTH_COOKIE_SECURE=false`. Apuntar Auth a una instancia de desarrollo para no
crear cuentas de prueba en producción. Vite ya tiene proxy `/api` a localhost:8080.
Para probar `node server.js` localmente, usar `BACKEND_URL=http://localhost:8080` y
ajustar `FRONTEND_URL` al origen utilizado (por ejemplo localhost:4173).

## Datos previos y limitaciones

El usuario confirmó que todavía no hay usuarios de la app: no se necesita migrar
accesos. AutoMigrate agrega `auth_user_id` nullable con índice único y conserva
los datos anteriores. Se eliminó la creación automática de `admin@test.com/admin123`.
Si esa cuenta de prueba ya existe, permanece sin acceso: sus credenciales locales
ya no se aceptan. No se borró ningún dato ni se vinculó ninguna cuenta por email.

Si se decide rescatar un perfil de pruebas, verificar la identidad y asignar su
UUID de Auth a `professionals.auth_user_id` administrativamente. No registrar otra
cuenta con ese email esperando que adopte sus datos. El campo histórico `password`
permanece por compatibilidad de esquema; no se consulta ni se guarda un hash nuevo.

Si Auth crea la identidad pero falla la escritura del perfil, iniciar sesión de
nuevo permite completar el perfil. Los datos opcionales se pueden editar después.
Una identidad existente en Auth puede iniciar sesión y obtener su perfil nuevo en
esta app, conforme al registro abierto que ya tenía el producto.

Auth Service aún no ofrece cambio/recuperación de contraseña, verificación de email
ni audience por aplicación. La UI informa que el cambio de contraseña no está
disponible, y se retiró el endpoint local que modificaba una contraseña independiente.
No se afirma que estas funciones estén resueltas por esta integración.

El rate limit de esta app es por proceso y no protege el endpoint público directo
de Auth; el proveedor debe mantener sus propios controles para producción.

## Validación local

```bash
cd backend
GOCACHE=/private/tmp/psych-auth-go-cache go test ./...
cd ../frontend
npm run build
node --test api-proxy.test.js
```
