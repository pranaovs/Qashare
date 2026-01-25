// ================= MEMBER MODEL =================

class Member {
  final String userId;
  final String name;
  final String email;
  final bool guest;
  final int joinedAt;

  Member({
    required this.userId,
    required this.name,
    required this.email,
    required this.guest,
    required this.joinedAt,
  });

  factory Member.fromJson(Map<String, dynamic> json) {
    return Member(
      userId: json["user_id"] as String,
      name: json["name"] as String,
      email: json["email"] as String,
      guest: json["guest"] as bool,
      joinedAt: json["joined_at"] as int,
    );
  }
}

// ================= GROUP DETAILS MODEL =================

class GroupDetails {
  final String groupId;
  final String name;
  final String description;
  final String createdBy;
  final int createdAt;
  final List<Member> members;

  GroupDetails({
    required this.groupId,
    required this.name,
    required this.description,
    required this.createdBy,
    required this.createdAt,
    required this.members,
  });

  factory GroupDetails.fromJson(Map<String, dynamic> json) {
    final List membersJson = json["members"] ?? [];

    return GroupDetails(
      groupId: json["group_id"] as String,
      name: json["name"] as String,
      description: json["description"] ?? "",
      createdBy: json["created_by"] as String,
      createdAt: json["created_at"] as int,
      members: membersJson.map((e) => Member.fromJson(e)).toList(),
    );
  }
}

// ================= RESULT WRAPPER =================

class GroupDetailsResult {
  final bool isSuccess;
  final String? errorMessage;
  final GroupDetails? group;

  GroupDetailsResult._({
    required this.isSuccess,
    this.errorMessage,
    this.group,
  });

  factory GroupDetailsResult.success(GroupDetails group) {
    return GroupDetailsResult._(
      isSuccess: true,
      group: group,
    );
  }

  factory GroupDetailsResult.error(String msg) {
    return GroupDetailsResult._(
      isSuccess: false,
      errorMessage: msg,
    );
  }
}
