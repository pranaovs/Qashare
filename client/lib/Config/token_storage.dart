import 'package:flutter_secure_storage/flutter_secure_storage.dart';

class TokenStorage {
  static const _storage = FlutterSecureStorage();
  static const _accessKey = "access_token";
  static const _refreshKey = "refresh_token";

  // ── Access Token ──
  static Future<void> saveAccessToken(String token) async {
    await _storage.write(key: _accessKey, value: token);
  }

  static Future<String?> getAccessToken() async {
    return await _storage.read(key: _accessKey);
  }

  // ── Refresh Token ──
  static Future<void> saveRefreshToken(String token) async {
    await _storage.write(key: _refreshKey, value: token);
  }

  static Future<String?> getRefreshToken() async {
    return await _storage.read(key: _refreshKey);
  }

  // ── Save Both ──
  static Future<void> saveTokens({
    required String accessToken,
    required String refreshToken,
  }) async {
    await Future.wait([
      _storage.write(key: _accessKey, value: accessToken),
      _storage.write(key: _refreshKey, value: refreshToken),
    ]);
  }

  // ── Legacy alias (old code compatibility during transition) ──
  static Future<String?> getToken() async => getAccessToken();
  static Future<void> saveToken(String token) async => saveAccessToken(token);

  // ── Clear All ──
  static Future<void> clear() async {
    await Future.wait([
      _storage.delete(key: _accessKey),
      _storage.delete(key: _refreshKey),
    ]);
  }
}
