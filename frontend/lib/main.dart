import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:go_router/go_router.dart';
import 'providers/auth_provider.dart';
import 'providers/groups_provider.dart';
import 'providers/expenses_provider.dart';
import 'screens/splash_screen.dart';
import 'screens/login_screen.dart';
import 'screens/register_screen.dart';
import 'screens/groups_list_screen.dart';
import 'screens/create_group_screen.dart';
import 'screens/group_details_screen.dart';
import 'screens/profile_screen.dart';
import 'screens/create_expense_screen.dart';
import 'screens/expense_details_screen.dart';
import 'screens/settlements_screen.dart';

void main() {
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MultiProvider(
      providers: [
        ChangeNotifierProvider(create: (_) => AuthProvider()),
        ChangeNotifierProvider(create: (_) => GroupsProvider()),
        ChangeNotifierProvider(create: (_) => ExpensesProvider()),
      ],
      child: Consumer<AuthProvider>(
        builder: (context, authProvider, _) {
          return MaterialApp.router(
            title: 'Shared Expenses',
            debugShowCheckedModeBanner: false,
            theme: ThemeData(
              colorScheme: ColorScheme.fromSeed(
                seedColor: Colors.teal,
                brightness: Brightness.light,
              ),
              useMaterial3: true,
              cardTheme: CardThemeData(
                elevation: 2,
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12),
                ),
              ),
              inputDecorationTheme: InputDecorationTheme(
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(8),
                ),
              ),
            ),
            darkTheme: ThemeData(
              colorScheme: ColorScheme.fromSeed(
                seedColor: Colors.teal,
                brightness: Brightness.dark,
              ),
              useMaterial3: true,
              cardTheme: CardThemeData(
                elevation: 2,
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12),
                ),
              ),
              inputDecorationTheme: InputDecorationTheme(
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(8),
                ),
              ),
            ),
            themeMode: ThemeMode.system,
            routerConfig: _createRouter(authProvider),
          );
        },
      ),
    );
  }

  GoRouter _createRouter(AuthProvider authProvider) {
    return GoRouter(
      initialLocation: '/',
      refreshListenable: authProvider,
      redirect: (context, state) {
        final isAuthenticated = authProvider.isAuthenticated;
        final isLoading = authProvider.isLoading;

        if (isLoading) {
          return null;
        }

        final isOnAuthScreen = state.matchedLocation == '/login' ||
            state.matchedLocation == '/register';

        if (!isAuthenticated && !isOnAuthScreen) {
          return '/login';
        }

        if (isAuthenticated && isOnAuthScreen) {
          return '/groups';
        }

        if (isAuthenticated && state.matchedLocation == '/') {
          return '/groups';
        }

        return null;
      },
      routes: [
        GoRoute(
          path: '/',
          builder: (context, state) => const SplashScreen(),
        ),
        GoRoute(
          path: '/login',
          builder: (context, state) => const LoginScreen(),
        ),
        GoRoute(
          path: '/register',
          builder: (context, state) => const RegisterScreen(),
        ),
        GoRoute(
          path: '/groups',
          builder: (context, state) => const GroupsListScreen(),
        ),
        GoRoute(
          path: '/groups/create',
          builder: (context, state) => const CreateGroupScreen(),
        ),
        GoRoute(
          path: '/groups/:id',
          builder: (context, state) {
            final id = state.pathParameters['id']!;
            return GroupDetailsScreen(groupId: id);
          },
        ),
        GoRoute(
          path: '/profile',
          builder: (context, state) => const ProfileScreen(),
        ),
        GoRoute(
          path: '/expenses/create',
          builder: (context, state) {
            final groupId = state.extra as String;
            return CreateExpenseScreen(groupId: groupId);
          },
        ),
        GoRoute(
          path: '/expenses/:id',
          builder: (context, state) {
            final id = state.pathParameters['id']!;
            return ExpenseDetailsScreen(expenseId: id);
          },
        ),
        GoRoute(
          path: '/groups/:id/settlements',
          builder: (context, state) {
            final id = state.pathParameters['id']!;
            return SettlementsScreen(groupId: id);
          },
        ),
      ],
    );
  }
}
