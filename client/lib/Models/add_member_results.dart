class AddMemberResult {
  final bool isSuccess;
  final String? errorMessage;
  final List<String>? addedMembers;

  AddMemberResult._({
    required this.isSuccess,
    this.errorMessage,
    this.addedMembers,
  });

  factory AddMemberResult.success(List<String> ids) {
    return AddMemberResult._(
      isSuccess: true,
      addedMembers: ids,
    );
  }

  factory AddMemberResult.error(String msg) {
    return AddMemberResult._(
      isSuccess: false,
      errorMessage: msg,
    );
  }
}
