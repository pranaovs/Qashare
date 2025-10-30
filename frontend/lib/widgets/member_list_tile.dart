import 'package:flutter/material.dart';
import '../models/user.dart';
import '../utils/formatters.dart';

class MemberListTile extends StatelessWidget {
  final GroupUser member;
  final bool isAdmin;
  final bool canRemove;
  final VoidCallback? onRemove;

  const MemberListTile({
    super.key,
    required this.member,
    this.isAdmin = false,
    this.canRemove = false,
    this.onRemove,
  });

  @override
  Widget build(BuildContext context) {
    return ListTile(
      leading: CircleAvatar(
        backgroundColor: Theme.of(context).colorScheme.primaryContainer,
        child: Text(
          Formatters.getInitials(member.name),
          style: TextStyle(
            color: Theme.of(context).colorScheme.onPrimaryContainer,
          ),
        ),
      ),
      title: Row(
        children: [
          Text(member.name),
          if (isAdmin) ...[
            const SizedBox(width: 8),
            Chip(
              label: const Text('Admin'),
              visualDensity: VisualDensity.compact,
              backgroundColor: Theme.of(context).colorScheme.primaryContainer,
              labelStyle: TextStyle(
                fontSize: 10,
                color: Theme.of(context).colorScheme.onPrimaryContainer,
              ),
            ),
          ],
        ],
      ),
      subtitle: Text(member.email),
      trailing: canRemove
          ? IconButton(
              icon: const Icon(Icons.remove_circle_outline, color: Colors.red),
              onPressed: onRemove,
            )
          : null,
    );
  }
}
