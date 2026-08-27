import 'package:flutter/material.dart';

/// Se muestra mientras AuthController resuelve si hay sesion guardada.
/// Sin esto, la primera pantalla (Home) monta sus providers y dispara
/// llamadas a la API antes de saber si hay token, generando 401 de mas.
class SplashScreen extends StatelessWidget {
  const SplashScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return const Scaffold(body: Center(child: CircularProgressIndicator()));
  }
}
