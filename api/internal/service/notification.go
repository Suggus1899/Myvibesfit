package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"myvibesfit/api/internal/domain"
)

type NotificationService struct {
	tokens domain.DeviceTokenRepository
	sender domain.PushSender
	logger *slog.Logger
}

func NewNotificationService(tokens domain.DeviceTokenRepository, sender domain.PushSender, logger *slog.Logger) *NotificationService {
	return &NotificationService{tokens: tokens, sender: sender, logger: logger}
}

func (s *NotificationService) Register(ctx context.Context, userID uuid.UUID, token, platform string) (domain.DeviceToken, error) {
	if token == "" {
		return domain.DeviceToken{}, fmt.Errorf("%w: token vacio", domain.ErrInvalidInput)
	}
	if !domain.ValidPushPlatform(platform) {
		return domain.DeviceToken{}, fmt.Errorf("%w: platform debe ser ios o android", domain.ErrInvalidInput)
	}
	return s.tokens.Upsert(ctx, userID, token, platform)
}

func (s *NotificationService) Unregister(ctx context.Context, userID uuid.UUID, token string) error {
	if token == "" {
		return fmt.Errorf("%w: token vacio", domain.ErrInvalidInput)
	}
	// Un token que ya no esta (o que migro a otra cuenta) no es un error para
	// quien cierra sesion: el objetivo es que deje de recibir, y ya no recibe.
	if _, err := s.tokens.Delete(ctx, userID, token); err != nil {
		return err
	}
	return nil
}

// NotifyUser no devuelve error a proposito: notificar es un efecto colateral
// de la operacion de negocio, nunca su condicion de exito. Asignar un
// programa no puede fallar porque el proveedor de push este caido.
func (s *NotificationService) NotifyUser(ctx context.Context, userID uuid.UUID, n domain.Notification) {
	tokens, err := s.tokens.ListByUser(ctx, userID)
	if err != nil {
		s.logger.Error("no se pudieron leer los device tokens", "user_id", userID, "error", err)
		return
	}
	if len(tokens) == 0 {
		return
	}

	raw := make([]string, 0, len(tokens))
	for _, t := range tokens {
		raw = append(raw, t.Token)
	}

	dead, err := s.sender.Send(ctx, raw, n)
	if err != nil {
		s.logger.Error("fallo el envio de push", "user_id", userID, "error", err)
		// Sin return: los tokens que el proveedor alcanzo a marcar como
		// muertos se limpian igual, aunque el envio haya fallado en conjunto.
	}
	for _, t := range dead {
		if _, err := s.tokens.Delete(ctx, userID, t); err != nil {
			s.logger.Error("no se pudo borrar un device token muerto", "error", err)
		}
	}
}
