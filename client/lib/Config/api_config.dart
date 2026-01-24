class ApiConfig {
  static String _server = "http://devbox:8080";

  static String get baseUrl => _server;

  static void setServer(String server) {
    _server = server;
  }
}
