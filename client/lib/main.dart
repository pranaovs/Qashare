import 'package:dynamic_color/dynamic_color.dart';
import 'package:flutter/material.dart';
import 'package:qashare/Screens/creategroup_page.dart';
import 'package:qashare/Screens/home_page.dart';
import 'package:qashare/Screens/profile_page.dart';

import 'Screens/createexpense_page.dart';
import 'Screens/login_page.dart';
import 'Screens/signup_page.dart';

void main() {
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return DynamicColorBuilder(
      builder: (ColorScheme? lightDynamic, ColorScheme? darkDynamic) {
        return MaterialApp(
          debugShowCheckedModeBanner: false,
          title: "Qashare",
          themeMode: ThemeMode.system,
          theme: ThemeData(
            useMaterial3: true,
            colorScheme:
                lightDynamic ?? ColorScheme.fromSeed(seedColor: Colors.blue),
          ),
          darkTheme: ThemeData(
            useMaterial3: true,
            colorScheme:
                darkDynamic ??
                ColorScheme.fromSeed(
                  seedColor: Colors.blue,
                  brightness: Brightness.dark,
                ),
          ),
          initialRoute: "/login",
          routes: {
            '/login': (context) => const LoginPage(),
            '/signup': (context) => const SignupPage(),
            '/home': (context) => const HomePage(),
            '/profile': (context) => const ProfilePage(),
            '/creategroup': (context) => const CreategroupPage(),
            "/create-expense": (context) {
              final args =
                  ModalRoute.of(context)!.settings.arguments
                      as Map<String, dynamic>;

              return CreateExpensePage(
                groupId: args["groupId"],
                members: args["members"],
              );
            }, // Add more routes as needed
          },
        );
      },
    );
  }
}
