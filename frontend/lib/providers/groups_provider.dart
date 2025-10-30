import 'package:flutter/material.dart';
import '../models/group.dart';
import '../models/user.dart';
import '../services/api_service.dart';

class GroupsProvider with ChangeNotifier {
  final ApiService _apiService = ApiService();

  List<Group> _groups = [];
  Group? _selectedGroup;
  bool _isLoading = false;
  String? _error;

  List<Group> get groups => _groups;
  Group? get selectedGroup => _selectedGroup;
  bool get isLoading => _isLoading;
  String? get error => _error;

  Future<void> loadGroups() async {
    _isLoading = true;
    _error = null;
    notifyListeners();

    try {
      _groups = await _apiService.getMyGroups();
      _isLoading = false;
      notifyListeners();
    } catch (e) {
      _error = _apiService.getErrorMessage(e);
      _isLoading = false;
      notifyListeners();
    }
  }

  Future<void> loadGroup(String groupId) async {
    _isLoading = true;
    _error = null;
    notifyListeners();

    try {
      _selectedGroup = await _apiService.getGroup(groupId);
      _isLoading = false;
      notifyListeners();
    } catch (e) {
      _error = _apiService.getErrorMessage(e);
      _isLoading = false;
      notifyListeners();
    }
  }

  Future<bool> createGroup(String name, String? description) async {
    _isLoading = true;
    _error = null;
    notifyListeners();

    try {
      await _apiService.createGroup(name, description);
      await loadGroups();
      return true;
    } catch (e) {
      _error = _apiService.getErrorMessage(e);
      _isLoading = false;
      notifyListeners();
      return false;
    }
  }

  Future<bool> addMembers(String groupId, List<String> userIds) async {
    try {
      await _apiService.addGroupMembers(groupId, userIds);
      await loadGroup(groupId);
      return true;
    } catch (e) {
      _error = _apiService.getErrorMessage(e);
      notifyListeners();
      return false;
    }
  }

  Future<bool> removeMember(String groupId, String userId) async {
    try {
      await _apiService.removeGroupMembers(groupId, [userId]);
      await loadGroup(groupId);
      return true;
    } catch (e) {
      _error = _apiService.getErrorMessage(e);
      notifyListeners();
      return false;
    }
  }

  Future<User?> searchUserByEmail(String email) async {
    try {
      return await _apiService.searchUserByEmail(email);
    } catch (e) {
      _error = _apiService.getErrorMessage(e);
      notifyListeners();
      return null;
    }
  }

  void clearError() {
    _error = null;
    notifyListeners();
  }

  bool isUserAdmin(String userId) {
    return _selectedGroup?.createdBy == userId;
  }
}
