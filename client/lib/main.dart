import 'package:flutter/material.dart';
import 'package:qashare/Screens/home_page.dart';
import 'package:qashare/Screens/profile_page.dart';

import 'Screens/login_page.dart';
import 'Screens/signup_page.dart';

void main() {
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      debugShowCheckedModeBanner: false,
      title: "Qashare",
      initialRoute: "/login",
      routes: {
        '/login': (context) => const LoginPage(),
        '/signup': (context) => const SignupPage(),
        '/home' : (context) => const HomePage(),
        '/profile' : (context) => const ProfilePage()
      },
    );
  }
}
