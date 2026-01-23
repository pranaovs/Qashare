import "package:http/http.dart" as http;
import 'package:flutter/material.dart';
import 'dart:convert';
import 'package:qashare/Models/auth_model.dart';


class ApiService {
  static const String _baseurl = "http://devbox:8080";

  // ================= REGISTER =================
  static Future<RegisterResult> registerUser({
    required String username,
    required String name,
    required String email,
    required String password,}) async{

    final url= Uri.parse("$_baseurl/auth/register");

    try {
      final response = await http.post(url,
        headers: {
        "Content-Type":"application/json",
        },
        body: jsonEncode({
          "username":username,
          "name":name,
          "email":email,
          "password":password
        }),
      );

      //Success
      if (response.statusCode==200){
        final data=jsonDecode(response.body);

        return RegisterResult.success(
          userId: data["user_id"],
          name: data["name"],
          email: data["email"],
          guest: data["guest"],
          createdAt: data["created_at"],
        );
      }

      // ❌ VALIDATION ERROR
      if (response.statusCode == 400) {
        return RegisterResult.error(
          "Invalid input. Please check your details.",
        );
      }

      // ❌ DUPLICATE / DB ERROR
      if (response.statusCode == 500) {
        return RegisterResult.error(
          "User already exists or server error.",
        );
      }

      return RegisterResult.error(
        "Unexpected error (${response.statusCode})",
      );

    }catch(e){
      return RegisterResult.error(
        "Cannot connect to server"
      );
    }
  }

  //==================LOGIN=================

  static Future<LoginResult> loginUser({
    required String email,
    required String password,
  }) async{
    final url= Uri.parse("$_baseurl/auth/login");


    try{
      final response = await http.post(url,
        headers: {
          "Content-Type":"application/json"
        },
        body: jsonEncode({
          "email":email,
          "password":password
        }),
      );

      //Sucess
      if (response.statusCode==200){
        final data = jsonDecode(response.body);
        return LoginResult.success(
          token: data["token"],
          message: data["message"],
        );
      }

      // ❌ BAD REQUEST
      if (response.statusCode == 400) {
        return LoginResult.error("Invalid request data");
      }

      // ❌ WRONG CREDENTIALS
      if (response.statusCode == 401) {
        return LoginResult.error("Invalid email or password");
      }

      // ❌ SERVER ERROR
      if (response.statusCode == 500) {
        return LoginResult.error("Server error. Try again later.");
      }

      return LoginResult.error("Unexpected error (${response.statusCode})");
    } catch (e) {
      return LoginResult.error("Cannot connect to server");
    }

  }
}

