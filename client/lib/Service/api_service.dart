import "package:http/http.dart" as http;
import 'dart:convert';

import 'package:qashare/Config/api_config.dart';
import 'package:qashare/Models/add_member_results.dart';
import 'package:qashare/Models/auth_model.dart';
import 'package:qashare/Models/expense_list_model.dart';
import 'package:qashare/Models/expense_model.dart';
import 'package:qashare/Models/group_model.dart';
import 'package:qashare/Models/groupdetail_model.dart';
import 'package:qashare/Models/user_models.dart';
import 'package:qashare/Models/userlookup_model.dart';

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

      if (response.statusCode == 201) {
        final data = jsonDecode(response.body);
        return RegisterResult.success(
          userId: data["user_id"],
          name: data["name"],
          email: data["email"],
          guest: data["guest"],
          createdAt: data["created_at"],
        );
      }

      if (response.statusCode == 400) {
        return RegisterResult.error("Invalid input.");
      }

      if (response.statusCode == 409) {
        return RegisterResult.error("User already exists.");
      }

      if (response.statusCode == 500) {
        return RegisterResult.error("Server error.");
      }

      return RegisterResult.error("Unexpected error (${response.statusCode})");
    } catch (_) {
      return RegisterResult.error("Cannot connect to server");
    }
  }

  // ================= LOGIN =================
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

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        return LoginResult.success(
          token: data["token"],
          message: data["message"],
        );
      }

      if (response.statusCode == 400)
        return LoginResult.error("Invalid request");
      if (response.statusCode == 401)
        return LoginResult.error("Wrong credentials");
      if (response.statusCode == 500) return LoginResult.error("Server error");

      return LoginResult.error("Unexpected error (${response.statusCode})");
    } catch (_) {
      return LoginResult.error("Cannot connect to server");
    }
  }

  // ================= PROFILE =================
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

      if (response.statusCode == 401)
        return UserResult.error("Session expired");
      if (response.statusCode == 500) return UserResult.error("Server error");

      return UserResult.error("Unexpected error (${response.statusCode})");
    } catch (_) {
      return UserResult.error("Unable to connect to server");
    }
  }

  // ================= GROUP LIST =================
  static Future<GroupListResult> displayGroup(String token) async {
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
        final decoded = jsonDecode(response.body);

        // ✅ backend returned empty list
        if (decoded == null) {
          return GroupListResult.success([]);
        }

        // ✅ backend returned []
        if (decoded is List) {
          final groups = decoded
              .map((e) => Group.fromJson(e))
              .toList();
          return GroupListResult.success(groups);
        }

        // ❌ backend returned unexpected structure
        return GroupListResult.error("Invalid response format");
      }

      if (response.statusCode == 401) {
        return GroupListResult.error("Session expired");
      }

      if (response.statusCode == 500) {
        return GroupListResult.error("Server error");
      }

      return GroupListResult.error(
        "Unexpected error (${response.statusCode})",
      );
    } catch (e) {
      return GroupListResult.error("Unable to connect to server");
    }
  }

  // ================= CREATE GROUP =================
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
        return GroupCreateResult.success(Group.fromJson(data));
      }

      if (response.statusCode == 400)
        return GroupCreateResult.error("Invalid data");
      if (response.statusCode == 401)
        return GroupCreateResult.error("Session expired");
      if (response.statusCode == 500)
        return GroupCreateResult.error("Server error");

      return GroupCreateResult.error(
        "Unexpected error (${response.statusCode})",
      );
    } catch (_) {
      return GroupCreateResult.error("Unable to connect to server");
    }
  }

  // ================= GROUP DETAILS =================
  static Future<GroupDetailsResult> getGroupDetails({
    required String token,
    required String groupId,
  }) async {
    final url = Uri.parse("${ApiConfig.baseUrl}/groups/$groupId");

    try {
      final response = await http.get(
        url,
        headers: {
          "Content-Type": "application/json",
          "Authorization": "Bearer $token",
        },
      );

      if (response.statusCode == 200) {
        return GroupDetailsResult.success(
          GroupDetails.fromJson(jsonDecode(response.body)),
        );
      }

      if (response.statusCode == 401)
        return GroupDetailsResult.error("Session expired");
      if (response.statusCode == 403)
        return GroupDetailsResult.error("Not group member");
      if (response.statusCode == 404)
        return GroupDetailsResult.error("Group not found");
      if (response.statusCode == 500)
        return GroupDetailsResult.error("Server error");

      return GroupDetailsResult.error(
        "Unexpected error (${response.statusCode})",
      );
    } catch (e) {
      return GroupDetailsResult.error(e.toString());
    }
  }

  // ================= ADD MEMBER =================
  static Future<AddMemberResult> addMembersToGroup({
    required String token,
    required String groupId,
    required List<String> userIds,
  }) async {
    final url = Uri.parse("${ApiConfig.baseUrl}/groups/$groupId/members");

    try {
      final response = await http.post(
        url,
        headers: {
          "Content-Type": "application/json",
          "Authorization": "Bearer $token",
        },
        body: jsonEncode({"user_ids": userIds}),
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        final List added = data["added_members"] ?? [];
        return AddMemberResult.success(List<String>.from(added));
      }

      if (response.statusCode == 401)
        return AddMemberResult.error("Session expired");
      if (response.statusCode == 403)
        return AddMemberResult.error("Only admin allowed");
      if (response.statusCode == 404)
        return AddMemberResult.error("Group not found");
      if (response.statusCode == 400)
        return AddMemberResult.error("Invalid user IDs");
      if (response.statusCode == 500)
        return AddMemberResult.error("Server error");

      return AddMemberResult.error("Unexpected error (${response.statusCode})");
    } catch (e) {
      return AddMemberResult.error(e.toString());
    }
  }

  // ================= REMOVE MEMBER =================
  static Future<AddMemberResult> removeMembersFromGroup({
    required String token,
    required String groupId,
    required List<String> userIds,
  }) async {
    final url = Uri.parse("${ApiConfig.baseUrl}/groups/$groupId/members");

    try {
      final response = await http.delete(
        url,
        headers: {
          "Content-Type": "application/json",
          "Authorization": "Bearer $token",
        },
        body: jsonEncode({"user_ids": userIds}),
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        final List removed = data["removed_members"] ?? [];
        return AddMemberResult.success(List<String>.from(removed));
      }

      if (response.statusCode == 401)
        return AddMemberResult.error("Session expired");
      if (response.statusCode == 403)
        return AddMemberResult.error("Only admin allowed");
      if (response.statusCode == 404)
        return AddMemberResult.error("Group not found");
      if (response.statusCode == 400)
        return AddMemberResult.error("Cannot remove admin");
      if (response.statusCode == 500)
        return AddMemberResult.error("Server error");

      return AddMemberResult.error("Unexpected error");
    } catch (e) {
      return AddMemberResult.error(e.toString());
    }
  }

  // ================= SEARCH USER =================
  static Future<UserLookupResult> searchUserByEmail({
    required String token,
    required String email,
  }) async {
    final url = Uri.parse("${ApiConfig.baseUrl}/users/search/email/$email");

    try {
      final response = await http.get(
        url,
        headers: {
          "Content-Type": "application/json",
          "Authorization": "Bearer $token",
        },
      );

      if (response.statusCode == 200) {
        return UserLookupResult.success(
          UserLookup.fromJson(jsonDecode(response.body)),
        );
      }

      if (response.statusCode == 400)
        return UserLookupResult.error("Invalid email");
      if (response.statusCode == 401)
        return UserLookupResult.error("Session expired");
      if (response.statusCode == 500)
        return UserLookupResult.error("User not found");

      return UserLookupResult.error("Unexpected error");
    } catch (e) {
      return UserLookupResult.error(e.toString());
    }
  }

  // ================= GROUP EXPENSES =================
  static Future<ExpenseListResult> getGroupExpenses({
    required String token,
    required String groupId,
  }) async {
    final url = Uri.parse("${ApiConfig.baseUrl}/groups/$groupId/expenses");

    try {
      final response = await http.get(
        url,
        headers: {
          "Content-Type": "application/json",
          "Authorization": "Bearer $token",
        },
      );

      if (response.statusCode == 200) {
        final List data = jsonDecode(response.body);
        final expenses = data.map((e) => GroupExpense.fromJson(e)).toList();
        return ExpenseListResult.success(expenses);
      }

      if (response.statusCode == 401)
        return ExpenseListResult.error("Session expired");
      if (response.statusCode == 403)
        return ExpenseListResult.error("Not group member");
      if (response.statusCode == 404)
        return ExpenseListResult.error("Group not found");
      if (response.statusCode == 500)
        return ExpenseListResult.error("Server error");

      return ExpenseListResult.error("Unexpected error");
    } catch (e) {
      return ExpenseListResult.error(e.toString());
    }
  }

  // ================= CREATE EXPENSE =================
  static Future<BasicResult> createExpenseAdvanced({
    required String token,
    required ExpenseRequest request,
  }) async {
    final url = Uri.parse("${ApiConfig.baseUrl}/expenses/");

    try {
      final response = await http.post(
        url,
        headers: {
          "Content-Type": "application/json",
          "Authorization": "Bearer $token",
        },
        body: jsonEncode(request.toJson()),
      );

      if (response.statusCode == 200) return BasicResult.success();
      if (response.statusCode == 400)
        return BasicResult.error("Split mismatch");
      if (response.statusCode == 401)
        return BasicResult.error("Session expired");
      if (response.statusCode == 403)
        return BasicResult.error("Not group member");
      if (response.statusCode == 500) return BasicResult.error("Server error");

      return BasicResult.error("Unexpected error");
    } catch (e) {
      return BasicResult.error(e.toString());
    }
  }

  static Future<UserLookupResult> createGuestUser({required String token, required String email}) async {
    final url = Uri.parse("${ApiConfig.baseUrl}/auth/guest");
    try {
      final response = await http.post(url,
          headers: {
            "Content-Type": "application/json",
            "Authorization": "Bearer $token",
          },
          body: jsonEncode({"email":email}),
      );

      if (response.statusCode == 201) {
        return UserLookupResult.success(
          UserLookup.fromJson(jsonDecode(response.body)),
        );
      }

      if (response.statusCode == 400) {
        return UserLookupResult.error("Invalid email");
      }

      if (response.statusCode == 401) {
        return UserLookupResult.error("SESSION_EXPIRED");
      }

      if (response.statusCode == 409) {
        return UserLookupResult.error("EMAIL_EXISTS");
      }

      if (response.statusCode == 500) {
        return UserLookupResult.error("SERVER_ERROR");
      }

      return UserLookupResult.error("UNEXPECTED_ERROR");
    } catch (_) {
      return UserLookupResult.error("NETWORK_ERROR");
    }
  }
  
}
