class ApiConfig {
  static String _server = "qashare.devserver.ts.net:80";

  static String get baseUrl => "http://$_server";

  static void setServer(String server) {
    _server = server;
  }
}
