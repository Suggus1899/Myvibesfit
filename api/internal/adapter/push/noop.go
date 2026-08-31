// Package push implementa domain.PushSender. NoopSender es el que se cablea
// cuando no hay credenciales del proveedor, para que el desarrollo local no
// dependa de una cuenta de Firebase.
package push

import (
	"context"
	"log/slog"

	"myvibesfit/api/internal/domain"
)

type NoopSender struct{ logger *slog.Logger }

func NewNoopSender(logger *slog.Logger) *NoopSender { return &NoopSender{logger: logger} }

var _ domain.PushSender = (*NoopSender)(nil)

// Send no envia nada y no devuelve tokens muertos: si devolviera alguno, el
// servicio borraria tokens validos solo por estar corriendo sin credenciales.
func (s *NoopSender) Send(ctx context.Context, tokens []string, n domain.Notification) ([]string, error) {
	s.logger.Info("push deshabilitado, no se envia",
		"titulo", n.Title, "destinatarios", len(tokens))
	return nil, nil
}
