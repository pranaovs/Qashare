class ApiConfig {
  static String _server = "devbox:8080";

  static String get baseUrl => "http://$_server";

  static void setServer(String server) {
    _server = server;
  }
}
