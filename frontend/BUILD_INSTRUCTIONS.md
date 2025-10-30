# Shared Expenses Flutter App - Build Instructions

## Prerequisites

### Required Software
- **Flutter SDK**: Version 3.9.2 or later
- **FVM** (Flutter Version Management): Already configured in this project
- **Dart SDK**: Comes with Flutter
- **Android Studio** (for Android development)
  - Android SDK
  - Android Emulator
- **Xcode** (for iOS development, macOS only)
- **VS Code** or **Android Studio** (recommended IDEs)

### Backend Requirements
- The Go backend server must be running on `http://localhost:8080`
- See `backend/` directory for backend setup instructions

## Project Structure

```
frontend/
├── lib/
│   ├── config/          # API configuration
│   ├── models/          # Data models (User, Group, Expense)
│   ├── services/        # API and Storage services
│   ├── providers/       # State management (Provider pattern)
│   ├── screens/         # All UI screens
│   ├── widgets/         # Reusable widgets
│   ├── utils/           # Validators and formatters
│   └── main.dart        # App entry point
├── pubspec.yaml         # Dependencies
└── .fvmrc              # FVM configuration
```

## Setup Instructions

### 1. Install Flutter via FVM

If you don't have FVM installed:

```bash
# Install FVM
dart pub global activate fvm

# Use the Flutter version specified in .fvmrc
fvm install
fvm use stable
```

### 2. Install Dependencies

```bash
cd frontend
fvm flutter pub get
```

### 3. Configure API Endpoint

Edit `lib/config/api_config.dart` to match your setup:

```dart
// For Android Emulator (default)
static const String baseUrl = 'http://10.0.2.2:8080';

// For iOS Simulator, change to:
// static const String baseUrl = 'http://localhost:8080';

// For physical device, use your computer's IP:
// static const String baseUrl = 'http://192.168.1.xxx:8080';
```

### 4. Run the Backend Server

Make sure the Go backend is running:

```bash
cd ../backend
go run main.go
# Server should start on http://localhost:8080
```

## Running the Application

### Android Emulator

1. **Start Android Emulator:**
   ```bash
   # List available emulators
   fvm flutter emulators
   
   # Launch an emulator
   fvm flutter emulators --launch <emulator_id>
   ```

2. **Run the app:**
   ```bash
   fvm flutter run
   ```

### iOS Simulator (macOS only)

1. **Start iOS Simulator:**
   ```bash
   open -a Simulator
   ```

2. **Update API config** to use `http://localhost:8080`

3. **Run the app:**
   ```bash
   fvm flutter run
   ```

### Physical Device

1. **Enable USB Debugging** on Android or **Developer Mode** on iOS

2. **Connect device via USB**

3. **Update API config** to use your computer's local IP address

4. **Run:**
   ```bash
   fvm flutter run
   ```

### Web (Development Only)

```bash
# Update baseUrl to http://localhost:8080 in api_config.dart
fvm flutter run -d chrome
```

**Note:** flutter_secure_storage doesn't work perfectly on web. Use for testing only.

## Building for Production

### Android APK

```bash
# Build release APK
fvm flutter build apk --release

# Output: build/app/outputs/flutter-apk/app-release.apk
```

### Android App Bundle (for Google Play)

```bash
# Build app bundle
fvm flutter build appbundle --release

# Output: build/app/outputs/bundle/release/app-release.aab
```

### iOS (macOS only)

```bash
# Build iOS app
fvm flutter build ios --release

# Then open Xcode to archive and distribute
open ios/Runner.xcworkspace
```

## Development Workflow

### Hot Reload

While the app is running, make code changes and press:
- `r` - Hot reload (preserves app state)
- `R` - Hot restart (resets app state)
- `q` - Quit

### Code Analysis

```bash
# Run static analysis
fvm flutter analyze

# Format code
fvm flutter format lib/
```

### Testing

```bash
# Run all tests
fvm flutter test

# Run specific test file
fvm flutter test test/widget_test.dart
```

## Troubleshooting

### Common Issues

#### 1. "Connection refused" or "Network error"

**Problem:** Cannot connect to backend API

**Solutions:**
- Ensure backend is running on port 8080
- Check firewall settings
- For Android emulator, use `10.0.2.2` instead of `localhost`
- For iOS simulator, use `localhost`
- For physical device, use your computer's IP address

**Verify backend is running:**
```bash
curl http://localhost:8080/health
# Should return: ok
```

#### 2. "Failed to load" or package errors

**Solution:**
```bash
fvm flutter clean
fvm flutter pub get
```

#### 3. Android build fails

**Solutions:**
- Update Android SDK: Open Android Studio → SDK Manager → Update
- Accept licenses: `flutter doctor --android-licenses`
- Clear Gradle cache:
  ```bash
  cd android
  ./gradlew clean
  cd ..
  ```

#### 4. iOS build fails (macOS)

**Solutions:**
- Update CocoaPods: `sudo gem install cocoapods`
- Clean pods:
  ```bash
  cd ios
  rm -rf Pods Podfile.lock
  pod install
  cd ..
  ```

#### 5. "flutter_secure_storage" issues

**Android:** Requires minimum SDK 18 (already configured)

**iOS:** Requires iOS 9.0+ (already configured)

**Linux:** May need additional dependencies:
```bash
sudo apt-get install libsecret-1-dev
```

### Debug Mode vs Release Mode

**Debug Mode:** (default with `flutter run`)
- Includes debugging info
- Larger file size
- Slower performance
- Hot reload enabled

**Release Mode:** (with `--release` flag)
- Optimized performance
- Smaller file size
- No debugging info
- Used for production

## Features Implemented

### ✅ Authentication
- User registration
- Login with email/password
- JWT token storage (encrypted)
- Auto-login on app restart
- Logout

### ✅ Group Management
- View all groups
- Create new group
- View group details with members
- Add members (by email search)
- Remove members (admin only)

### ✅ Expense Management
- Create expenses with splits
- View expense details
- Split expenses unevenly
- Equal split quick action
- Mark incomplete amounts/splits
- Delete expenses (by creator or admin)

### ✅ UI/UX
- Material Design 3
- Dark mode support
- Pull-to-refresh
- Loading states
- Error handling
- Empty states

## API Endpoints Used

| Method | Endpoint | Purpose |
|--------|----------|---------|
| POST | `/auth/register` | Register new user |
| POST | `/auth/login` | Login |
| GET | `/auth/me` | Get current user |
| GET | `/groups/me` | List user's groups |
| POST | `/groups/` | Create group |
| GET | `/groups/:id` | Get group details |
| POST | `/groups/:id/members` | Add members |
| DELETE | `/groups/:id/members` | Remove members |
| GET | `/users/search/email/:email` | Search user by email |
| POST | `/expenses/` | Create expense |
| GET | `/expenses/:id` | Get expense details |
| DELETE | `/expenses/:id` | Delete expense |

## Known Limitations

1. **No expense listing by group:** Backend doesn't have `/groups/:id/expenses` endpoint yet
   - Workaround: Expenses are cached locally after creation

2. **No settlement calculation:** Not implemented in backend
   - Planned for future release

3. **No expense filtering:** Backend doesn't support filtering
   - Planned for future release

4. **No expense editing:** Update endpoint exists but UI not implemented
   - Planned for Phase 4

5. **No expense images:** Not in backend schema
   - Planned for future release

## Performance Tips

1. **Reduce app size:**
   ```bash
   # Split APK by ABI
   flutter build apk --split-per-abi
   ```

2. **Enable R8/ProGuard** (Android):
   Already configured in `android/app/build.gradle`

3. **Tree shaking:** Automatically enabled in release builds

## Environment Variables

Currently using hardcoded values in `api_config.dart`.

For production, consider using `flutter_dotenv`:

```yaml
# pubspec.yaml
dependencies:
  flutter_dotenv: ^5.1.0
```

```bash
# .env
API_BASE_URL=https://api.yourapp.com
```

## License

This project is licensed under GPL v3.

## Support

For issues or questions:
1. Check backend logs: `backend/logs/`
2. Check Flutter logs: `fvm flutter logs`
3. Enable debug logging in Dio (see `api_service.dart`)

## Next Steps

1. ✅ **Phase 1:** Authentication & Navigation (COMPLETED)
2. ✅ **Phase 2:** Group Management (COMPLETED)
3. ✅ **Phase 3:** Expense Management (COMPLETED)
4. ⏳ **Phase 4:** Polish & Testing
   - Add expense editing
   - Improve error messages
   - Add unit tests
   - Add widget tests
5. ⏳ **Phase 5:** Future Enhancements
   - Settlement calculation
   - Expense filtering
   - Export data
   - Offline support

## Quick Reference

```bash
# Start backend
cd backend && go run main.go

# Run app (development)
cd frontend && fvm flutter run

# Build APK (production)
fvm flutter build apk --release

# Install on connected device
fvm flutter install

# View logs
fvm flutter logs

# Check for issues
fvm flutter doctor

# Update dependencies
fvm flutter pub upgrade
```

---

**Happy coding! 🚀**
