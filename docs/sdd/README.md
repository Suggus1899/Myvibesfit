# Catálogo de Plantillas SDD — Myvibesfit

Este directorio contiene las plantillas estandarizadas para aplicar **Spec-Driven Development (SDD)** en cada tipo de cambio dentro del ecosistema de Myvibesfit.

---

## 📁 Índice de Plantillas

| # | Plantilla | Propósito / Caso de Uso | Enlace |
|---|---|---|---|
| **01** | **Feature Integral (Full-Stack)** | Nueva funcionalidad end-to-end (BD + Dominio + API + Web + Flutter) y también bugfixes (Anexo E) — secciones 1-9 obligatorias, Anexos A-E opcionales según lo que la feature necesite (contrato de API detallado, esquema extendido, sync offline, flujo de IA, RCA). | [01_FEATURE_SDD.md](./templates/01_FEATURE_SDD.md) |
| **02** | **ADR / Refactor de Arquitectura** | Decisiones arquitectónicas, rediseño de puertos, separación de UoW o refactors mayores sin una feature nueva de por medio. | [02_ADR_ARCHITECTURE.md](./templates/02_ADR_ARCHITECTURE.md) |

---

## 🚀 Guía Rápida: ¿Cómo usar una plantilla?

1. **Copiar `01_FEATURE_SDD.md`** (o `02_ADR_ARCHITECTURE.md` si es una decisión de arquitectura sin feature) a `docs/sdd/active/` con el formato `SDD-<NUMERO>-<nombre-corto>.md` (ejemplo: `docs/sdd/active/SDD-001-push-notifications.md`).
2. **Completar las secciones obligatorias** (1-9) y solo los anexos que apliquen, antes de escribir código de producción.
3. **Revisar y validar** con el equipo / checklist de arquitectura.
4. **Ejecutar el desarrollo** siguiendo el orden: Schema → Tests de Dominio → Adapters/Services → Handlers/DTOs → UI Web/Móvil.
5. **Mover a `docs/sdd/completed/`** una vez implementado y verificado con los tests en verde.
