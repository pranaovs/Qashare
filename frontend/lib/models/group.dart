import 'user.dart';

class Group {
  final String groupId;
  final String name;
  final String? description;
  final String createdBy;
  final int createdAt;
  final List<GroupUser> members;

  Group({
    required this.groupId,
    required this.name,
    this.description,
    required this.createdBy,
    required this.createdAt,
    this.members = const [],
  });

  factory Group.fromJson(Map<String, dynamic> json) {
    return Group(
      groupId: json['group_id'] as String,
      name: json['name'] as String,
      description: json['description'] as String?,
      createdBy: json['created_by'] as String,
      createdAt: json['created_at'] as int,
      members: (json['members'] as List<dynamic>?)
              ?.map((m) => GroupUser.fromJson(m as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'group_id': groupId,
      'name': name,
      'description': description,
      'created_by': createdBy,
      'created_at': createdAt,
      'members': members.map((m) => m.toJson()).toList(),
    };
  }

  DateTime get createdAtDateTime => DateTime.fromMillisecondsSinceEpoch(createdAt * 1000);
  
  int get memberCount => members.length;
}
