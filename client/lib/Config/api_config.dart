class ApiConfig {
  static String _server = const String.fromEnvironment(
    'API_HOST',
    defaultValue: 'https://qashare.pranaovs.me',
  );

  static final String _apiBasePath = const String.fromEnvironment(
    'API_BASE_PATH',
    defaultValue: '/api',
  );

  static String get baseUrl => '$_server$_apiBasePath/v1';
  static String get server => _server;

  static void setServer(String server) {
    _server = server;
  }
}
