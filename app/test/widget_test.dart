// Smoke test del sistema de colores (docs/DESIGN.md #2). Prueba AppColors
// directo, no AppTheme.from(): este ultimo tira de google_fonts, que
// descarga Manrope por red y no tiene nada que ver con la logica de color
// que este test quiere proteger.
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:myvibesfit_app/core/theme/app_colors.dart';

void main() {
  test('AppColors.dark() usa el brand por defecto y fondo oscuro', () {
    final colors = AppColors.dark();
    expect(colors.brand, const Color(0xFFC6FF4F));
    expect(colors.bg, const Color(0xFF0E0F12));
  });

  test('AppColors.light() usa fondo claro', () {
    final colors = AppColors.light();
    expect(colors.bg, const Color(0xFFFAFAF7));
  });

  test('brandColor personalizado del gimnasio se respeta, el resto no cambia', () {
    const customBrand = Color(0xFFFF0000);
    final colors = AppColors.dark(brand: customBrand);

    expect(colors.brand, customBrand);
    expect(colors.streakFire, const Color(0xFFFF6B35)); // nunca cambia con brand
  });

  test('lerp interpola entre dos paletas sin perder streakFire', () {
    final a = AppColors.dark();
    final b = AppColors.light();
    final mid = a.lerp(b, 0.5);

    expect(mid.streakFire, const Color(0xFFFF6B35));
  });
}
