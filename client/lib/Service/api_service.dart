import "package:http/http.dart" as http;
import 'dart:convert';

import 'package:qashare/Config/api_config.dart';
import 'package:qashare/Config/token_storage.dart';
import 'package:qashare/Models/add_member_results.dart';
import 'package:qashare/Models/auth_model.dart';
import 'package:qashare/Models/expense_list_model.dart';
import 'package:qashare/Models/expense_model.dart';
import 'package:qashare/Models/expensedetail_model.dart';
import 'package:qashare/Models/group_model.dart';
import 'package:qashare/Models/groupdetail_model.dart';
import 'package:qashare/Models/user_models.dart';
import 'package:qashare/Models/settle_model.dart';
import 'package:qashare/Models/spending_model.dart';
import 'package:qashare/Models/userlookup_model.dart';

class ApiService {
  // ============================================================
  //  INTERNAL: Authenticated HTTP helpers with auto-refresh
  // ============================================================

  /// Performs an authenticated HTTP request.
  /// If the server returns 401/403 (expired token), it will attempt to
  /// refresh the tokens once and retry the request.
  ///
  /// [method]  – "GET", "POST", "DELETE", etc.
  /// [url]     – full URL to call
  /// [body]    – optional JSON-encodable map
  ///
  /// Returns the [http.Response] or throws on network error.
  static Future<http.Response> _authenticatedRequest({
    required String method,
    required Uri url,
    Map<String, dynamic>? body,
  }) async {
    String? accessToken = await TokenStorage.getAccessToken();
    if (accessToken == null) {
      // No token at all – return a fake 401
      return http.Response('{"code":"NO_TOKEN","message":"Not logged in"}', 401);
    }

    // First attempt
    var response = await _rawRequest(
      method: method,
      url: url,
      accessToken: accessToken,
      body: body,
    );

    // If 401 or 403 → try refresh
    if (response.statusCode == 401 || response.statusCode == 403) {
      final refreshed = await _tryRefreshTokens();
      if (refreshed) {
        // Retry with new access token
        accessToken = await TokenStorage.getAccessToken();
        response = await _rawRequest(
          method: method,
          url: url,
          accessToken: accessToken!,
          body: body,
        );
      }
    }

    return response;
  }

  /// Raw HTTP call with Bearer token.
  static Future<http.Response> _rawRequest({
    required String method,
    required Uri url,
    required String accessToken,
    Map<String, dynamic>? body,
  }) async {
    final headers = {
      "Content-Type": "application/json",
      "Authorization": "Bearer $accessToken",
    };

    switch (method.toUpperCase()) {
      case "GET":
        return await http.get(url, headers: headers);
      case "POST":
        return await http.post(
          url,
          headers: headers,
          body: body != null ? jsonEncode(body) : null,
        );
      case "DELETE":
        return await http.delete(
          url,
          headers: headers,
          body: body != null ? jsonEncode(body) : null,
        );
      case "PUT":
        return await http.put(
          url,
          headers: headers,
          body: body != null ? jsonEncode(body) : null,
        );
      default:
        return await http.get(url, headers: headers);
    }
  }

  /// Attempts to refresh the access token using the stored refresh token.
  /// Returns `true` if successful (tokens are updated in storage).
  static Future<bool> _tryRefreshTokens() async {
    final refreshToken = await TokenStorage.getRefreshToken();
    if (refreshToken == null) return false;

    try {
      final url = Uri.parse("${ApiConfig.baseUrl}/auth/refresh");
      final response = await http.post(
        url,
        headers: {"Content-Type": "application/json"},
        body: jsonEncode({"refresh_token": refreshToken}),
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        await TokenStorage.saveTokens(
          accessToken: data["access_token"],
          refreshToken: data["refresh_token"],
        );
        return true;
      }

      // Refresh token is invalid/expired → clear everything
      await TokenStorage.clear();
      return false;
    } catch (_) {
      return false;
    }
  }

  // ============================================================
  //  AUTH ENDPOINTS
  // ============================================================

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
          "email": email,
          "name": name,
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
          accessToken: data["access_token"],
          refreshToken: data["refresh_token"],
          tokenType: data["token_type"] ?? "Bearer",
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

  // ================= REFRESH =================
  static Future<RefreshResult> refreshTokens() async {
    final refreshToken = await TokenStorage.getRefreshToken();
    if (refreshToken == null) {
      return RefreshResult.error("No refresh token");
    }

    final url = Uri.parse("${ApiConfig.baseUrl}/auth/refresh");

    try {
      final response = await http.post(
        url,
        headers: {"Content-Type": "application/json"},
        body: jsonEncode({"refresh_token": refreshToken}),
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        final result = RefreshResult.success(
          accessToken: data["access_token"],
          refreshToken: data["refresh_token"],
          tokenType: data["token_type"] ?? "Bearer",
        );

        // Persist the new tokens
        await TokenStorage.saveTokens(
          accessToken: result.accessToken!,
          refreshToken: result.refreshToken!,
        );

        return result;
      }

      if (response.statusCode == 400)
        return RefreshResult.error("Missing refresh token");
      if (response.statusCode == 401)
        return RefreshResult.error("Invalid refresh token");
      if (response.statusCode == 403)
        return RefreshResult.error("Refresh token expired");
      if (response.statusCode == 500)
        return RefreshResult.error("Server error");

      return RefreshResult.error("Unexpected error (${response.statusCode})");
    } catch (_) {
      return RefreshResult.error("Cannot connect to server");
    }
  }

  // ================= LOGOUT =================
  static Future<BasicResult> logout() async {
    try {
      final url = Uri.parse("${ApiConfig.baseUrl}/auth/logout");
      final response = await _authenticatedRequest(method: "POST", url: url);

      // Clear tokens locally regardless of server response
      await TokenStorage.clear();

      if (response.statusCode == 200) {
        return BasicResult.success();
      }
      return BasicResult.success(); // Still clear locally on error
    } catch (_) {
      await TokenStorage.clear();
      return BasicResult.success();
    }
  }

  // ================= LOGOUT ALL =================
  static Future<BasicResult> logoutAll() async {
    try {
      final url = Uri.parse("${ApiConfig.baseUrl}/auth/logout-all");
      final response = await _authenticatedRequest(method: "POST", url: url);

      await TokenStorage.clear();

      if (response.statusCode == 200) {
        return BasicResult.success();
      }
      return BasicResult.success();
    } catch (_) {
      await TokenStorage.clear();
      return BasicResult.success();
    }
  }

  // ============================================================
  //  PROTECTED ENDPOINTS (all use _authenticatedRequest)
  // ============================================================

  // ================= PROFILE =================
  static Future<UserResult> getCurrentUser() async {
    final url = Uri.parse("${ApiConfig.baseUrl}/auth/me");

    try {
      final response = await _authenticatedRequest(method: "GET", url: url);

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
  static Future<GroupListResult> displayGroup() async {
    final url = Uri.parse("${ApiConfig.baseUrl}/groups/me");

    try {
      final response = await _authenticatedRequest(method: "GET", url: url);

      if (response.statusCode == 200) {
        final decoded = jsonDecode(response.body);

        if (decoded == null) {
          return GroupListResult.success([]);
        }

        if (decoded is List) {
          final groups = decoded.map((e) => Group.fromJson(e)).toList();
          return GroupListResult.success(groups);
        }

        return GroupListResult.error("Invalid response format");
      }

      if (response.statusCode == 401) {
        return GroupListResult.error("Session expired");
      }

      if (response.statusCode == 500) {
        return GroupListResult.error("Server error");
      }

      return GroupListResult.error("Unexpected error (${response.statusCode})");
    } catch (e) {
      return GroupListResult.error("Unable to connect to server");
    }
  }

  // ================= CREATE GROUP =================
  static Future<GroupCreateResult> createGroup({
    required String name,
    required String description,
  }) async {
    final url = Uri.parse("${ApiConfig.baseUrl}/groups/");

    try {
      final response = await _authenticatedRequest(
        method: "POST",
        url: url,
        body: {"name": name, "description": description},
      );

      if (response.statusCode == 201) {
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
    required String groupId,
  }) async {
    final url = Uri.parse("${ApiConfig.baseUrl}/groups/$groupId");

    try {
      final response = await _authenticatedRequest(method: "GET", url: url);

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
    required String groupId,
    required List<String> userIds,
  }) async {
    final url = Uri.parse("${ApiConfig.baseUrl}/groups/$groupId/members");

    try {
      final response = await _authenticatedRequest(
        method: "POST",
        url: url,
        body: {"user_ids": userIds},
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
    required String groupId,
    required List<String> userIds,
  }) async {
    final url = Uri.parse("${ApiConfig.baseUrl}/groups/$groupId/members");

    try {
      final response = await _authenticatedRequest(
        method: "DELETE",
        url: url,
        body: {"user_ids": userIds},
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
    required String email,
  }) async {
    final url = Uri.parse("${ApiConfig.baseUrl}/users/search/email/$email");

    try {
      final response = await _authenticatedRequest(method: "GET", url: url);

      if (response.statusCode == 200) {
        return UserLookupResult.success(
          UserLookup.fromJson(jsonDecode(response.body)),
        );
      }

      if (response.statusCode == 400)
        return UserLookupResult.error("Invalid email");
      if (response.statusCode == 401)
        return UserLookupResult.error("Session expired");
      if (response.statusCode == 404)
        return UserLookupResult.error("User not found");
      if (response.statusCode == 500)
        return UserLookupResult.error("Server error");

      return UserLookupResult.error("Unexpected error");
    } catch (e) {
      return UserLookupResult.error(e.toString());
    }
  }

  // ================= GROUP EXPENSES =================
  static Future<ExpenseListResult> getGroupExpenses({
    required String groupId,
  }) async {
    final url = Uri.parse("${ApiConfig.baseUrl}/groups/$groupId/expenses");

    try {
      final response = await _authenticatedRequest(method: "GET", url: url);

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
    required ExpenseRequest request,
  }) async {
    final url = Uri.parse("${ApiConfig.baseUrl}/groups/$groupId/expenses");

    try {
      final response = await _authenticatedRequest(
        method: "POST",
        url: url,
        body: request.toJson(),
      );

      if (response.statusCode == 201) return BasicResult.success();
      if (response.statusCode == 200) return BasicResult.success();
      if (response.statusCode == 400)
        return BasicResult.error("Split mismatch");
      if (response.statusCode == 401)
        return BasicResult.error("Session expired");
      if (response.statusCode == 403)
        return BasicResult.error("Not group member");
      if (response.statusCode == 500) return BasicResult.error("Server error");

      return BasicResult.error("Unexpected error (${response.statusCode})");
    } catch (e) {
      return BasicResult.error(e.toString());
    }
  }

  // ================= CREATE GUEST USER =================
  static Future<UserLookupResult> createGuestUser({
    required String email,
  }) async {
    final url = Uri.parse("${ApiConfig.baseUrl}/users/guest");
    try {
      final response = await _authenticatedRequest(
        method: "POST",
        url: url,
        body: {"email": email},
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

  static Future<ExpenseDetailResult> getExpenseDetails({
    required String token,
    required String expenseId,
  }) async {
    final url = Uri.parse("${ApiConfig.baseUrl}/expenses/$expenseId");

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
        return ExpenseDetailResult.success(ExpenseDetail.fromJson(data));
      }

      if (response.statusCode == 401) {
        return ExpenseDetailResult.error("Session expired");
      }

      if (response.statusCode == 403) {
        return ExpenseDetailResult.error("Not allowed to view this expense");
      }

      if (response.statusCode == 404) {
        return ExpenseDetailResult.error("Expense not found");
      }

      if (response.statusCode == 500) {
        return ExpenseDetailResult.error("Server error");
      }

      return ExpenseDetailResult.error("Unexpected error");
    } catch (e) {
      return ExpenseDetailResult.error("Unable to connect to server");
    }
  }

  // ================= GROUP SETTLEMENTS =================
  static Future<SettleResult> getGroupSettlements({
    required String token,
    required String groupId,
  }) async {
    final url = Uri.parse("${ApiConfig.baseUrl}/groups/$groupId/settle");

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
        final settlements = data.map((e) => Settlement.fromJson(e)).toList();
        return SettleResult.success(settlements);
      }

      if (response.statusCode == 401)
        return SettleResult.error("Session expired");
      if (response.statusCode == 403)
        return SettleResult.error("Not group member");
      if (response.statusCode == 404)
        return SettleResult.error("Group not found");
      if (response.statusCode == 500) return SettleResult.error("Server error");

      return SettleResult.error("Unexpected error");
    } catch (e) {
      return SettleResult.error(e.toString());
    }
  }

  // ================= SETTLEMENT HISTORY =================
  static Future<SettleResult> getSettlementHistory({
    required String token,
    required String groupId,
  }) async {
    final url = Uri.parse("${ApiConfig.baseUrl}/groups/$groupId/settlements");

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
        final settlements = data.map((e) => Settlement.fromJson(e)).toList();
        return SettleResult.success(settlements);
      }

      if (response.statusCode == 401)
        return SettleResult.error("Session expired");
      if (response.statusCode == 403)
        return SettleResult.error("Not group member");
      if (response.statusCode == 404)
        return SettleResult.error("Group not found");
      if (response.statusCode == 500) return SettleResult.error("Server error");

      return SettleResult.error("Unexpected error");
    } catch (e) {
      return SettleResult.error(e.toString());
    }
  }

  // ================= SETTLEMENT DETAILS =================
  static Future<SettlementDetailResult> getSettlementDetails({
    required String token,
    required String settlementId,
  }) async {
    final url = Uri.parse("${ApiConfig.baseUrl}/settlements/$settlementId");

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
        return SettlementDetailResult.success(Settlement.fromJson(data));
      }

      if (response.statusCode == 401)
        return SettlementDetailResult.error("Session expired");
      if (response.statusCode == 403)
        return SettlementDetailResult.error("Access denied");
      if (response.statusCode == 404)
        return SettlementDetailResult.error("Settlement not found");
      if (response.statusCode == 500)
        return SettlementDetailResult.error("Server error");

      return SettlementDetailResult.error("Unexpected error");
    } catch (e) {
      return SettlementDetailResult.error(e.toString());
    }
  }

  // ================= UPDATE EXPENSE =================
  static Future<ExpenseDetailResult> updateExpense({
    required String token,
    required String expenseId,
    required Map<String, dynamic> body,
  }) async {
    final url = Uri.parse("${ApiConfig.baseUrl}/expenses/$expenseId");

    try {
      final response = await http.put(
        url,
        headers: {
          "Content-Type": "application/json",
          "Authorization": "Bearer $token",
        },
        body: jsonEncode(body),
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        return ExpenseDetailResult.success(ExpenseDetail.fromJson(data));
      }

      if (response.statusCode == 400)
        return ExpenseDetailResult.error("Invalid data or split mismatch");
      if (response.statusCode == 401)
        return ExpenseDetailResult.error("Session expired");
      if (response.statusCode == 403)
        return ExpenseDetailResult.error("No permission to edit");
      if (response.statusCode == 404)
        return ExpenseDetailResult.error("Expense not found");
      if (response.statusCode == 500)
        return ExpenseDetailResult.error("Server error");

      return ExpenseDetailResult.error("Unexpected error");
    } catch (e) {
      return ExpenseDetailResult.error(e.toString());
    }
  }

  // ================= DELETE EXPENSE =================
  static Future<BasicResult> deleteExpense({
    required String token,
    required String expenseId,
  }) async {
    final url = Uri.parse("${ApiConfig.baseUrl}/expenses/$expenseId");

    try {
      final response = await http.delete(
        url,
        headers: {
          "Content-Type": "application/json",
          "Authorization": "Bearer $token",
        },
      );

      if (response.statusCode == 200) return BasicResult.success();

      if (response.statusCode == 401)
        return BasicResult.error("Session expired");
      if (response.statusCode == 403)
        return BasicResult.error("No permission to delete");
      if (response.statusCode == 404)
        return BasicResult.error("Expense not found");
      if (response.statusCode == 500) return BasicResult.error("Server error");

      return BasicResult.error("Unexpected error");
    } catch (e) {
      return BasicResult.error(e.toString());
    }
  }

  // ================= SETTLE PAYMENT =================
  static Future<BasicResult> settlePayment({
    required String token,
    required String groupId,
    required String userId,
    required double amount,
    required String title,
  }) async {
    final url = Uri.parse("${ApiConfig.baseUrl}/groups/$groupId/settle");

    try {
      final response = await http.post(
        url,
        headers: {
          "Content-Type": "application/json",
          "Authorization": "Bearer $token",
        },
        body: jsonEncode({
          "group_id": groupId,
          "user_id": userId,
          "amount": amount,
          "title": title,
        }),
      );

      if (response.statusCode == 201) return BasicResult.success();
      if (response.statusCode == 200) return BasicResult.success();
      if (response.statusCode == 400)
        return BasicResult.error(
          "Cannot settle with yourself or invalid amount",
        );
      if (response.statusCode == 401)
        return BasicResult.error("Session expired");
      if (response.statusCode == 403)
        return BasicResult.error("Not a member of this group");
      if (response.statusCode == 404)
        return BasicResult.error("Group not found");
      if (response.statusCode == 500) return BasicResult.error("Server error");

      return BasicResult.error("Unexpected error (${response.statusCode})");
    } catch (e) {
      return BasicResult.error(e.toString());
    }
  }

  // ================= USER SPENDINGS =================
  static Future<SpendingResult> getUserSpendings({
    required String token,
    required String groupId,
  }) async {
    final url = Uri.parse("${ApiConfig.baseUrl}/groups/$groupId/spendings");

    try {
      final response = await http.get(
        url,
        headers: {
          "Content-Type": "application/json",
          "Authorization": "Bearer $token",
        },
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body) as List;
        final spendings = data.map((e) => UserSpending.fromJson(e)).toList();
        return SpendingResult.success(spendings);
      }

      if (response.statusCode == 401)
        return SpendingResult.error("Session expired");
      if (response.statusCode == 403)
        return SpendingResult.error("Not a group member");
      if (response.statusCode == 404)
        return SpendingResult.error("Group not found");
      if (response.statusCode == 500)
        return SpendingResult.error("Server error");

      return SpendingResult.error("Unexpected error");
    } catch (e) {
      return SpendingResult.error(e.toString());
    }
  }
}
