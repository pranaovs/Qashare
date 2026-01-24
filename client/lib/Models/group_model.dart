class Group {
  final String groupId;
  final String name;
  final String description;
  final String createdBy;
  final int createdAt;

  Group({
    required this.groupId,
    required this.name,
    required this.description,
    required this.createdBy,
    required this.createdAt,
  });

  factory Group.fromJson(Map<String, dynamic> json) {
    return Group(
      groupId: json["group_id"],
      name: json["name"],
      description: json["description"] ?? "",
      createdBy: json["created_by"],
      createdAt: json["created_at"],
    );
  }
}

class GroupListResult {
  final bool isSuccess;
  final String? errorMessage;
  final List<Group>? groups;

  GroupListResult._({required this.isSuccess, this.errorMessage, this.groups});

  factory GroupListResult.success(List<Group> groups) {
    return GroupListResult._(isSuccess: true, groups: groups);
  }

  factory GroupListResult.error(String msg) {
    return GroupListResult._(isSuccess: false, errorMessage: msg);
  }
}

class GroupCreateResult {
  final bool isSuccess;
  final String? errorMessage;
  final Group? group;

  GroupCreateResult._({required this.isSuccess, this.errorMessage, this.group});

  factory GroupCreateResult.success(Group group) {
    return GroupCreateResult._(isSuccess: true, group: group);
  }

  factory GroupCreateResult.error(String msg) {
    return GroupCreateResult._(isSuccess: false, errorMessage: msg);
  }
}
