import 'package:dio/dio.dart';
import '../config/api_config.dart';
import '../models/user.dart';
import '../models/group.dart';
import '../models/expense.dart';
import '../models/settlement.dart';
import 'storage_service.dart';

class ApiService {
  late Dio _dio;
  final StorageService _storage = StorageService();
  String? _cachedBaseUrl;
  bool _initialized = false;
  Future<void>? _initFuture;

  ApiService() {
    _dio = Dio(BaseOptions(
      baseUrl: ApiConfig.baseUrl,
      connectTimeout: ApiConfig.connectTimeout,
      receiveTimeout: ApiConfig.receiveTimeout,
      headers: {
        'Content-Type': 'application/json',
      },
    ));

    _dio.interceptors.add(InterceptorsWrapper(
      onRequest: (options, handler) async {
        final token = await _storage.getToken();
        if (token != null) {
          options.headers['Authorization'] = 'Bearer $token';
        }
        return handler.next(options);
      },
      onError: (error, handler) {
        return handler.next(error);
      },
    ));
  }

  Future<void> ensureInitialized() async {
    if (_initialized) return;
    
    // Prevent multiple simultaneous initializations
    if (_initFuture != null) {
      return _initFuture;
    }
    
    _initFuture = _initializeBaseUrl();
    await _initFuture;
    _initFuture = null;
  }

  Future<void> _initializeBaseUrl() async {
    if (_initialized) return;
    
    final serverUrl = await _storage.getServerUrl();
    if (serverUrl != null && serverUrl != _dio.options.baseUrl) {
      _dio.options.baseUrl = serverUrl;
      _cachedBaseUrl = serverUrl;
    } else {
      _cachedBaseUrl = ApiConfig.baseUrl;
    }
    _initialized = true;
  }

  Future<void> updateBaseUrl(String url) async {
    await _storage.saveServerUrl(url);
    _dio.options.baseUrl = url;
    _cachedBaseUrl = url;
    _initialized = true;
  }

  Future<String> getCurrentBaseUrl() async {
    await ensureInitialized();
    return _cachedBaseUrl ?? ApiConfig.baseUrl;
  }

  // Authentication
  Future<String> register(String name, String email, String password) async {
    try {
      final response = await _dio.post(ApiConfig.register, data: {
        'name': name,
        'email': email,
        'password': password,
      });
      return response.data['user_id'] as String;
    } catch (e) {
      rethrow;
    }
  }

  Future<String> login(String email, String password) async {
    try {
      final response = await _dio.post(ApiConfig.login, data: {
        'email': email,
        'password': password,
      });
      return response.data['token'] as String;
    } catch (e) {
      rethrow;
    }
  }

  Future<User> getCurrentUser() async {
    try {
      final response = await _dio.get(ApiConfig.me);
      return User.fromJson(response.data as Map<String, dynamic>);
    } catch (e) {
      rethrow;
    }
  }

  // Users
  Future<User> getUser(String userId) async {
    try {
      final response = await _dio.get(ApiConfig.userById(userId));
      return User.fromJson(response.data as Map<String, dynamic>);
    } catch (e) {
      rethrow;
    }
  }

  Future<User> searchUserByEmail(String email) async {
    try {
      final response = await _dio.get(ApiConfig.searchUserByEmail(email));
      return User.fromJson(response.data as Map<String, dynamic>);
    } catch (e) {
      rethrow;
    }
  }

  // Groups
  Future<List<Group>> getMyGroups() async {
    try {
      final response = await _dio.get(ApiConfig.myGroups);
      if (response.data == null) {
        return [];
      }
      return (response.data as List)
          .map((g) => Group.fromJson(g as Map<String, dynamic>))
          .toList();
    } catch (e) {
      rethrow;
    }
  }

  Future<List<Group>> getAdminGroups() async {
    try {
      final response = await _dio.get(ApiConfig.adminGroups);
      if (response.data == null) {
        return [];
      }
      return (response.data as List)
          .map((g) => Group.fromJson(g as Map<String, dynamic>))
          .toList();
    } catch (e) {
      rethrow;
    }
  }

  Future<Group> getGroup(String groupId) async {
    try {
      final response = await _dio.get(ApiConfig.groupById(groupId));
      return Group.fromJson(response.data as Map<String, dynamic>);
    } catch (e) {
      rethrow;
    }
  }

  Future<String> createGroup(String name, String? description) async {
    try {
      final response = await _dio.post(ApiConfig.groups, data: {
        'name': name,
        'description': description,
      });
      return response.data['group_id'] as String;
    } catch (e) {
      rethrow;
    }
  }

  Future<void> addGroupMembers(String groupId, List<String> userIds) async {
    try {
      await _dio.post(ApiConfig.groupMembers(groupId), data: {
        'user_ids': userIds,
      });
    } catch (e) {
      rethrow;
    }
  }

  Future<void> removeGroupMembers(String groupId, List<String> userIds) async {
    try {
      await _dio.delete(ApiConfig.groupMembers(groupId), data: {
        'user_ids': userIds,
      });
    } catch (e) {
      rethrow;
    }
  }

  Future<List<Settlement>> getGroupSettlements(String groupId) async {
    try {
      final response = await _dio.get(ApiConfig.groupSettlements(groupId));
      if (response.data == null) {
        return [];
      }
      return (response.data as List)
          .map((s) => Settlement.fromJson(s as Map<String, dynamic>))
          .toList();
    } catch (e) {
      rethrow;
    }
  }

  // Expenses
  Future<String> createExpense(ExpenseRequest expense) async {
    try {
      final response = await _dio.post(ApiConfig.expenses, data: expense.toJson());
      return response.data['expense_id'] as String;
    } catch (e) {
      rethrow;
    }
  }

  Future<Expense> getExpense(String expenseId) async {
    try {
      final response = await _dio.get(ApiConfig.expenseById(expenseId));
      return Expense.fromJson(response.data as Map<String, dynamic>);
    } catch (e) {
      rethrow;
    }
  }

  Future<void> updateExpense(String expenseId, ExpenseRequest expense) async {
    try {
      await _dio.put(ApiConfig.expenseById(expenseId), data: expense.toJson());
    } catch (e) {
      rethrow;
    }
  }

  Future<void> deleteExpense(String expenseId) async {
    try {
      await _dio.delete(ApiConfig.expenseById(expenseId));
    } catch (e) {
      rethrow;
    }
  }

  String? getErrorMessage(dynamic error) {
    if (error is DioException) {
      if (error.response != null) {
        final data = error.response!.data;
        if (data is Map<String, dynamic> && data.containsKey('error')) {
          return data['error'] as String;
        }
        return 'Server error: ${error.response!.statusCode}';
      } else if (error.type == DioExceptionType.connectionTimeout ||
          error.type == DioExceptionType.receiveTimeout) {
        return 'Connection timeout';
      } else if (error.type == DioExceptionType.connectionError) {
        return 'No internet connection';
      }
      return 'Network error';
    }
    return error.toString();
  }
}
