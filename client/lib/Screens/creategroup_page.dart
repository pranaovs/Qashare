import 'package:flutter/material.dart';
import 'package:qashare/Config/token_storage.dart';
import 'package:qashare/Service/api_service.dart';

class CreategroupPage extends StatefulWidget {
  const CreategroupPage({super.key});

  @override
  State<CreategroupPage> createState() => _CreategroupPageState();
}

class _CreategroupPageState extends State<CreategroupPage> {
  final _formKey = GlobalKey<FormState>();
  final _nameController = TextEditingController();
  final _descriptionController = TextEditingController();

  bool _loading = false;

  @override
  void dispose() {
    _nameController.dispose();
    _descriptionController.dispose();
    super.dispose();
  }

  Future<void> _handleCreate() async {
    if (!_formKey.currentState!.validate()) return;
    setState(() => _loading = true);

    final token = await TokenStorage.getToken();
    if (token == null) {
      setState(() => _loading = false);
      return;
    }
    final result = await ApiService.createGroup(
      token: token,
      name: _nameController.text,
      description: _descriptionController.text,
    );
    setState(() => _loading = false);
    if (result.isSuccess) {
      _showSuccess("Group created");

      Future.delayed(const Duration(milliseconds: 600), () {
        Navigator.pop(context, true); // tells Home to refresh
      });
    } else {
      _showError(result.errorMessage ?? "Failed to create group");
    }
  }

  void _showSuccess(String msg) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(msg),
        duration: const Duration(milliseconds: 900),
        backgroundColor: Theme.of(context).colorScheme.primary,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
        behavior: SnackBarBehavior.floating,
      ),
    );
  }

  void _showError(String msg) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(msg),
        backgroundColor: Theme.of(context).colorScheme.error,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
        duration: const Duration(seconds: 1),
        behavior: SnackBarBehavior.floating,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text("Create Group")),
      body: Padding(
        padding: const EdgeInsets.all(20),
        child: Form(
          key: _formKey,
          child: Column(
            children: [
              TextFormField(
                controller: _nameController,
                decoration: const InputDecoration(labelText: "Group Name"),
                validator: (v) =>
                    v == null || v.trim().isEmpty ? "Enter group name" : null,
              ),
              const SizedBox(height: 15),
              TextFormField(
                controller: _descriptionController,
                decoration: const InputDecoration(
                  labelText: "Description (optional)",
                ),
                maxLines: 2,
              ),
              const SizedBox(height: 30),

              SizedBox(
                width: double.infinity,
                height: 45,
                child: ElevatedButton(
                  onPressed: _loading ? null : _handleCreate,
                  child: _loading
                      ? const SizedBox(
                          height: 20,
                          width: 20,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : const Text("Create Group"),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
