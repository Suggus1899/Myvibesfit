import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

import 'app_colors.dart';

class AppTheme {
  AppTheme._();

  static const radiusSm = 8.0;
  static const radiusMd = 16.0;
  static const radiusLg = 24.0;

  static ThemeData from({required Brightness brightness, Color? brandColor}) {
    final colors = brightness == Brightness.dark
        ? AppColors.dark(brand: brandColor)
        : AppColors.light(brand: brandColor);

    final base = GoogleFonts.manropeTextTheme(
      brightness == Brightness.dark ? ThemeData.dark().textTheme : ThemeData.light().textTheme,
    );

    final textTheme = base.copyWith(
      displayMedium: GoogleFonts.manrope(fontSize: 32, height: 38 / 32, fontWeight: FontWeight.w800, color: colors.text),
      headlineMedium: GoogleFonts.manrope(fontSize: 24, height: 30 / 24, fontWeight: FontWeight.w700, color: colors.text),
      titleLarge: GoogleFonts.manrope(fontSize: 18, height: 24 / 18, fontWeight: FontWeight.w700, color: colors.text),
      bodyMedium: GoogleFonts.manrope(fontSize: 15, height: 22 / 15, fontWeight: FontWeight.w500, color: colors.text),
      bodySmall: GoogleFonts.manrope(fontSize: 13, height: 18 / 13, fontWeight: FontWeight.w500, color: colors.textMuted),
      labelLarge: GoogleFonts.manrope(fontSize: 12, height: 16 / 12, fontWeight: FontWeight.w700, letterSpacing: 0.48, color: colors.textMuted),
    );

    final colorScheme = brightness == Brightness.dark
        ? ColorScheme.dark(primary: colors.brand, onPrimary: colors.brandInk, surface: colors.surface, error: colors.danger)
        : ColorScheme.light(primary: colors.brand, onPrimary: colors.brandInk, surface: colors.surface, error: colors.danger);

    return ThemeData(
      brightness: brightness,
      scaffoldBackgroundColor: colors.bg,
      colorScheme: colorScheme,
      textTheme: textTheme,
      extensions: [colors],
      appBarTheme: AppBarTheme(backgroundColor: colors.bg, foregroundColor: colors.text, elevation: 0),
      cardTheme: CardThemeData(
        color: colors.surfaceRaised,
        elevation: 0,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(radiusMd)),
      ),
      elevatedButtonTheme: ElevatedButtonThemeData(
        style: ElevatedButton.styleFrom(
          backgroundColor: colors.brand,
          foregroundColor: colors.brandInk,
          minimumSize: const Size.fromHeight(52),
          textStyle: GoogleFonts.manrope(fontWeight: FontWeight.w700, fontSize: 15),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(radiusMd)),
        ),
      ),
      outlinedButtonTheme: OutlinedButtonThemeData(
        style: OutlinedButton.styleFrom(
          minimumSize: const Size.fromHeight(52),
          side: BorderSide(color: colors.border),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(radiusMd)),
        ),
      ),
      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: colors.surfaceRaised,
        contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
        border: OutlineInputBorder(borderRadius: BorderRadius.circular(radiusMd), borderSide: BorderSide(color: colors.border)),
        enabledBorder: OutlineInputBorder(borderRadius: BorderRadius.circular(radiusMd), borderSide: BorderSide(color: colors.border)),
        focusedBorder: OutlineInputBorder(borderRadius: BorderRadius.circular(radiusMd), borderSide: BorderSide(color: colors.brand, width: 2)),
      ),
      dividerTheme: DividerThemeData(color: colors.border, space: 1),
    );
  }
}
