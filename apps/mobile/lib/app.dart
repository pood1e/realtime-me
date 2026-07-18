import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

import 'features/home/home_screen.dart';
import 'features/pairing/credential_gate.dart';
import 'features/pairing/pairing_screen.dart';
import 'features/terminal/terminal_screen.dart';
import 'features/thread/thread_screen.dart';
import 'features/workspace/workspace_screen.dart';
import 'ui/theme.dart';

final _router = GoRouter(
  initialLocation: '/',
  routes: [
    GoRoute(path: '/', builder: (context, state) => const HomeScreen()),
    GoRoute(
      path: '/pairing',
      builder: (context, state) => const PairingScreen(closeOnSuccess: true),
    ),
    GoRoute(
      path: '/workspaces/:uid',
      builder: (context, state) => CredentialGate(
        child: WorkspaceScreen(workspaceUid: state.pathParameters['uid']!),
      ),
    ),
    GoRoute(
      path: '/threads/:uid',
      builder: (context, state) => CredentialGate(
        child: ThreadScreen(threadUid: state.pathParameters['uid']!),
      ),
    ),
    GoRoute(
      path: '/terminals/:uid',
      builder: (context, state) => CredentialGate(
        child: TerminalScreen(terminalUid: state.pathParameters['uid']!),
      ),
    ),
  ],
);

class RealtimeMeApp extends StatelessWidget {
  const RealtimeMeApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp.router(
      title: 'Realtime Me',
      debugShowCheckedModeBanner: false,
      theme: buildTheme(Brightness.light),
      darkTheme: buildTheme(Brightness.dark),
      themeMode: ThemeMode.system,
      routerConfig: _router,
    );
  }
}
