# ADR-XXX: [Título de la Decisión de Arquitectura / Refactor]
> **Tipo:** Architecture Decision Record & Refactoring SDD  
> **Fecha:** AAAA-MM-DD  
> **Estado:** `Proposed` | `Accepted` | `Rejected` | `Superseded`  
> **Decisores:** [Nombres / Equipo]

---

## 1. Contexto y Problema

- **Situación actual:** [Describir el diseño actual y sus limitaciones / deuda técnica].
- **Fricciones detectadas:** [Ejemplo: Alto acoplamiento, cuellos de botella de rendimiento, límites de extensibilidad].
- **Métricas / Señales de cambio:** [Ejemplo: `TxRepos` superó los 6 agregados, latencias altas en sync, etc.].

---

## 2. Decisión Tomada

- **Arquitectura propuesta:** [Descripción concisa de la solución arquitectónica].
- **Patrones aplicados:** [Ejemplo: Ports & Adapters, CQRS segregado, UnitOfWork específico por caso de uso, Outbox Pattern].

---

## 3. Opciones Consideradas y Análisis Comparativo

| Criterio | Opción A: [Nombre] | Opción B: [Nombre] (Elegida) | Opción C: [Nombre] |
|---|---|---|---|
| **Complejidad** | Baja | Media | Alta |
| **Aislamiento** | Pobre | Excelente | Excelente |
| **Rendimiento** | Regular | Óptimo | Óptimo |
| **Esfuerzo de Migración** | Bajo | Moderado | Muy Alto |

### Justificación de la Elección
- [Detallar por qué la opción elegida es la mejor para Myvibesfit considerando monorepo, VPS y mantenimiento].

---

## 4. Diseño Técnico Detallado

### 4.1 Cambios en Interfaces / Puertos de Dominio (`internal/domain/`)
```go
// Antes:
// ...

// Después:
// ...
```

### 4.2 Cambios en Servicios y Transacciones (`internal/service/`)
- [Detalle de orquestación y nuevos flujos].

### 4.3 Adaptadores y Persistencia (`internal/adapter/`)
- [Cambios en adaptadores Postgres, SQLC o integraciones externas].

---

## 5. Plan de Migración y Despliegue (Sin Downtime)

1. **Paso 1:** [Crear nuevas interfaces y adaptadores en paralelo].
2. **Paso 2:** [Migrar servicios de forma gradual].
3. **Paso 3:** [Deprecar y eliminar puertos/estructuras obsoletas].

---

## 6. Consecuencias

### Positivas (+)
- [Beneficio 1: ej. Pruebas unitarias 100% aisladas sin DB].
- [Beneficio 2: ej. Menor contención de bloqueos transaccionales].

### Negativas / Trade-offs (-)
- [Trade-off 1: ej. Mayor cantidad de ficheros / boilerplate de mapeo].

---

## 7. Plan de Verificación

- [ ] Análisis de impacto con GitNexus (`impact` sin callers residuales rotos).
- [ ] Suite de pruebas de regresión en verde (`go test ./...`).
- [ ] Validación de compilación en todas las capas del monorepo.
