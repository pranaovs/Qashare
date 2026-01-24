import 'package:flutter/material.dart';

class HomePage extends StatefulWidget {
  const HomePage({super.key});

  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.black,
      appBar:AppBar(
        backgroundColor: Colors.black,
        title: const Text("My Groups",style: TextStyle(color: Colors.white),),
        actions: [
          Padding(
            padding: const EdgeInsets.only(right: 15),
            child: GestureDetector(
              onTap: (){
                Navigator.pushNamed(context, "/profile");
              },
              child: const CircleAvatar(
                radius: 18,
                backgroundColor: Colors.black,
                child: Icon(
                  Icons.person,
                  color: Colors.white,
                  size: 30,
                ),
              ),
            ),
          ),
        ],
      ) ,
      body: Center(
        child: const Text(
          "Home Page",
          style: TextStyle(color: Colors.white),
        ),
      ),
    ); 
  }
} 
