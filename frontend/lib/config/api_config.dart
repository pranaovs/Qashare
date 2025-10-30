class ApiConfig {
  // Change to http://localhost:8080 for iOS simulator
  // Change to http://10.0.2.2:8080 for Android emulator
  static const String baseUrl = 'http://10.0.2.2:8080';
  
  // Auth endpoints
  static const String register = '/auth/register';
  static const String login = '/auth/login';
  static const String me = '/auth/me';
  
  // User endpoints
  static const String users = '/users';
  static String userById(String id) => '/users/$id';
  static String searchUserByEmail(String email) => '/users/search/email/$email';
  
  // Group endpoints
  static const String groups = '/groups';
  static const String myGroups = '/groups/me';
  static const String adminGroups = '/groups/admin';
  static String groupById(String id) => '/groups/$id';
  static String groupMembers(String id) => '/groups/$id/members';
  
  // Expense endpoints
  static const String expenses = '/expenses';
  static String expenseById(String id) => '/expenses/$id';
  
  // Timeouts
  static const Duration connectTimeout = Duration(seconds: 30);
  static const Duration receiveTimeout = Duration(seconds: 30);
}
