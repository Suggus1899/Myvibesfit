// Package docs_test no contiene codigo de produccion: contrasta las
// afirmaciones verificables de docs/SDD_METHODOLOGY.md contra el codigo.
//
// Existe porque ese documento ya afirmo tres veces cosas que el codigo
// desmentia (el motor de progresion "sin llamadores", device_token "sin uso",
// el conteo de flujos de TxRepos) y nadie lo noto hasta releerlo a mano. Un
// documento que se desactualiza en silencio es peor que no tenerlo.
package docs_test

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

const methodologyPath = "../../../docs/SDD_METHODOLOGY.md"

// marcador lee un comentario HTML con forma <!-- clave: valor -->.
func marcador(t *testing.T, doc, clave string) string {
	t.Helper()
	re := regexp.MustCompile(`<!--\s*` + regexp.QuoteMeta(clave) + `:\s*(.*?)\s*-->`)
	m := re.FindStringSubmatch(doc)
	if m == nil {
		t.Fatalf("falta el marcador <!-- %s: ... --> en %s. Si lo quitaste a proposito, "+
			"borra tambien el test que lo verifica.", clave, methodologyPath)
	}
	return m[1]
}

func lista(v string) []string {
	partes := strings.Split(v, ",")
	out := make([]string, 0, len(partes))
	for _, p := range partes {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}

func leerDoc(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(methodologyPath)
	if err != nil {
		t.Fatalf("no se pudo leer la metodologia: %v", err)
	}
	return string(b)
}

// TestAccionesAuditadas compara la lista declarada en la metodologia con las
// que realmente escriben en audit_log. Falla en las dos direcciones: si el
// documento promete una auditoria que no existe, y si alguien agrega una
// mutacion auditada sin anotarla.
func TestAccionesAuditadas(t *testing.T) {
	declaradas := lista(marcador(t, leerDoc(t), "audited-actions"))

	archivos, err := filepath.Glob("../service/*.go")
	if err != nil || len(archivos) == 0 {
		t.Fatalf("no se encontraron los services: %v", err)
	}

	// audit.Log(ctx, orgID, actorID, "accion", "entidad", ...)
	re := regexp.MustCompile(`\.Log\(\s*ctx\s*,[^,]+,[^,]+,\s*"([a-z_]+\.[a-z_]+)"`)
	vistas := map[string]bool{}
	for _, f := range archivos {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("no se pudo leer %s: %v", f, err)
		}
		for _, m := range re.FindAllStringSubmatch(string(b), -1) {
			vistas[m[1]] = true
		}
	}

	reales := make([]string, 0, len(vistas))
	for a := range vistas {
		reales = append(reales, a)
	}
	sort.Strings(reales)

	if strings.Join(declaradas, ",") != strings.Join(reales, ",") {
		t.Fatalf("la metodologia y el codigo no coinciden en que se audita.\n"+
			"  declarado en el doc: %v\n"+
			"  encontrado en codigo: %v\n"+
			"Actualiza el marcador <!-- audited-actions: ... --> o agrega el audit.Log que falta.",
			declaradas, reales)
	}
}

// TestConteoDeTxRepos evita que la metodologia razone sobre un TxRepos que ya
// no existe: la regla de oro 7 decide cuando partirlo, y esa decision no puede
// apoyarse en un numero viejo.
func TestConteoDeTxRepos(t *testing.T) {
	declarado, err := strconv.Atoi(marcador(t, leerDoc(t), "txrepos-fields"))
	if err != nil {
		t.Fatalf("el marcador txrepos-fields no es un numero: %v", err)
	}

	b, err := os.ReadFile("../domain/unit_of_work.go")
	if err != nil {
		t.Fatalf("no se pudo leer unit_of_work.go: %v", err)
	}
	src := string(b)

	ini := strings.Index(src, "type TxRepos struct {")
	if ini < 0 {
		t.Fatal("no se encontro la definicion de TxRepos")
	}
	fin := strings.Index(src[ini:], "}")
	if fin < 0 {
		t.Fatal("definicion de TxRepos sin cerrar")
	}

	campos := 0
	for _, linea := range strings.Split(src[ini:ini+fin], "\n")[1:] {
		l := strings.TrimSpace(linea)
		if l == "" || strings.HasPrefix(l, "//") {
			continue
		}
		campos++
	}

	if campos != declarado {
		t.Fatalf("TxRepos tiene %d campos y la metodologia declara %d. "+
			"Actualiza <!-- txrepos-fields: N --> y la fila de §6.", campos, declarado)
	}
}

// TestTablasSinUso comprueba que lo que la metodologia llama "modelado sin uso"
// siga sin queries. Es la afirmacion que fallo con device_token: se implemento
// y el documento siguio diciendo que no se usaba.
func TestTablasSinUso(t *testing.T) {
	declaradas := lista(marcador(t, leerDoc(t), "tables-unused"))

	archivos, err := filepath.Glob("../../db/queries/*.sql")
	if err != nil || len(archivos) == 0 {
		t.Fatalf("no se encontraron las queries: %v", err)
	}

	var todas strings.Builder
	for _, f := range archivos {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("no se pudo leer %s: %v", f, err)
		}
		todas.Write(b)
	}
	sql := todas.String()

	for _, tabla := range declaradas {
		// Se busca la tabla como palabra: "habit" no debe casar con "client_habit".
		re := regexp.MustCompile(`\b` + regexp.QuoteMeta(tabla) + `\b`)
		if re.MatchString(sql) {
			t.Errorf("la metodologia declara %q como tabla sin uso, pero aparece en db/queries/. "+
				"Sacala del marcador <!-- tables-unused: ... --> y de la fila de §6.", tabla)
		}
	}
}
