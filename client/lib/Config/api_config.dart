class ApiConfig {
  static String _server = "https://qashare.pranaovs.me";
  static const String _apiBasePath = String.fromEnvironment(
    'API_BASE_PATH',
    defaultValue: '/api',
  );

  static String get baseUrl => "$_server$_apiBasePath/v1";

  static void setServer(String server) {
    _server = server;
  }
}
