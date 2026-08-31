package push

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"myvibesfit/api/internal/domain"
)

// fcmScope es el unico permiso que necesita el service account.
const fcmScope = "https://www.googleapis.com/auth/firebase.messaging"

// errTokenDead marca los tokens que FCM reporta como invalidos. Se traduce a
// la lista `dead` del port para que el servicio los borre.
var errTokenDead = errors.New("device token invalido")

// FCMSender habla el HTTP v1 de FCM directamente. Se evito firebase-admin-go
// a proposito: para enviar un mensaje arrastraba grpc, protobuf, genproto y
// appengine, cuando esto es un POST con un bearer token.
type FCMSender struct {
	projectID string
	client    *http.Client
	baseURL   string
	logger    *slog.Logger
}

var _ domain.PushSender = (*FCMSender)(nil)

func NewFCMSender(ctx context.Context, credentialsJSON []byte, projectID string, logger *slog.Logger) (*FCMSender, error) {
	creds, err := google.CredentialsFromJSON(ctx, credentialsJSON, fcmScope)
	if err != nil {
		return nil, fmt.Errorf("credenciales de FCM invalidas: %w", err)
	}
	if projectID == "" {
		projectID = creds.ProjectID
	}
	if projectID == "" {
		return nil, errors.New("falta el project id de FCM: no vino en las credenciales ni en FCM_PROJECT_ID")
	}

	logger.Info("push notifications activas", "project_id", projectID)
	return &FCMSender{
		projectID: projectID,
		client:    oauth2.NewClient(ctx, creds.TokenSource),
		baseURL:   "https://fcm.googleapis.com",
		logger:    logger,
	}, nil
}

// Send envia un mensaje por token. El HTTP v1 no tiene multicast (el SDK
// oficial lo simula agrupando del lado del cliente), y un usuario tiene un
// punado de dispositivos, asi que el bucle alcanza.
func (s *FCMSender) Send(ctx context.Context, tokens []string, n domain.Notification) ([]string, error) {
	var dead []string
	var firstErr error

	for _, t := range tokens {
		err := s.sendOne(ctx, t, n)
		switch {
		case err == nil:
		case errors.Is(err, errTokenDead):
			dead = append(dead, t)
		case firstErr == nil:
			// Se sigue con el resto de los dispositivos: que un telefono falle
			// no es razon para no avisarle a los otros.
			firstErr = err
		}
	}
	return dead, firstErr
}

type fcmMessage struct {
	Message struct {
		Token        string            `json:"token"`
		Notification fcmNotification   `json:"notification"`
		Data         map[string]string `json:"data,omitempty"`
	} `json:"message"`
}

type fcmNotification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type fcmErrorResponse struct {
	Error struct {
		Code    int    `json:"code"`
		Status  string `json:"status"`
		Message string `json:"message"`
		Details []struct {
			ErrorCode string `json:"errorCode"`
		} `json:"details"`
	} `json:"error"`
}

func (s *FCMSender) sendOne(ctx context.Context, token string, n domain.Notification) error {
	var msg fcmMessage
	msg.Message.Token = token
	msg.Message.Notification = fcmNotification{Title: n.Title, Body: n.Body}
	msg.Message.Data = n.Data

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal del mensaje: %w", err)
	}

	url := fmt.Sprintf("%s/v1/projects/%s/messages:send", s.baseURL, s.projectID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("envio a FCM: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	var fcmErr fcmErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&fcmErr); err != nil {
		return fmt.Errorf("FCM respondio %d con un cuerpo ilegible", resp.StatusCode)
	}
	if isDeadTokenError(resp.StatusCode, fcmErr) {
		return errTokenDead
	}
	return fmt.Errorf("FCM respondio %d (%s): %s", resp.StatusCode, fcmErr.Error.Status, fcmErr.Error.Message)
}

// isDeadTokenError distingue "este token ya no sirve" de "fallo el envio".
// Solo el primero justifica borrar la fila: tratar un 503 de FCM como token
// muerto desregistraria dispositivos sanos en cada caida del proveedor.
func isDeadTokenError(status int, e fcmErrorResponse) bool {
	for _, d := range e.Error.Details {
		if d.ErrorCode == "UNREGISTERED" || d.ErrorCode == "INVALID_ARGUMENT" {
			return true
		}
	}
	return status == http.StatusNotFound
}
