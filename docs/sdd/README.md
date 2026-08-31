# Catálogo de Plantillas SDD — Myvibesfit

Este directorio contiene las plantillas estandarizadas para aplicar **Spec-Driven Development (SDD)** en cada tipo de cambio dentro del ecosistema de Myvibesfit.

---

## 📁 Índice de Plantillas

| # | Plantilla | Propósito / Caso de Uso | Enlace |
|---|---|---|---|
| **01** | **Feature Integral (Full-Stack)** | Nueva funcionalidad end-to-end (BD + Dominio + API + Web + Flutter) y también bugfixes (Anexo E) — secciones 1-9 obligatorias, Anexos A-E opcionales según lo que la feature necesite (contrato de API detallado, esquema extendido, sync offline, flujo de IA, RCA). | [01_FEATURE_SDD.md](./templates/01_FEATURE_SDD.md) |
| **02** | **ADR / Refactor de Arquitectura** | Decisiones arquitectónicas, rediseño de puertos, separación de UoW o refactors mayores sin una feature nueva de por medio. | [02_ADR_ARCHITECTURE.md](./templates/02_ADR_ARCHITECTURE.md) |

---

## 📋 SDDs Activos (Estado Actual)

| SDD | Título | Estado | Fase actual | Próximo paso |
|-----|--------|--------|-------------|--------------|
| **SDD-001** | Push Notifications (FCM) | `Approved` | Backend implementado, falta Flutter + device real | Implementar adapter Flutter + probar en device |
| **SDD-002** | Progress Photos + ObjectStore | `Draft` | Pendiente revisión arquitectónica | Completar §1-9, validar Definition of Ready |
| **SDD-003** | Worker Operation (cron, streaks) | `Draft` | Pendiente revisión arquitectónica | Completar §1-9, validar Definition of Ready |

> **Nota:** Ver `docs/SDD_METHODOLOGY.md` §2 para Definition of Ready (Fase 1) y Definition of Done (Fase 5) actualizadas.

---

## 🚀 Guía Rápida: ¿Cómo usar una plantilla?

1. **Copiar `01_FEATURE_SDD.md`** (o `02_ADR_ARCHITECTURE.md` si es una decisión de arquitectura sin feature) a `docs/sdd/active/` con el formato `SDD-<NUMERO>-<nombre-corto>.md` (ejemplo: `docs/sdd/active/SDD-001-push-notifications.md`).
2. **Completar las secciones obligatorias** (1-9) y solo los anexos que apliquen, **antes de escribir código de producción**.
3. **Validar Definition of Ready (Fase 1):**
   - [ ] Modelo de dominio + invariantes
   - [ ] Contratos DDL + DTOs
   - [ ] **Impact analysis previo con GitNexus `impact()`**
   - [ ] Anexo(s) identificados
   - [ ] Revisión arquitectónica (checklist en `SDD_METHODOLOGY.md` §3)
4. **Ejecutar el desarrollo** en orden estricto:
   - **Fase 2:** Schema & Ports → `sqlc generate` → `go build` OK
   - **Fase 3:** Domain Tests + Service Tests + Handler Tests + Flutter Tests
   - **Fase 4:** Adapters → Services → Handlers/DTOs → Web → Flutter
   - **Fase 5:** `gitnexus impact()` + `detect_changes()` + CI verde + AuditLog → mover a `completed/`
5. **Mover a `docs/sdd/completed/`** solo tras **Definition of Done completa** (ver `SDD_METHODOLOGY.md` §2 Fase 5).

---

## ⚙️ Gates Automatizados (Obligatorios en CI/Pre-commit)

| Gate | Cuándo | Comando | Qué falla |
|------|--------|---------|-----------|
| **Impact Analysis** | Antes de editar símbolo público | `gitnexus impact({target: "Simbolo", direction: "upstream", repo: "Myvibesfit"})` | Risk HIGH/CRITICAL sin aprobación; UNKNOWN sin confirmar con grep |
| **Change Detection** | Pre-commit / Pre-push | `gitnexus detect_changes({scope: "all", repo: "Myvibesfit"})` | `partial: true` o `truncated: true` sin explicación; `risk_level: unknown` |
| **Pipeline CI** | Push / PR | `go vet/test/build + pnpm lint/build + flutter analyze/test` | Cualquier check en rojo |
| **AuditLog Coverage** | PR review | Revisión manual + grep `AuditLogger.Log` | Mutaciones críticas sin audit log |

---

## 📂 Estructura de Directorios SDD

```
docs/sdd/
├── templates/
│   ├── 01_FEATURE_SDD.md          # Plantilla feature full-stack (9 secciones + 5 anexos)
│   └── 02_ADR_ARCHITECTURE.md     # Plantilla decisión arquitectura / refactor
├── active/
│   ├── SDD-001-push-notifications.md
│   ├── SDD-002-progress-photos.md
│   └── SDD-003-worker-operation.md
└── completed/
    └── README.md                  # Índice histórico (crear al mover primer SDD)
```

---

## 🔗 Referencias Cruzadas

- **Metodología completa:** `docs/SDD_METHODOLOGY.md` (5 fases, reglas de oro, tooling, deuda)
- **Arquitectura:** `docs/ARCHITECTURE.md` (decisiones, endpoints, motor progresión, IA, sync)
- **Diseño:** `docs/DESIGN.md` (tokens, tipografía, componentes, ThemeExtension)
- **Fases de construcción:** `docs/PHASES.md` (0-10, dependencias, qué NO entra en v1)
- **Estado del proyecto:** `AGENTS.md` (qué está verificado, deuda conocida, comandos)
- **Contexto del agente:** `CLAUDE.md` (convenciones workspace, comunicación)

---

*Última actualización: 2026-08-31 — 3 SDDs activos. Los gates se revisaron contra el código:
los que el repo no podía pasar (tests de handler, auditoría de 8 acciones) bajaron a objetivo y
quedaron en §6 con su trigger. `api/internal/docs` verifica en cada `go test` que la metodología
no vuelva a afirmar lo que el código desmiente.*
