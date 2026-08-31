package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	PlatformIOS     = "ios"
	PlatformAndroid = "android"
)

func ValidPushPlatform(p string) bool { return p == PlatformIOS || p == PlatformAndroid }

// DeviceToken es el token push de un dispositivo. Cuelga de app_user y no de
// la organizacion: el usuario puede cambiar de gimnasio y el telefono sigue
// siendo suyo.
type DeviceToken struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	Token      string
	Platform   string
	LastSeenAt time.Time
	CreatedAt  time.Time
}

type DeviceTokenRepository interface {
	// Upsert resuelve el conflicto por token: es UNIQUE global, asi que
	// cuando otro usuario entra en el mismo telefono la fila cambia de dueno
	// en vez de chocar.
	Upsert(ctx context.Context, userID uuid.UUID, token, platform string) (DeviceToken, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]DeviceToken, error)
	// Delete solo borra el token si sigue siendo de ese usuario: uno que ya
	// migro a otra cuenta no es del que pide la baja.
	Delete(ctx context.Context, userID uuid.UUID, token string) (bool, error)
}

// Notification es lo que el usuario ve. Data lleva el deep-link y nada mas:
// el cuerpo de una push viaja por servidores de terceros, asi que no
// transporta datos del gimnasio ni PII.
type Notification struct {
	Title string
	Body  string
	Data  map[string]string
}

// PushSender es el port hacia el proveedor de push, con la misma forma que
// SuggestionProposer: el domain declara la intencion y el adapter conoce el
// SDK. Devuelve los tokens que el proveedor reporto como muertos para que el
// caller los borre, sin necesidad de un job de limpieza aparte.
type PushSender interface {
	Send(ctx context.Context, tokens []string, n Notification) (dead []string, err error)
}

func PlanAssignedNotification(programName string) Notification {
	return Notification{
		Title: "Tenes un plan nuevo",
		Body:  fmt.Sprintf("Tu coach te asigno %s. Entra para verlo.", programName),
		Data:  map[string]string{"route": "/workout"},
	}
}

func StreakAtRiskNotification(currentDays int) Notification {
	body := "Registra tu habito de hoy para no perder la racha."
	if currentDays > 1 {
		body = fmt.Sprintf("Llevas %d dias seguidos. Registra el de hoy para no cortar la racha.", currentDays)
	}
	return Notification{
		Title: "No cortes la racha",
		Body:  body,
		Data:  map[string]string{"route": "/habits"},
	}
}
