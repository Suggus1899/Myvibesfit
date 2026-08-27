import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'core/router/app_router.dart';
import 'core/theme/app_theme.dart';

void main() {
  runApp(const ProviderScope(child: MyvibesfitApp()));
}

class MyvibesfitApp extends ConsumerWidget {
  const MyvibesfitApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final router = ref.watch(routerProvider);
    return MaterialApp.router(
      title: 'Myvibesfit',
      debugShowCheckedModeBanner: false,
      theme: AppTheme.from(brightness: Brightness.light),
      darkTheme: AppTheme.from(brightness: Brightness.dark),
      themeMode: ThemeMode.system,
      routerConfig: router,
    );
  }
}
