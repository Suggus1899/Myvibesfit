import 'package:flutter/material.dart';

/// Tokens de color del sistema de diseño (docs/DESIGN.md). El gimnasio solo
/// personaliza [brand]; todo lo demás es fijo para que la app se sienta
/// coherente sin importar quién la use.
@immutable
class AppColors extends ThemeExtension<AppColors> {
  final Color brand;
  final Color brandInk;
  final Color bg;
  final Color surface;
  final Color surfaceRaised;
  final Color border;
  final Color text;
  final Color textMuted;
  final Color success;
  final Color warning;
  final Color danger;
  final Color info;
  final Color streakFire;

  const AppColors({
    required this.brand,
    required this.brandInk,
    required this.bg,
    required this.surface,
    required this.surfaceRaised,
    required this.border,
    required this.text,
    required this.textMuted,
    required this.success,
    required this.warning,
    required this.danger,
    required this.info,
    required this.streakFire,
  });

  static const _defaultBrand = Color(0xFFC6FF4F);
  static const _streakFire = Color(0xFFFF6B35); // nunca cambia con brand

  factory AppColors.dark({Color? brand}) => AppColors(
        brand: brand ?? _defaultBrand,
        brandInk: const Color(0xFF1B2B0E),
        bg: const Color(0xFF0E0F12),
        surface: const Color(0xFF17181C),
        surfaceRaised: const Color(0xFF1E2025),
        border: const Color(0xFF2A2C32),
        text: const Color(0xFFF5F5F0),
        textMuted: const Color(0xFF9A9CA5),
        success: const Color(0xFF4ADE80),
        warning: const Color(0xFFFBBF24),
        danger: const Color(0xFFF87171),
        info: const Color(0xFF60A5FA),
        streakFire: _streakFire,
      );

  factory AppColors.light({Color? brand}) => AppColors(
        brand: brand ?? _defaultBrand,
        brandInk: const Color(0xFF1B2B0E),
        bg: const Color(0xFFFAFAF7),
        surface: const Color(0xFFFFFFFF),
        surfaceRaised: const Color(0xFFF2F2ED),
        border: const Color(0xFFE4E4DD),
        text: const Color(0xFF14151A),
        textMuted: const Color(0xFF6B6D76),
        success: const Color(0xFF4ADE80),
        warning: const Color(0xFFFBBF24),
        danger: const Color(0xFFF87171),
        info: const Color(0xFF60A5FA),
        streakFire: _streakFire,
      );

  @override
  AppColors copyWith({
    Color? brand,
    Color? brandInk,
    Color? bg,
    Color? surface,
    Color? surfaceRaised,
    Color? border,
    Color? text,
    Color? textMuted,
    Color? success,
    Color? warning,
    Color? danger,
    Color? info,
    Color? streakFire,
  }) {
    return AppColors(
      brand: brand ?? this.brand,
      brandInk: brandInk ?? this.brandInk,
      bg: bg ?? this.bg,
      surface: surface ?? this.surface,
      surfaceRaised: surfaceRaised ?? this.surfaceRaised,
      border: border ?? this.border,
      text: text ?? this.text,
      textMuted: textMuted ?? this.textMuted,
      success: success ?? this.success,
      warning: warning ?? this.warning,
      danger: danger ?? this.danger,
      info: info ?? this.info,
      streakFire: streakFire ?? this.streakFire,
    );
  }

  @override
  AppColors lerp(ThemeExtension<AppColors>? other, double t) {
    if (other is! AppColors) return this;
    return AppColors(
      brand: Color.lerp(brand, other.brand, t)!,
      brandInk: Color.lerp(brandInk, other.brandInk, t)!,
      bg: Color.lerp(bg, other.bg, t)!,
      surface: Color.lerp(surface, other.surface, t)!,
      surfaceRaised: Color.lerp(surfaceRaised, other.surfaceRaised, t)!,
      border: Color.lerp(border, other.border, t)!,
      text: Color.lerp(text, other.text, t)!,
      textMuted: Color.lerp(textMuted, other.textMuted, t)!,
      success: Color.lerp(success, other.success, t)!,
      warning: Color.lerp(warning, other.warning, t)!,
      danger: Color.lerp(danger, other.danger, t)!,
      info: Color.lerp(info, other.info, t)!,
      streakFire: Color.lerp(streakFire, other.streakFire, t)!,
    );
  }
}

extension AppColorsX on BuildContext {
  AppColors get colors => Theme.of(this).extension<AppColors>()!;
}
