import "package:http/http.dart" as http;
import 'package:qashare/Config/api_config.dart';
import 'dart:convert';
import 'package:qashare/Models/auth_model.dart';
import 'package:qashare/Models/group_model.dart';
import 'package:qashare/Models/user_models.dart';

class ApiService {
  // ================= REGISTER =================
  static Future<RegisterResult> registerUser({
    required String username,
    required String name,
    required String email,
    required String password,
  }) async {
    final url = Uri.parse("${ApiConfig.baseUrl}/auth/register");

    try {
      final response = await http.post(
        url,
        headers: {"Content-Type": "application/json"},
        body: jsonEncode({
          "username": username,
          "name": name,
          "email": email,
          "password": password,
        }),
      );

      //Success
      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);

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

      // ❌ USER ALREADY EXISTS
      if (response.statusCode == 409) {
        return RegisterResult.error("User already exists.");
      }

      // ❌ SERVER ERROR
      if (response.statusCode == 500) {
        return RegisterResult.error("Server error. Try again later.");
      }

      return RegisterResult.error("Unexpected error (${response.statusCode})");
    } catch (e) {
      return RegisterResult.error("Cannot connect to server");
    }
  }

  //==================LOGIN=================

  static Future<LoginResult> loginUser({
    required String email,
    required String password,
  }) async {
    final url = Uri.parse("${ApiConfig.baseUrl}/auth/login");

    try {
      final response = await http.post(
        url,
        headers: {"Content-Type": "application/json"},
        body: jsonEncode({"email": email, "password": password}),
      );

      //Success
      if (response.statusCode == 200) {
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

  //======================GET USER PROFILE===============
  static Future<UserResult> getCurrentUser(String token) async {
    final url = Uri.parse("${ApiConfig.baseUrl}/auth/me");

    try {
      final response = await http.get(
        url,
        headers: {
          "Content-Type": "application/json",
          "Authorization": "Bearer $token",
        },
      );
      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);

        return UserResult.success(
          userId: data["user_id"],
          name: data["name"],
          email: data["email"],
          guest: data["guest"],
          createdAt: data["created_at"],
        );
      }
      if (response.statusCode == 401) {
        return UserResult.error("Session expired. Please login again.");
      }

      if (response.statusCode == 500) {
        return UserResult.error("Server error. Try again later.");
      }

      return UserResult.error("Unexpected error (${response.statusCode})");
    } catch (e) {
      return UserResult.error("Unable to connect to server");
    }
  }

  //====================DISPLAY GROUP============
  static Future<GroupListResult> displayGroup(token) async {
    final url = Uri.parse("${ApiConfig.baseUrl}/groups/me");

    try {
      final response = await http.get(
        url,
        headers: {
          "Content-Type": "application/json",
          "Authorization": "Bearer $token",
        },
      );

      if (response.statusCode == 200) {
        final decode = jsonDecode(response.body);

        if (decode == null) {
          return GroupListResult.success([]);
        }

        final List data = decode as List;

        final groups = data.map((item) => Group.fromJson(item)).toList();
        return GroupListResult.success(groups);
      }

      if (response.statusCode == 401) {
        return GroupListResult.error("Session expired. Please login again.");
      }

      if (response.statusCode == 500) {
        return GroupListResult.error("Server error. Try again later.");
      }

      return GroupListResult.error("Unexpected error (${response.statusCode})");
    } catch (e) {
      return GroupListResult.error(e.toString());
    }
  }

  //=============CREATE GROUP===============
  static Future<GroupCreateResult> createGroup({
    required String token,
    required String name,
    required String description,
  }) async {
    final url = Uri.parse("${ApiConfig.baseUrl}/groups/");
    try {
      final response = await http.post(
        url,
        headers: {
          "Content-Type": "application/json",
          "Authorization": "Bearer $token",
        },
        body: jsonEncode({"name": name, "description": description}),
      );
      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        final group = Group.fromJson(data);
        return GroupCreateResult.success(group);
      }
      if (response.statusCode == 400) {
        return GroupCreateResult.error("Invalid request data");
      }
      if (response.statusCode == 401) {
        return GroupCreateResult.error("Session expired. Please login again.");
      }
      if (response.statusCode == 500) {
        return GroupCreateResult.error("Server error. Try again later.");
      }
      return GroupCreateResult.error(
        "Unexpected error (${response.statusCode})",
      );
    } catch (e) {
      return GroupCreateResult.error("Unable to connect to server");
    }
  }
}
