class ApiConfig {
  static String _server = "https://qashare.pranaovs.me";

  static String get baseUrl => "$_server";

  static void setServer(String server) {
    _server = server;
  }
}
