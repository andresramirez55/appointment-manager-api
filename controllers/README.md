# Controladores HTTP

Cada controlador vive en una carpeta que representa su dominio. El archivo
`controller.go` recibe solicitudes HTTP, valida los datos de entrada y delega
la lógica de negocio a su servicio correspondiente.

| Carpeta | Responsabilidad |
| --- | --- |
| `appointment` | Crear, consultar, actualizar y cancelar turnos. |
| `auth` | Registro, acceso y perfil del profesional. |
| `availability` | Horarios de atención y excepciones de disponibilidad. |
| `block` | Bloqueos de agenda. |
| `office` | Alta y administración de lugares de atención. |
| `note` | Notas asociadas a turnos. |
| `patient` | Gestión de pacientes. |
| `public` | Reserva online y acciones disponibles sin autenticación. |

La carpeta identifica el dominio; por eso todos los paquetes exponen el tipo
`Controller` y el constructor `New`.
