import 'package:flutter/material.dart';
import 'package:qashare/Config/token_storage.dart';

class AuthcheckPage extends StatefulWidget {
  const AuthcheckPage({super.key});

  @override
  State<AuthcheckPage> createState() => _AuthcheckPageState();
}

class _AuthcheckPageState extends State<AuthcheckPage> {
  @override

  void initState(){
    super.initState();
    _check();
  }

  Future<void> _check() async{
    final token =  await TokenStorage.getToken();

    if (!mounted) return;

    if (token != null){
      Navigator.pushReplacementNamed(context, '/home');
    }
    else{
      Navigator.pushReplacementNamed(context, '/login');
    }
  }

  Widget build(BuildContext context) {
    return Scaffold(
      body: Center(child: CircularProgressIndicator(
          color: Theme.of(context).colorScheme.primary
        ),
      ),
    );
  }
}
