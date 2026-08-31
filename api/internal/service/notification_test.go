package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"

	"myvibesfit/api/internal/domain"
)

type fakeDeviceTokens struct {
	domain.DeviceTokenRepository

	tokens  []domain.DeviceToken
	deleted []string
	listErr error
}

func (f *fakeDeviceTokens) ListByUser(context.Context, uuid.UUID) ([]domain.DeviceToken, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.tokens, nil
}

func (f *fakeDeviceTokens) Delete(_ context.Context, _ uuid.UUID, token string) (bool, error) {
	f.deleted = append(f.deleted, token)
	return true, nil
}

func (f *fakeDeviceTokens) Upsert(_ context.Context, userID uuid.UUID, token, platform string) (domain.DeviceToken, error) {
	return domain.DeviceToken{UserID: userID, Token: token, Platform: platform}, nil
}

type fakeSender struct {
	dead   []string
	err    error
	calls  int
	gotAll []string
}

func (f *fakeSender) Send(_ context.Context, tokens []string, _ domain.Notification) ([]string, error) {
	f.calls++
	f.gotAll = append(f.gotAll, tokens...)
	return f.dead, f.err
}

func quietLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestNotifyUserBorraLosTokensMuertos(t *testing.T) {
	repo := &fakeDeviceTokens{tokens: []domain.DeviceToken{
		{Token: "vivo", Platform: domain.PlatformAndroid},
		{Token: "muerto", Platform: domain.PlatformIOS},
	}}
	sender := &fakeSender{dead: []string{"muerto"}}
	svc := NewNotificationService(repo, sender, quietLogger())

	svc.NotifyUser(context.Background(), uuid.New(), domain.PlanAssignedNotification("Full Body"))

	if len(repo.deleted) != 1 || repo.deleted[0] != "muerto" {
		t.Fatalf("esperaba que borrara solo el token muerto, borro: %v", repo.deleted)
	}
	if len(sender.gotAll) != 2 {
		t.Fatalf("esperaba que enviara a los 2 tokens, envio a %d", len(sender.gotAll))
	}
}

func TestNotifyUserSinTokensNoLlamaAlProveedor(t *testing.T) {
	repo := &fakeDeviceTokens{}
	sender := &fakeSender{}
	svc := NewNotificationService(repo, sender, quietLogger())

	svc.NotifyUser(context.Background(), uuid.New(), domain.StreakAtRiskNotification(4))

	if sender.calls != 0 {
		t.Fatalf("sin tokens registrados no deberia llamar al proveedor, llamo %d veces", sender.calls)
	}
}

// El envio es un efecto colateral: que el proveedor falle no puede convertirse
// en un error de la operacion de negocio que lo disparo.
func TestNotifyUserTragaElErrorDelProveedor(t *testing.T) {
	repo := &fakeDeviceTokens{tokens: []domain.DeviceToken{{Token: "t1"}}}
	sender := &fakeSender{err: errors.New("fcm caido")}
	svc := NewNotificationService(repo, sender, quietLogger())

	svc.NotifyUser(context.Background(), uuid.New(), domain.PlanAssignedNotification("Push Pull"))

	if sender.calls != 1 {
		t.Fatalf("esperaba 1 intento de envio, hubo %d", sender.calls)
	}
}

func TestNotifyUserConErrorDeLecturaNoEnvia(t *testing.T) {
	repo := &fakeDeviceTokens{listErr: errors.New("db caida")}
	sender := &fakeSender{}
	svc := NewNotificationService(repo, sender, quietLogger())

	svc.NotifyUser(context.Background(), uuid.New(), domain.PlanAssignedNotification("Upper"))

	if sender.calls != 0 {
		t.Fatalf("si no pudo leer los tokens no deberia enviar, llamo %d veces", sender.calls)
	}
}

func TestRegisterValidaPlataformaYToken(t *testing.T) {
	svc := NewNotificationService(&fakeDeviceTokens{}, &fakeSender{}, quietLogger())

	cases := []struct {
		name     string
		token    string
		platform string
		wantErr  bool
	}{
		{"android valido", "abc", domain.PlatformAndroid, false},
		{"ios valido", "abc", domain.PlatformIOS, false},
		{"web rechazado por el CHECK de la tabla", "abc", "web", true},
		{"plataforma vacia", "abc", "", true},
		{"token vacio", "", domain.PlatformIOS, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.Register(context.Background(), uuid.New(), tc.token, tc.platform)
			if tc.wantErr && !errors.Is(err, domain.ErrInvalidInput) {
				t.Fatalf("esperaba ErrInvalidInput, obtuve %v", err)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("no esperaba error, obtuve %v", err)
			}
		})
	}
}
