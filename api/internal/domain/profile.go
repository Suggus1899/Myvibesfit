package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const (
	DefaultBiologicalSex = "unspecified"
	UnitSystemMetric     = "metric"
	UnitSystemImperial   = "imperial"
)

// ClientProfile es el onboarding del cliente: define en que unidades ve la
// app y da a la IA el contexto (objetivo, equipamiento, limitaciones) que
// sin esto tiene que adivinar.
type ClientProfile struct {
	UserID             uuid.UUID
	BirthDate          *time.Time
	Sex                string
	HeightCm           *float64
	Experience         string
	PrimaryGoal        string
	DaysPerWeek        int16
	SessionMinutes     int16
	AvailableEquipment []string
	Limitations        string
	UnitSystem         string
	OnboardedAt        *time.Time
}

// IsOnboarded distingue un perfil guardado de uno que nunca se completo: la
// app usa esto para decidir si mostrar el flujo de alta.
func (p ClientProfile) IsOnboarded() bool { return p.OnboardedAt != nil }

type ProfileRepository interface {
	Get(ctx context.Context, userID uuid.UUID) (ClientProfile, error)
	Upsert(ctx context.Context, p ClientProfile) (ClientProfile, error)
}
