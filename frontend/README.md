# Shared Expenses - Flutter Frontend

A FOSS (GPL v3) expense sharing mobile application built with Flutter. An ad-free, privacy-focused alternative to proprietary expense-sharing apps.

## Features

✨ **Core Features:**
- 🔐 Email/password authentication with JWT
- 👥 Create and manage expense groups
- 💰 Add expenses with custom split amounts
- 👤 Search and add members by email
- 🎨 Material Design 3 with dark mode
- 🔒 Secure token storage (encrypted)
- 🚫 No ads, no tracking, completely free

## Screenshots

_Screenshots coming soon..._

## Quick Start

### Prerequisites
- Flutter SDK 3.9.2+
- FVM (recommended)
- Backend server running on http://localhost:8080

### Installation

```bash
# Clone the repository
cd frontend

# Install dependencies
fvm flutter pub get

# Configure API endpoint (if needed)
# Edit lib/config/api_config.dart

# Run the app
fvm flutter run
```

## Documentation

- 📖 **[Build Instructions](BUILD_INSTRUCTIONS.md)** - Complete setup and deployment guide
- 📋 **[Frontend Requirements](../FRONTEND_REQUIREMENTS.md)** - Detailed specifications

## Project Structure

```
lib/
├── config/          # API configuration
├── models/          # Data models
├── services/        # API & Storage services
├── providers/       # State management
├── screens/         # UI screens
├── widgets/         # Reusable components
├── utils/           # Helpers & validators
└── main.dart        # Entry point
```

## Tech Stack

- **Framework:** Flutter 3.9.2+
- **Language:** Dart
- **State Management:** Provider
- **Networking:** Dio
- **Routing:** go_router
- **Storage:** flutter_secure_storage
- **UI:** Material Design 3

## Development

```bash
# Run in debug mode
fvm flutter run

# Run tests
fvm flutter test

# Format code
fvm flutter format lib/

# Analyze code
fvm flutter analyze

# Build release APK
fvm flutter build apk --release
```

## API Integration

This app connects to the Go backend API. Ensure the backend is running before starting the app.

**Default API endpoints:**
- Android Emulator: `http://10.0.2.2:8080`
- iOS Simulator: `http://localhost:8080`
- Physical Device: `http://<your-ip>:8080`

## Contributing

This is a FOSS project! Contributions are welcome.

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Open a Pull Request

## License

This project is licensed under the GPL v3 License - see the [LICENSE](../LICENSE) file for details.

## Roadmap

- [x] Phase 1: Authentication & Navigation
- [x] Phase 2: Group Management
- [x] Phase 3: Expense Management (Basic)
- [ ] Phase 4: Polish & Testing
- [ ] Phase 5: Advanced Features
  - [ ] Settlement calculation
  - [ ] Expense filtering
  - [ ] Data export
  - [ ] Offline support
  - [ ] Push notifications

## Support

For issues or questions, please open an issue on GitHub.

## Acknowledgments

Built with ❤️ using Flutter and open-source libraries.

---

**No ads. No tracking. Just expense sharing.**
