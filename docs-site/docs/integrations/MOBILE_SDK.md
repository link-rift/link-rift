# Mobile SDK

> Last Updated: 2025-01-24

Comprehensive guide for integrating Linkrift into iOS, Android, and React Native applications.

---

## Table of Contents

- [Overview](#overview)
- [iOS SDK (Swift)](#ios-sdk-swift)
  - [Installation](#ios-installation)
  - [Configuration](#ios-configuration)
  - [Basic Usage](#ios-basic-usage)
  - [Deep Linking](#ios-deep-linking)
  - [Analytics](#ios-analytics)
- [Android SDK (Kotlin)](#android-sdk-kotlin)
  - [Installation](#android-installation)
  - [Configuration](#android-configuration)
  - [Basic Usage](#android-basic-usage)
  - [Deep Linking](#android-deep-linking)
  - [Analytics](#android-analytics)
- [React Native SDK](#react-native-sdk)
  - [Installation](#rn-installation)
  - [Configuration](#rn-configuration)
  - [Basic Usage](#rn-basic-usage)
  - [Deep Linking](#rn-deep-linking)
- [Deep Linking](#deep-linking)
  - [Universal Links (iOS)](#universal-links-ios)
  - [App Links (Android)](#app-links-android)
  - [Deferred Deep Linking](#deferred-deep-linking)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

---

## Overview

The Linkrift Mobile SDKs provide native integrations for:

- **URL Shortening**: Create and manage short links
- **Deep Linking**: Route users to specific app content
- **Attribution**: Track install sources and conversions
- **Analytics**: Monitor link performance

**Supported Platforms:**
| Platform | Language | Min Version |
|----------|----------|-------------|
| iOS | Swift 5.5+ | iOS 14.0+ |
| Android | Kotlin 1.8+ | API 24+ |
| React Native | TypeScript | 0.70+ |

---

## iOS SDK (Swift)

### iOS Installation

**Swift Package Manager (Recommended):**

```swift
// Package.swift
dependencies: [
    .package(url: "https://github.com/linkrift/linkrift-ios.git", from: "2.0.0")
]
```

Or in Xcode:
1. File > Add Package Dependencies
2. Enter: `https://github.com/linkrift/linkrift-ios.git`
3. Select version 2.0.0 or later

**CocoaPods:**

```ruby
# Podfile
pod 'Linkrift', '~> 2.0'
```

**Carthage:**

```
github "linkrift/linkrift-ios" ~> 2.0
```

### iOS Configuration

```swift
// AppDelegate.swift
import UIKit
import Linkrift

@main
class AppDelegate: UIResponder, UIApplicationDelegate {

    func application(
        _ application: UIApplication,
        didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?
    ) -> Bool {

        // Configure Linkrift SDK
        Linkrift.configure(
            apiKey: "your-api-key",
            options: LinkriftOptions(
                environment: .production,
                logLevel: .warning,
                timeout: 30,
                retryPolicy: .exponentialBackoff(maxRetries: 3)
            )
        )

        // Enable deep link handling
        Linkrift.shared.handleDeepLinks = true

        // Set up attribution callback
        Linkrift.shared.onAttribution = { attribution in
            print("Attribution: \(attribution)")
            // Send to analytics service
        }

        return true
    }

    // Handle Universal Links
    func application(
        _ application: UIApplication,
        continue userActivity: NSUserActivity,
        restorationHandler: @escaping ([UIUserActivityRestoring]?) -> Void
    ) -> Bool {
        guard userActivity.activityType == NSUserActivityTypeBrowsingWeb,
              let url = userActivity.webpageURL else {
            return false
        }

        return Linkrift.shared.handleUniversalLink(url)
    }

    // Handle URL Schemes
    func application(
        _ app: UIApplication,
        open url: URL,
        options: [UIApplication.OpenURLOptionsKey: Any] = [:]
    ) -> Bool {
        return Linkrift.shared.handleURLScheme(url)
    }
}
```

```swift
// SceneDelegate.swift (for iOS 13+)
import UIKit
import Linkrift

class SceneDelegate: UIResponder, UIWindowSceneDelegate {

    func scene(
        _ scene: UIScene,
        continue userActivity: NSUserActivity
    ) {
        guard let url = userActivity.webpageURL else { return }
        Linkrift.shared.handleUniversalLink(url)
    }

    func scene(
        _ scene: UIScene,
        openURLContexts URLContexts: Set<UIOpenURLContext>
    ) {
        guard let url = URLContexts.first?.url else { return }
        Linkrift.shared.handleURLScheme(url)
    }
}
```

### iOS Basic Usage

```swift
import Linkrift

// MARK: - Creating Short Links

class LinkService {

    // Simple URL shortening
    func shortenURL(_ url: String) async throws -> ShortLink {
        return try await Linkrift.shared.createLink(
            originalURL: url
        )
    }

    // With custom options
    func createCustomLink(
        url: String,
        customCode: String?,
        title: String?,
        expiresAt: Date?
    ) async throws -> ShortLink {

        let options = CreateLinkOptions(
            customCode: customCode,
            title: title,
            expiresAt: expiresAt,
            metadata: [
                "source": "ios_app",
                "campaign": "summer_2025"
            ]
        )

        return try await Linkrift.shared.createLink(
            originalURL: url,
            options: options
        )
    }

    // Create deep link
    func createDeepLink(
        path: String,
        params: [String: String],
        fallbackURL: String
    ) async throws -> ShortLink {

        let deepLink = DeepLinkBuilder()
            .setPath(path)
            .setParameters(params)
            .setFallbackURL(fallbackURL)
            .setIOSAppStoreID("123456789")
            .setAndroidPackageName("com.example.app")
            .build()

        return try await Linkrift.shared.createDeepLink(deepLink)
    }
}

// MARK: - Usage in SwiftUI

import SwiftUI

struct ShareView: View {
    @State private var shortURL: String = ""
    @State private var isLoading = false
    @State private var error: Error?

    let originalURL: String

    var body: some View {
        VStack(spacing: 20) {
            if isLoading {
                ProgressView()
            } else if !shortURL.isEmpty {
                Text(shortURL)
                    .font(.headline)

                Button("Copy") {
                    UIPasteboard.general.string = shortURL
                }

                ShareLink(item: URL(string: shortURL)!) {
                    Label("Share", systemImage: "square.and.arrow.up")
                }
            } else {
                Button("Shorten URL") {
                    Task {
                        await shortenURL()
                    }
                }
            }

            if let error = error {
                Text(error.localizedDescription)
                    .foregroundColor(.red)
            }
        }
        .padding()
    }

    private func shortenURL() async {
        isLoading = true
        error = nil

        do {
            let link = try await Linkrift.shared.createLink(originalURL: originalURL)
            shortURL = link.shortURL
        } catch {
            self.error = error
        }

        isLoading = false
    }
}

// MARK: - Resolving Links

extension LinkService {

    func resolveLink(_ shortCode: String) async throws -> LinkDetails {
        return try await Linkrift.shared.resolveLink(shortCode: shortCode)
    }

    func getLinkAnalytics(_ shortCode: String) async throws -> LinkAnalytics {
        return try await Linkrift.shared.getAnalytics(
            shortCode: shortCode,
            period: .last30Days
        )
    }
}
```

### iOS Deep Linking

```swift
// DeepLinkRouter.swift
import UIKit
import Linkrift

protocol DeepLinkHandler {
    func canHandle(path: String) -> Bool
    func handle(path: String, params: [String: String])
}

class DeepLinkRouter {
    static let shared = DeepLinkRouter()

    private var handlers: [DeepLinkHandler] = []

    func register(handler: DeepLinkHandler) {
        handlers.append(handler)
    }

    func route(deepLink: ResolvedDeepLink) {
        guard let handler = handlers.first(where: { $0.canHandle(path: deepLink.path) }) else {
            // Handle unknown deep link
            print("No handler for path: \(deepLink.path)")
            return
        }

        handler.handle(path: deepLink.path, params: deepLink.parameters)
    }
}

// Product detail handler
class ProductDeepLinkHandler: DeepLinkHandler {

    func canHandle(path: String) -> Bool {
        return path.hasPrefix("/product/")
    }

    func handle(path: String, params: [String: String]) {
        let productID = path.replacingOccurrences(of: "/product/", with: "")

        // Navigate to product
        DispatchQueue.main.async {
            let productVC = ProductViewController(productID: productID)
            UIApplication.shared.topViewController?.navigationController?.pushViewController(
                productVC,
                animated: true
            )
        }
    }
}

// Setup in AppDelegate
func setupDeepLinking() {
    DeepLinkRouter.shared.register(handler: ProductDeepLinkHandler())
    DeepLinkRouter.shared.register(handler: ProfileDeepLinkHandler())
    DeepLinkRouter.shared.register(handler: PromotionDeepLinkHandler())

    Linkrift.shared.onDeepLink = { resolvedLink in
        DeepLinkRouter.shared.route(deepLink: resolvedLink)
    }
}
```

### iOS Analytics

```swift
// Track link events
Linkrift.shared.trackEvent(
    shortCode: "abc123",
    event: .click,
    metadata: [
        "screen": "home",
        "position": "banner"
    ]
)

// Get link analytics
Task {
    let analytics = try await Linkrift.shared.getAnalytics(
        shortCode: "abc123",
        period: .last7Days
    )

    print("Total clicks: \(analytics.totalClicks)")
    print("Unique visitors: \(analytics.uniqueVisitors)")

    for country in analytics.clicksByCountry {
        print("\(country.name): \(country.clicks)")
    }
}
```

---

## Android SDK (Kotlin)

### Android Installation

**Gradle (Kotlin DSL):**

```kotlin
// build.gradle.kts (app level)
dependencies {
    implementation("io.linkrift:linkrift-android:2.0.0")
}
```

**Gradle (Groovy):**

```groovy
// build.gradle (app level)
dependencies {
    implementation 'io.linkrift:linkrift-android:2.0.0'
}
```

**Repository setup:**

```kotlin
// settings.gradle.kts
dependencyResolutionManagement {
    repositories {
        mavenCentral()
        maven { url = uri("https://maven.linkrift.io") }
    }
}
```

### Android Configuration

```kotlin
// Application.kt
import android.app.Application
import io.linkrift.Linkrift
import io.linkrift.LinkriftOptions

class MyApplication : Application() {

    override fun onCreate() {
        super.onCreate()

        // Initialize Linkrift SDK
        Linkrift.initialize(
            context = this,
            apiKey = "your-api-key",
            options = LinkriftOptions.Builder()
                .setEnvironment(Environment.PRODUCTION)
                .setLogLevel(LogLevel.WARNING)
                .setTimeout(30_000L)
                .setRetryPolicy(RetryPolicy.ExponentialBackoff(maxRetries = 3))
                .build()
        )

        // Enable App Links handling
        Linkrift.getInstance().enableAppLinks()

        // Set attribution callback
        Linkrift.getInstance().setOnAttributionListener { attribution ->
            Log.d("Linkrift", "Attribution: $attribution")
        }
    }
}
```

```xml
<!-- AndroidManifest.xml -->
<manifest xmlns:android="http://schemas.android.com/apk/res/android">

    <application
        android:name=".MyApplication"
        ...>

        <!-- Deep link activity -->
        <activity
            android:name=".DeepLinkActivity"
            android:exported="true">

            <!-- App Links -->
            <intent-filter android:autoVerify="true">
                <action android:name="android.intent.action.VIEW" />
                <category android:name="android.intent.category.DEFAULT" />
                <category android:name="android.intent.category.BROWSABLE" />
                <data
                    android:scheme="https"
                    android:host="lnkr.ft" />
                <data
                    android:scheme="https"
                    android:host="your-custom-domain.com" />
            </intent-filter>

            <!-- Custom URL Scheme -->
            <intent-filter>
                <action android:name="android.intent.action.VIEW" />
                <category android:name="android.intent.category.DEFAULT" />
                <category android:name="android.intent.category.BROWSABLE" />
                <data
                    android:scheme="linkrift"
                    android:host="open" />
            </intent-filter>
        </activity>

    </application>
</manifest>
```

### Android Basic Usage

```kotlin
import io.linkrift.Linkrift
import io.linkrift.model.*
import kotlinx.coroutines.flow.Flow

// MARK: - Creating Short Links

class LinkRepository {

    private val linkrift = Linkrift.getInstance()

    // Simple URL shortening
    suspend fun shortenUrl(url: String): Result<ShortLink> {
        return linkrift.createLink(originalUrl = url)
    }

    // With custom options
    suspend fun createCustomLink(
        url: String,
        customCode: String? = null,
        title: String? = null,
        expiresAt: Long? = null
    ): Result<ShortLink> {

        val options = CreateLinkOptions(
            customCode = customCode,
            title = title,
            expiresAt = expiresAt,
            metadata = mapOf(
                "source" to "android_app",
                "campaign" to "summer_2025"
            )
        )

        return linkrift.createLink(
            originalUrl = url,
            options = options
        )
    }

    // Create deep link
    suspend fun createDeepLink(
        path: String,
        params: Map<String, String>,
        fallbackUrl: String
    ): Result<ShortLink> {

        val deepLink = DeepLinkBuilder()
            .setPath(path)
            .setParameters(params)
            .setFallbackUrl(fallbackUrl)
            .setAndroidPackageName(BuildConfig.APPLICATION_ID)
            .setIOSAppStoreId("123456789")
            .build()

        return linkrift.createDeepLink(deepLink)
    }
}

// MARK: - Usage in ViewModel

class ShareViewModel(
    private val linkRepository: LinkRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow(ShareUiState())
    val uiState: StateFlow<ShareUiState> = _uiState.asStateFlow()

    fun shortenUrl(originalUrl: String) {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }

            linkRepository.shortenUrl(originalUrl)
                .onSuccess { link ->
                    _uiState.update {
                        it.copy(
                            isLoading = false,
                            shortUrl = link.shortUrl
                        )
                    }
                }
                .onFailure { error ->
                    _uiState.update {
                        it.copy(
                            isLoading = false,
                            error = error.message
                        )
                    }
                }
        }
    }

    fun copyToClipboard(context: Context) {
        val shortUrl = _uiState.value.shortUrl ?: return
        val clipboard = context.getSystemService(Context.CLIPBOARD_SERVICE) as ClipboardManager
        val clip = ClipData.newPlainText("Short URL", shortUrl)
        clipboard.setPrimaryClip(clip)
    }

    fun share(context: Context) {
        val shortUrl = _uiState.value.shortUrl ?: return
        val intent = Intent(Intent.ACTION_SEND).apply {
            type = "text/plain"
            putExtra(Intent.EXTRA_TEXT, shortUrl)
        }
        context.startActivity(Intent.createChooser(intent, "Share link"))
    }
}

data class ShareUiState(
    val isLoading: Boolean = false,
    val shortUrl: String? = null,
    val error: String? = null
)

// MARK: - Jetpack Compose UI

@Composable
fun ShareScreen(
    originalUrl: String,
    viewModel: ShareViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()
    val context = LocalContext.current

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(16.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center
    ) {
        when {
            uiState.isLoading -> {
                CircularProgressIndicator()
            }
            uiState.shortUrl != null -> {
                Text(
                    text = uiState.shortUrl!!,
                    style = MaterialTheme.typography.headlineMedium
                )

                Spacer(modifier = Modifier.height(16.dp))

                Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    Button(onClick = { viewModel.copyToClipboard(context) }) {
                        Text("Copy")
                    }
                    Button(onClick = { viewModel.share(context) }) {
                        Text("Share")
                    }
                }
            }
            else -> {
                Button(onClick = { viewModel.shortenUrl(originalUrl) }) {
                    Text("Shorten URL")
                }
            }
        }

        uiState.error?.let { error ->
            Spacer(modifier = Modifier.height(8.dp))
            Text(
                text = error,
                color = MaterialTheme.colorScheme.error
            )
        }
    }
}
```

### Android Deep Linking

```kotlin
// DeepLinkActivity.kt
import android.content.Intent
import android.os.Bundle
import androidx.appcompat.app.AppCompatActivity
import io.linkrift.Linkrift

class DeepLinkActivity : AppCompatActivity() {

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        handleIntent(intent)
    }

    override fun onNewIntent(intent: Intent?) {
        super.onNewIntent(intent)
        intent?.let { handleIntent(it) }
    }

    private fun handleIntent(intent: Intent) {
        val uri = intent.data ?: return

        Linkrift.getInstance().resolveDeepLink(uri) { result ->
            result
                .onSuccess { resolvedLink ->
                    routeDeepLink(resolvedLink)
                }
                .onFailure { error ->
                    // Handle error or open fallback URL
                    openFallback(uri)
                }

            finish()
        }
    }

    private fun routeDeepLink(resolvedLink: ResolvedDeepLink) {
        val intent = when {
            resolvedLink.path.startsWith("/product/") -> {
                val productId = resolvedLink.path.removePrefix("/product/")
                Intent(this, ProductActivity::class.java).apply {
                    putExtra("product_id", productId)
                }
            }
            resolvedLink.path.startsWith("/profile/") -> {
                val userId = resolvedLink.path.removePrefix("/profile/")
                Intent(this, ProfileActivity::class.java).apply {
                    putExtra("user_id", userId)
                }
            }
            else -> {
                Intent(this, MainActivity::class.java)
            }
        }

        startActivity(intent)
    }

    private fun openFallback(uri: Uri) {
        val browserIntent = Intent(Intent.ACTION_VIEW, uri)
        startActivity(browserIntent)
    }
}
```

### Android Analytics

```kotlin
// Track link events
Linkrift.getInstance().trackEvent(
    shortCode = "abc123",
    event = LinkEvent.CLICK,
    metadata = mapOf(
        "screen" to "home",
        "position" to "banner"
    )
)

// Get link analytics
viewModelScope.launch {
    Linkrift.getInstance()
        .getAnalytics(shortCode = "abc123", period = Period.LAST_7_DAYS)
        .onSuccess { analytics ->
            Log.d("Analytics", "Total clicks: ${analytics.totalClicks}")
            Log.d("Analytics", "Unique visitors: ${analytics.uniqueVisitors}")

            analytics.clicksByCountry.forEach { country ->
                Log.d("Analytics", "${country.name}: ${country.clicks}")
            }
        }
}
```

---

## React Native SDK

### RN Installation

```bash
# npm
npm install @linkrift/react-native

# yarn
yarn add @linkrift/react-native

# iOS: Install pods
cd ios && pod install
```

### RN Configuration

```typescript
// App.tsx
import { useEffect } from 'react';
import { Linkrift } from '@linkrift/react-native';

export default function App() {
  useEffect(() => {
    // Initialize SDK
    Linkrift.initialize({
      apiKey: 'your-api-key',
      environment: 'production',
    });

    // Set up deep link listener
    const subscription = Linkrift.onDeepLink((link) => {
      console.log('Deep link received:', link);
      handleDeepLink(link);
    });

    return () => {
      subscription.remove();
    };
  }, []);

  return <NavigationContainer>{/* Your app */}</NavigationContainer>;
}
```

**iOS Configuration:**

```swift
// ios/YourApp/AppDelegate.swift
import LinkriftReactNative

@objc class AppDelegate: RCTAppDelegate {

  override func application(
    _ application: UIApplication,
    continue userActivity: NSUserActivity,
    restorationHandler: @escaping ([UIUserActivityRestoring]?) -> Void
  ) -> Bool {
    return LinkriftReactNative.handleUniversalLink(userActivity)
  }

  override func application(
    _ app: UIApplication,
    open url: URL,
    options: [UIApplication.OpenURLOptionsKey: Any] = [:]
  ) -> Bool {
    return LinkriftReactNative.handleURLScheme(url)
  }
}
```

**Android Configuration:**

```kotlin
// android/app/src/main/java/com/yourapp/MainActivity.kt
import io.linkrift.reactnative.LinkriftModule

class MainActivity : ReactActivity() {

    override fun onNewIntent(intent: Intent?) {
        super.onNewIntent(intent)
        LinkriftModule.handleIntent(intent)
    }
}
```

### RN Basic Usage

```typescript
// hooks/useLinkrift.ts
import { useState } from 'react';
import { Linkrift, ShortLink, CreateLinkOptions } from '@linkrift/react-native';

export function useShortenUrl() {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const [shortLink, setShortLink] = useState<ShortLink | null>(null);

  const shorten = async (url: string, options?: CreateLinkOptions) => {
    setIsLoading(true);
    setError(null);

    try {
      const link = await Linkrift.createLink(url, options);
      setShortLink(link);
      return link;
    } catch (err) {
      setError(err as Error);
      throw err;
    } finally {
      setIsLoading(false);
    }
  };

  return { shorten, shortLink, isLoading, error };
}

// components/ShareButton.tsx
import React from 'react';
import { View, Text, TouchableOpacity, Share, Clipboard } from 'react-native';
import { useShortenUrl } from '../hooks/useLinkrift';

interface ShareButtonProps {
  url: string;
}

export function ShareButton({ url }: ShareButtonProps) {
  const { shorten, shortLink, isLoading, error } = useShortenUrl();

  const handleShorten = async () => {
    try {
      await shorten(url);
    } catch (err) {
      console.error('Failed to shorten URL:', err);
    }
  };

  const handleCopy = () => {
    if (shortLink) {
      Clipboard.setString(shortLink.shortUrl);
      // Show toast
    }
  };

  const handleShare = async () => {
    if (shortLink) {
      await Share.share({
        message: shortLink.shortUrl,
        url: shortLink.shortUrl,
      });
    }
  };

  if (isLoading) {
    return <ActivityIndicator />;
  }

  if (shortLink) {
    return (
      <View style={styles.container}>
        <Text style={styles.shortUrl}>{shortLink.shortUrl}</Text>
        <View style={styles.actions}>
          <TouchableOpacity onPress={handleCopy} style={styles.button}>
            <Text>Copy</Text>
          </TouchableOpacity>
          <TouchableOpacity onPress={handleShare} style={styles.button}>
            <Text>Share</Text>
          </TouchableOpacity>
        </View>
      </View>
    );
  }

  return (
    <TouchableOpacity onPress={handleShorten} style={styles.shortenButton}>
      <Text>Shorten URL</Text>
    </TouchableOpacity>
  );
}
```

### RN Deep Linking

```typescript
// navigation/DeepLinkConfig.ts
import { LinkingOptions } from '@react-navigation/native';

export const linking: LinkingOptions<RootStackParamList> = {
  prefixes: ['https://lnkr.ft', 'linkrift://'],
  config: {
    screens: {
      Product: 'product/:id',
      Profile: 'profile/:userId',
      Promotion: 'promo/:code',
      Home: '*',
    },
  },
  // Custom deep link resolution
  getStateFromPath: (path, options) => {
    // Handle Linkrift deep links
    return Linkrift.resolveDeepLinkPath(path, options);
  },
};

// App.tsx
import { NavigationContainer } from '@react-navigation/native';
import { linking } from './navigation/DeepLinkConfig';

function App() {
  return (
    <NavigationContainer linking={linking}>
      <RootNavigator />
    </NavigationContainer>
  );
}
```

---

## Deep Linking

### Universal Links (iOS)

1. **Create apple-app-site-association file:**

```json
{
  "applinks": {
    "apps": [],
    "details": [
      {
        "appID": "TEAM_ID.com.yourapp.bundle",
        "paths": ["*"]
      }
    ]
  }
}
```

2. **Host at:** `https://your-domain.com/.well-known/apple-app-site-association`

3. **Add Associated Domains capability in Xcode:**
   - `applinks:your-domain.com`
   - `applinks:lnkr.ft`

### App Links (Android)

1. **Create assetlinks.json:**

```json
[
  {
    "relation": ["delegate_permission/common.handle_all_urls"],
    "target": {
      "namespace": "android_app",
      "package_name": "com.yourapp",
      "sha256_cert_fingerprints": [
        "SHA256:YOUR:CERT:FINGERPRINT"
      ]
    }
  }
]
```

2. **Host at:** `https://your-domain.com/.well-known/assetlinks.json`

3. **Add intent filter with autoVerify in AndroidManifest.xml**

### Deferred Deep Linking

Handle deep links for users who don't have the app installed:

```typescript
// iOS
Linkrift.shared.onDeferredDeepLink = { link, isFirstOpen in
    if isFirstOpen {
        // User just installed the app from this link
        Analytics.track("app_install_from_link", [
            "campaign": link.campaign,
            "source": link.source
        ])
    }

    // Route to content
    DeepLinkRouter.shared.route(deepLink: link)
}

// Android
Linkrift.getInstance().setOnDeferredDeepLinkListener { link, isFirstOpen ->
    if (isFirstOpen) {
        Analytics.track("app_install_from_link", mapOf(
            "campaign" to link.campaign,
            "source" to link.source
        ))
    }

    routeDeepLink(link)
}
```

---

## Best Practices

1. **Initialize Early**: Initialize the SDK in Application/AppDelegate before any deep links arrive

2. **Handle Errors Gracefully**: Always provide fallback behavior when links fail to resolve

3. **Cache Links**: Cache frequently used short links to reduce API calls

4. **Track Attribution**: Use attribution data to measure campaign effectiveness

5. **Test Deep Links**: Test all deep link paths in development and staging

6. **Use Custom Domains**: Use branded domains for better click-through rates

---

## Troubleshooting

**Links not resolving:**
- Verify API key is correct
- Check network connectivity
- Ensure SDK is initialized before handling links

**Deep links not opening app:**
- Verify Universal Links/App Links configuration
- Check domain verification status
- Test with `adb shell am start` (Android) or Safari (iOS)

**Attribution not tracking:**
- Ensure `onAttribution` callback is set before first app open
- Check that deferred deep linking is enabled
- Verify install referrer is available

For more help, visit [developers.linkrift.io](https://developers.linkrift.io) or open an issue on GitHub.
