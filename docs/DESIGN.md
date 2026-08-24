# Myvibesfit — Diseño

Estética gamificada (referencia Duolingo) con un límite estricto: **el gimnasio solo
personaliza un token de color y su logo.** Todo lo demás — tipografía, espaciado, forma,
motion — es fijo. Así cien gimnasios distintos siguen pareciendo la misma app bien hecha,
no cien apps improvisadas.

## 1. Principio de marca

```
organization.brand_color  →  --brand           (un solo valor, viene de la BD)
organization.logo_url     →  header + splash
```

`--brand` alimenta: botón primario, progreso de racha, anillo de XP, iconos activos.
Nunca se usa para texto sobre fondo claro sin pasar el chequeo de contraste (mínimo 4.5:1).
Si un gimnasio no define color, cae al `--brand` por defecto del sistema.

## 2. Tokens de color

Modo claro y oscuro desde el día 1 (usuario en gimnasio con poca luz vs. oficina).

```
--brand           #C6FF4F   Verde lima — energía, éxito, progreso (default)
--brand-ink       #1B2B0E   Texto sobre --brand

--bg              #0E0F12 (oscuro) / #FAFAF7 (claro)
--surface         #17181C (oscuro) / #FFFFFF (claro)
--surface-raised  #1E2025 (oscuro) / #F2F2ED (claro)
--border          #2A2C32 (oscuro) / #E4E4DD (claro)

--text            #F5F5F0 (oscuro) / #14151A (claro)
--text-muted      #9A9CA5 (oscuro) / #6B6D76 (claro)

--success         #4ADE80    Serie completada, hábito cumplido
--warning         #FBBF24    Deload, sesión pendiente
--danger          #F87171    Fallo, racha rota
--info            #60A5FA    Sugerencia de IA pendiente

--streak-fire     #FF6B35    Único color que NO cambia con --brand: la racha es siempre fuego
```

## 3. Tipografía

**Manrope** (variable, Google Fonts) — geométrica, redondeada, legible en números grandes.
Los pesos exagerados (800) son lo que da la sensación "gamificada" sin ilustración cara.

```
display   32/38  weight 800   "48 kg"  — el número que el usuario mira mientras entrena
h1        24/30  weight 700   Título de pantalla
h2        18/24  weight 700   Título de tarjeta
body      15/22  weight 500
caption   13/18  weight 500   --text-muted
label     12/16  weight 700   uppercase, tracking 0.04em — headers de sección
```

## 4. Espaciado y forma

Escala de 4px: `4·8·12·16·24·32·48·64`. Radios generosos, coherentes con el tono lúdico:

```
radius-sm    8px    chips, badges
radius-md    16px   tarjetas, inputs
radius-lg    24px   modales, hojas inferiores
radius-full          avatares, botón de acción flotante
```

Sombra única y sutil (`0 2px 12px rgb(0 0 0 / 0.24)` en oscuro) — nada de sombras múltiples
tipo Material antiguo.

## 5. Componentes clave

**Anillo de progreso diario** (pantalla de inicio) — como el anillo de actividad de Apple,
pero con `--brand`: un vistazo dice si hoy toca entrenar y si los hábitos están cumplidos.

**Tarjeta de racha** — número grande + icono de fuego (`--streak-fire`, fijo), con estado
"en riesgo" si no se ha entrenado hoy y quedan <4h de día.

**Registro de serie** — la pantalla que más se usa, optimizada a una mano: stepper de peso
y reps con `±`, botón "completar serie" ocupa el 100% del ancho, temporizador de descanso
aparece automáticamente y vibra al terminar.

**Tarjeta de sugerencia IA** — nunca se aplica sola. Muestra el cambio propuesto, el motivo
en una frase, y dos botones iguales en tamaño: Aprobar / Rechazar. Ningún sesgo visual
hacia "aprobar".

**Celebración de logro** — confeti + haptic al desbloquear un achievement o cerrar racha
semanal. Es el único momento con animación grande; todo lo demás es sobrio a propósito.

## 6. Motion

```
micro        120ms  ease-out     tap de botón, check de serie
transition   200ms  ease-in-out  cambio de pantalla, expandir tarjeta
celebration  600ms  spring       confeti de logro, racha
```

Nada de motion decorativo en listas largas (historial, catálogo) — ahí gana la velocidad
de scroll sobre el efecto.

## 7. Flutter — mapeo a ThemeExtension

Los tokens anteriores viven en un `ThemeExtension<AppColors>` propio, no en el `ColorScheme`
de Material por defecto — Material 3 impone demasiados roles que no usamos. Un solo
`AppTheme.from(organization)` construye el tema en runtime a partir de `brand_color`.

```dart
class AppColors extends ThemeExtension<AppColors> {
  final Color brand, brandInk, bg, surface, surfaceRaised, border,
      text, textMuted, success, warning, danger, info, streakFire;
  // ...lerp() y copyWith() estándar de ThemeExtension
}
```

## 8. Accesibilidad

Contraste AA mínimo en todo texto. Todo elemento táctil ≥44×44px. El color nunca es el
único portador de estado (serie completada = check + color, no solo color). Soporte de
`TextScaler` del sistema sin romper layout — nada de alturas fijas en textos largos.
