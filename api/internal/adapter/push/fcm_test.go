package push

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"myvibesfit/api/internal/domain"
)

// newTestSender apunta el adapter a un FCM falso. Evita las credenciales
// reales: lo que hay que probar es la traduccion de errores, no OAuth.
func newTestSender(h http.HandlerFunc) (*FCMSender, *httptest.Server) {
	srv := httptest.NewServer(h)
	return &FCMSender{
		projectID: "proyecto-de-prueba",
		client:    srv.Client(),
		baseURL:   srv.URL,
		logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	}, srv
}

func TestSendMarcaMuertoElTokenNoRegistrado(t *testing.T) {
	sender, srv := newTestSender(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if strings.Contains(string(body), `"token":"muerto"`) {
			w.WriteHeader(http.StatusNotFound)
			io.WriteString(w, `{"error":{"code":404,"status":"NOT_FOUND","message":"Requested entity was not found.","details":[{"errorCode":"UNREGISTERED"}]}}`)
			return
		}
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `{"name":"projects/p/messages/1"}`)
	})
	defer srv.Close()

	dead, err := sender.Send(context.Background(), []string{"vivo", "muerto"}, domain.PlanAssignedNotification("Full Body"))
	if err != nil {
		t.Fatalf("un token muerto no es un error de envio: %v", err)
	}
	if len(dead) != 1 || dead[0] != "muerto" {
		t.Fatalf("esperaba solo el token muerto, obtuve %v", dead)
	}
}

// Una caida del proveedor no debe desregistrar dispositivos sanos.
func TestSendNoMarcaMuertoPorUnErrorDelProveedor(t *testing.T) {
	sender, srv := newTestSender(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		io.WriteString(w, `{"error":{"code":503,"status":"UNAVAILABLE","message":"The service is currently unavailable."}}`)
	})
	defer srv.Close()

	dead, err := sender.Send(context.Background(), []string{"sano-1", "sano-2"}, domain.StreakAtRiskNotification(3))
	if len(dead) != 0 {
		t.Fatalf("un 503 no debe marcar tokens como muertos, marco %v", dead)
	}
	if err == nil {
		t.Fatal("esperaba que el error del proveedor se propagara")
	}
}

// Un dispositivo roto no puede impedir el aviso a los demas.
func TestSendSigueConLosDemasTrasUnFallo(t *testing.T) {
	var recibidos int
	sender, srv := newTestSender(func(w http.ResponseWriter, r *http.Request) {
		recibidos++
		body, _ := io.ReadAll(r.Body)
		if strings.Contains(string(body), `"token":"roto"`) {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, `{"error":{"code":500,"status":"INTERNAL","message":"boom"}}`)
			return
		}
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `{"name":"projects/p/messages/1"}`)
	})
	defer srv.Close()

	_, err := sender.Send(context.Background(), []string{"roto", "sano"}, domain.PlanAssignedNotification("Upper"))
	if err == nil {
		t.Fatal("esperaba que reportara el fallo del primero")
	}
	if recibidos != 2 {
		t.Fatalf("esperaba que intentara con los 2 tokens, intento %d", recibidos)
	}
}

func TestIsDeadTokenError(t *testing.T) {
	var invalidArg fcmErrorResponse
	invalidArg.Error.Details = []struct {
		ErrorCode string `json:"errorCode"`
	}{{ErrorCode: "INVALID_ARGUMENT"}}

	cases := []struct {
		name   string
		status int
		resp   fcmErrorResponse
		want   bool
	}{
		{"404 sin detalles", http.StatusNotFound, fcmErrorResponse{}, true},
		{"argumento invalido", http.StatusBadRequest, invalidArg, true},
		{"proveedor caido", http.StatusServiceUnavailable, fcmErrorResponse{}, false},
		{"cuota agotada", http.StatusTooManyRequests, fcmErrorResponse{}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isDeadTokenError(tc.status, tc.resp); got != tc.want {
				t.Fatalf("isDeadTokenError = %v, esperaba %v", got, tc.want)
			}
		})
	}
}
