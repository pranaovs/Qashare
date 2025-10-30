import 'package:flutter/material.dart';
import '../models/expense.dart';
import '../services/api_service.dart';

class ExpensesProvider with ChangeNotifier {
  final ApiService _apiService = ApiService();

  final Map<String, List<Expense>> _expensesByGroup = {};
  Expense? _selectedExpense;
  bool _isLoading = false;
  String? _error;

  List<Expense> getExpensesForGroup(String groupId) {
    return _expensesByGroup[groupId] ?? [];
  }

  Expense? get selectedExpense => _selectedExpense;
  bool get isLoading => _isLoading;
  String? get error => _error;

  Future<void> loadExpense(String expenseId) async {
    _isLoading = true;
    _error = null;
    notifyListeners();

    try {
      _selectedExpense = await _apiService.getExpense(expenseId);
      _isLoading = false;
      notifyListeners();
    } catch (e) {
      _error = _apiService.getErrorMessage(e);
      _isLoading = false;
      notifyListeners();
    }
  }

  Future<bool> createExpense(ExpenseRequest expense) async {
    _isLoading = true;
    _error = null;
    notifyListeners();

    try {
      final expenseId = await _apiService.createExpense(expense);
      // Load the created expense and add to cache
      final createdExpense = await _apiService.getExpense(expenseId);
      _addExpenseToCache(createdExpense);
      _isLoading = false;
      notifyListeners();
      return true;
    } catch (e) {
      _error = _apiService.getErrorMessage(e);
      _isLoading = false;
      notifyListeners();
      return false;
    }
  }

  Future<bool> updateExpense(String expenseId, ExpenseRequest expense) async {
    _isLoading = true;
    _error = null;
    notifyListeners();

    try {
      await _apiService.updateExpense(expenseId, expense);
      await loadExpense(expenseId);
      return true;
    } catch (e) {
      _error = _apiService.getErrorMessage(e);
      _isLoading = false;
      notifyListeners();
      return false;
    }
  }

  Future<bool> deleteExpense(String expenseId, String groupId) async {
    _isLoading = true;
    _error = null;
    notifyListeners();

    try {
      await _apiService.deleteExpense(expenseId);
      _removeExpenseFromCache(expenseId, groupId);
      _isLoading = false;
      notifyListeners();
      return true;
    } catch (e) {
      _error = _apiService.getErrorMessage(e);
      _isLoading = false;
      notifyListeners();
      return false;
    }
  }

  void _addExpenseToCache(Expense expense) {
    if (!_expensesByGroup.containsKey(expense.groupId)) {
      _expensesByGroup[expense.groupId] = [];
    }
    _expensesByGroup[expense.groupId]!.add(expense);
  }

  void _removeExpenseFromCache(String expenseId, String groupId) {
    _expensesByGroup[groupId]?.removeWhere((e) => e.expenseId == expenseId);
  }

  void clearError() {
    _error = null;
    notifyListeners();
  }
}
