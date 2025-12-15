package main

// deploy.go runs the commands needed to package the game
// for the various app stores.

import (
	"errors"
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

var (
	app             = "PureFreecell"    // Application name.
	appPkg          = app + ".pkg"      // App package name.
	appApp          = app + ".app"      // App app name.
	appVer          = "1.2.0"           // App version - also in Info.plist
	dirMode         = os.FileMode(0755) // default directory permissions.
	IOSMinVersion   = "16.0"            // iOS 16 released 2022
	MACOSMinVersion = "14.0"            // macOS14 "Sonoma" released 2023
	WINMinVersion   = "xx.0"            // TODO, likely windows 10.0
	vulkanMacOS     = os.Getenv("VULKAN_SDK") + "/macOS"
	vulkanIOS       = os.Getenv("VULKAN_SDK") + "/iOS"

	// IDs for keychain certificate keys are defined outside the deploy script.
	// Verify certificates are available using:
	//   security find-identity -p basic -v
	macosDev  = os.Getenv("MACOS_DEV")  // macos Developer ID Application:
	macosInst = os.Getenv("MACOS_INST") // macos Developer ID Installer:
	appleDev  = os.Getenv("APPLE_DEV")  // ios
	appleDist = os.Getenv("APPLE_DIST") // ios distribution.

	// define the apple provisioning profile locations outside the deploy script
	devProfile      = os.Getenv("PureFreecellDevProfile")
	macStoreProfile = os.Getenv("PureFreecellMacStoreProfile")
	macDevelProfile = os.Getenv("PureFreecellMacDevelProfile")
	iosStoreProfile = os.Getenv("PureFreecellIOSStoreProfile")
)

// deploy creates packages for uploading to app stores.
// Expected to be run from this directory.
// All build output placed in a local 'builds' directory
func main() {
	usage := "usage: deploy [clean|macos|ios|win]"

	// build a deployment package.
	switch {
	case len(os.Args) <= 1:
		println(usage)
	case os.Args[1] == "clean":
		cleanOutput() // remove all generated output
	case os.Args[1] == "macos":
		// must be run on an apple computer that has:
		// o XCode developer tools installed.
		// o Vulkan SDK installed
		packageMACOS()
	case os.Args[1] == "ios":
		// same as macos
		packageIOS()
	case os.Args[1] == "win":
		// expecting an windows computer that has:
		// o Vulkan SDK installed
		packageWINDOWS()
	default:
		println(usage)
	}
}

// cleanOutput removes all generated files.
func cleanOutput() {
	println("Removing builds directory")
	os.RemoveAll("builds")
}

// =============================================================================
// runCmd* is a generic command line runner utility.
// It returns the command output if there were no errors.
func runCmdDir(dir, command string, args ...string) (output string) {
	print("... ..")
	cmd := exec.Command(command, args...)
	cmd.Dir = dir
	return execCmd(cmd)
}
func runCmdEnv(env []string, command string, args ...string) (output string) {
	print("... ..")
	for _, e := range env {
		print(e + " ")
	}
	cmd := exec.Command(command, args...)
	cmd.Env = append(os.Environ(), env...)
	return execCmd(cmd)
}
func runCmd(command string, args ...string) (output string) {
	print("... ..")
	cmd := exec.Command(command, args...)
	return execCmd(cmd)
}
func execCmd(cmd *exec.Cmd) (output string) {
	cmdOut, err := cmd.CombinedOutput()
	println(cmd.String())
	if err != nil {
		slog.Error("... ..runCmd", "output", string(output))
		return ""
	}
	return string(cmdOut)
}

// =============================================================================
// Creates a package for uploading to the mac app store.
// Ref: https://vulkan.lunarg.com/doc/sdk/1.4.328.1/mac/getting_started.html
func packageMACOS() {
	println("packaging macos...")

	// create the apple icon if it does not exist.
	if _, err := os.Stat("icon/PureFreecell.icns"); errors.Is(err, os.ErrNotExist) {
		println("... creating apple icon")
		createAppleIcon()
	}

	// create the OSX application bundle directory structure.
	platform := "builds/macos"
	os.RemoveAll(platform) // remove any existing output
	contents := platform + "/" + appApp + "/Contents"
	os.MkdirAll(contents+"/MacOS", dirMode)
	os.MkdirAll(contents+"/Resources/vulkan", dirMode)
	os.MkdirAll(contents+"/Frameworks", dirMode)

	println("...building macos executable")
	// go build -ldflags="-s -linkmode=external" -o builds/macos/freecell ..
	// add  "--tags", "debug", for a debug build.
	runCmd("go", "build", "-ldflags=-s -linkmode=external", "-o", platform+"/freecell", "..")

	// create the osx application bundle.
	println("...building macos bundle")
	runCmd("cp", "macos/Info.plist", contents)
	runCmd("mv", platform+"/freecell", contents+"/MacOS/PureFreecell")
	runCmd("cp", "PureFreecell.icns", contents+"/Resources/PureFreecell.icns")
	runCmd("cp", "-r", vulkanMacOS+"/share/vulkan/icd.d", contents+"/Resources/vulkan/")
	runCmd("cp", vulkanMacOS+"/lib/libMoltenVK.dylib", contents+"/Frameworks")
	runCmd("cp", vulkanMacOS+"/lib/libvulkan.1.4.321.dylib", contents+"/Frameworks")
	runCmdDir(contents+"/Frameworks", "ln", "-s", "libvulkan.1.4.321.dylib", "libvulkan.1.dylib")

	// set executable rpath to load dylibs from the app bundle.
	runCmd("install_name_tool", "-add_rpath", "@executable_path/../Frameworks", contents+"/MacOS/PureFreecell")

	// sign every executable in the application bundle.
	// Validate compliance using:
	//   codesign -dvvv builds/macos/PureFreecell.app
	//   codesign -dvvv builds/macos/PureFreecell.app/Contents/Frameworks/libMoltenVK.dylib
	runCmd("codesign", "--options", "runtime", "-fv", "-s", macosDev, contents+"/MacOS/PureFreecell")
	runCmd("codesign", "-fv", "-s", macosDev, contents+"/Frameworks/libMoltenVK.dylib")
	runCmd("codesign", "-fv", "-s", macosDev, contents+"/Frameworks/libvulkan.1.4.321.dylib")
	runCmd("codesign", "--options", "runtime", "-fv", "--entitlements", "macos/Entitlements.plist", "-s", macosDev, platform+"/"+appApp)

	// Use the "Developer ID Installer" certificate to create the app package.
	// Validate the package using:
	//   pkgutil --check-signature builds/macos_PureFreecell.pkg
	//   spctl --assess --ignore-cache --verbose --type install builds/macos_store_PureFreecell.pkg
	// Dump to validate the contents using:
	//   pkgutil --expand-full builds/macos_PureFreecell.pkg ./tmp_pkg_dir
	// Install to /Applications
	//   sudo installer -pkg builds/macos_PureFreecell.pkg -target /Applications
	runCmd("pkgbuild", "--version", appVer, "--root", "builds/macos",
		"--sign", macosInst, "--identifier", "com.galvanizedlogic.purefreecell",
		"--install-location", "/Applications",
		"builds/macos_"+appPkg)

	// Create a signed app store submission using productbuild with the "Developer ID Installer" certificate
	println("...packaging for app store")
	runCmd("productbuild", "--version", appVer, "--sign", macosInst,
		"--component", platform+"/"+appApp, "/Applications",
		"builds/macos_store_"+appPkg)

	println("...notarizing app")
	// Notarize the app - using a keychain profile, (my profile is called NotaryTool) see:
	//   https://scriptingosx.com/2021/07/notarize-a-command-line-tool-with-notarytool/
	// The --wait flag tells the tool to hang around until Apple's
	// servers return the notarization information
	// Debug notarytool problems using:
	//   xcrun notarytool submit builds/macos_store_PureFreecell.pkg --keychain-profile NotaryTool --wait
	//   xcrun notarytool log <submission_id> --keychain-profile NotaryTool
	//   pkgutil --check-signature builds/macos_store_PureFreecell.pkg
	runCmd("xcrun", "notarytool", "submit", "builds/macos_store_"+appPkg,
		"--keychain-profile", "NotaryTool", "--wait")
	// xcrun stapler staple builds/macos_store_PureFreecell.pkg
	runCmd("xcrun", "stapler", "staple", "builds/macos_store_"+appPkg)

	// Use the Transporter app from the mac app store to upload macos_store_PureFreecell.pkg.
}

// =============================================================================
// Creates "builds/iosPureFreecell.ipa" for uploading to the ios app store.
// Also see: https://www.khronos.org/blog/developing-with-vulkan-on-apple-ios
func packageIOS() {
	println("packaging ios...")

	// create the apple icon if it does not exist.
	if _, err := os.Stat("icon/PureFreecell.icns"); errors.Is(err, os.ErrNotExist) {
		println("... creating apple icon")
		createAppleIcon()
	}

	// create the ios app bundle directory structure.
	platform := "builds/ios"
	os.RemoveAll(platform)
	os.MkdirAll(platform+"/"+appApp+"/Frameworks", dirMode)
	os.MkdirAll(platform+"/Images.xcassets/AppIcon.appiconset", dirMode)

	println("...building ios executable")
	SDK := strings.TrimSpace(runCmd("xcrun", "--sdk", "iphoneos", "--show-sdk-path"))
	CLANG := strings.TrimSpace(runCmd("xcrun", "--sdk", "iphoneos", "--find", "clang"))
	FLAGS := `-isysroot ` + SDK + ` -arch arm64 -miphoneos-version-min=` + IOSMinVersion

	// The build command should look something like:
	// GOOS=ios GOARCH=arm64 CC=/Applications/Xcode.app/Contents/Developer/Toolchains/XcodeDefault.xctoolchain/usr/bin/clang CXX=/Applications/Xcode.app/Contents/Developer/Toolchains/XcodeDefault.xctoolchain/usr/bin/clang CGO_CFLAGS="-isysroot /Applications/Xcode.app/Contents/Developer/Platforms/iPhoneOS.platform/Developer/SDKs/iPhoneOS26.0.sdk -arch arm64 -miphoneos-version-min=16.0" CGO_LDFLAGS="-isysroot /Applications/Xcode.app/Contents/Developer/Platforms/iPhoneOS.platform/Developer/SDKs/iPhoneOS26.0.sdk -arch arm64 -miphoneos-version-min=16.0" CGO_ENABLED=1 /usr/local/go/bin/go build -ldflags=-s -o builds/ios/freecell ..
	// NOTE: The -mios-version-min must match the app's Info.plist MinimumOSVersion key
	env := []string{"GOOS=ios",
		"GOARCH=arm64",
		"CC=" + CLANG,
		"CXX=" + CLANG,
		"CGO_CFLAGS=" + FLAGS,
		"CGO_LDFLAGS=" + FLAGS,
		"CGO_ENABLED=1",
	}
	// add "--tags", "debug", to get the DEBUG version.
	runCmdEnv(env, "go", "build", "-ldflags=-s", "-o", platform+"/freecell", "..")

	// copy files to ios app directory.
	// and set executable rpath to load dylibs from the app bundle.
	appRoot := platform + "/" + appApp
	runCmd("mv", platform+"/freecell", appRoot+"/PureFreecell")
	runCmd("install_name_tool", "-add_rpath", "@executable_path/Frameworks", appRoot+"/PureFreecell")
	runCmd("xcrun", "copypng", "-compress", "-strip-PNG-text", "ios/Default-568h@2x.png", appRoot+"/Default-568h@2x.png")

	// Compile the asset catalog.
	runCmd("cp", "ios/Contents.json", platform+"/Images.xcassets/AppIcon.appiconset/")
	runCmd("cp", "ios/icon_120x120.png", platform+"/Images.xcassets/AppIcon.appiconset/")
	runCmd("cp", "ios/icon_167x167.png", platform+"/Images.xcassets/AppIcon.appiconset/")
	runCmd("cp", "ios/icon_76x76x2.png", platform+"/Images.xcassets/AppIcon.appiconset/")
	runCmd("cp", "ios/icon_1024x1024.png", platform+"/Images.xcassets/AppIcon.appiconset/")
	runCmd("xcrun", "actool", "--output-format", "human-readable-text", "--notices",
		"--warnings", "--output-partial-info-plist", "ios/assetcatalog.plist", "--app-icon", "AppIcon",
		"--compress-pngs", "--enable-on-demand-resources", "YES", "--target-device", "iphone",
		"--target-device", "ipad", "--minimum-deployment-target", IOSMinVersion, "--platform", "iphoneos",
		"--product-type", "com.apple.product-type.application", "--compile", platform+"/PureFreecell.app", platform+"/Images.xcassets")

	// Copy app contents into the app directory structure
	// Include vulkan frameworks from the VulkanSDK as IOS does not support naked dylibs.
	runCmd("cp", "ios/Info.plist", appRoot+"/Info.plist")
	runCmd("cp", devProfile, appRoot+"/embedded.mobileprovision")
	runCmd("cp", "-R", vulkanIOS+"/lib/MoltenVK.xcframework/ios-arm64/MoltenVK.framework", appRoot+"/Frameworks/")
	os.MkdirAll(appRoot+"/vulkan", dirMode)
	runCmd("cp", "-R", vulkanIOS+"/share/vulkan/icd.d", appRoot+"/vulkan")

	// Create the store package using the app contents before signing.
	// Use the distribution provisioning profile and the distribution entitilements.
	pkgRoot := "builds/ios/" + appPkg
	pkgPay := pkgRoot + "/Payload"
	pkgApp := pkgRoot + "/Payload/" + appApp
	os.RemoveAll(pkgRoot)
	os.MkdirAll(pkgPay, dirMode)
	runCmd("cp", "-r", platform+"/"+appApp, pkgPay)
	runCmd("cp", iosStoreProfile, pkgApp+"/embedded.mobileprovision")

	// Sign the developer app to be able to test on developer devices.
	// Check available codesign identities using:
	// - security find-identity -p codesigning -v login.keychain
	//
	// Test on devices using:
	// - xcrun devicectl list devices (to get deviceID)
	// - xcrun devicectl device install app --device <deviceID> builds/ios/PureFreecell.app
	// or...
	// - xcrun simctl list (to get simulatorID)
	// - xcrun simctl install <simulatorID> builds/ios/PureFreecell.app
	// Check logs using console app for the given device.
	runCmd("codesign", "-fv", "-s", appleDev, appRoot+"/Frameworks/MoltenVK.framework")
	runCmd("codesign", "--options", "runtime", "-f", "--sign", appleDev,
		"--entitlements", "ios/entitlements.plist", "--timestamp=none", appRoot)

	// sign the store upload packagewith the distribution certificate.
	runCmd("codesign", "-fv", "-s", appleDist, pkgApp+"/Frameworks/MoltenVK.framework")
	runCmd("codesign", "--options", "runtime", "-fv", "-s", appleDist,
		"--entitlements", "ios/entitlements-dist.plist",
		"--preserve-metadata=identifier,flags", pkgApp)

	// Create the ipa using ditto instead of zip.
	// Upload the ipa to the app store using macos Transporter.
	runCmd("ditto", "-V", "-c", "-k", "--norsrc", pkgRoot, platform+app+".ipa")
}

// createAppleIcon uses apple xcode developer tools to create
// an apple icon file from an image. See:
// https://stackoverflow.com/questions/12306223/how-to-manually-create-icns-files-using-iconutil
func createAppleIcon() {
	iconImage := "icon/freecellIcon1024.png" // expecting png.
	iconDir := "PureFreecell.iconset"        // must be appName.iconset
	os.MkdirAll(iconDir, dirMode)
	runCmd("sips", "-z", "16", "16", iconImage, "--out", iconDir+"/icon_16x16.png")
	runCmd("sips", "-z", "32", "32", iconImage, "--out", iconDir+"/icon_16x16@2x.png")
	runCmd("sips", "-z", "32", "32", iconImage, "--out", iconDir+"/icon_32x32.png")
	runCmd("sips", "-z", "64", "64", iconImage, "--out", iconDir+"/icon_32x32@2x.png")
	runCmd("sips", "-z", "128", "128", iconImage, "--out", iconDir+"/icon_128x128.png")
	runCmd("sips", "-z", "256", "256", iconImage, "--out", iconDir+"/icon_128x128@2x.png")
	runCmd("sips", "-z", "256", "256", iconImage, "--out", iconDir+"/icon_256x256.png")
	runCmd("sips", "-z", "512", "512", iconImage, "--out", iconDir+"/icon_256x256@2x.png")
	runCmd("sips", "-z", "512", "512", iconImage, "--out", iconDir+"/icon_512x512.png")
	runCmd("cp", iconImage, iconDir+"/icon_512x512@2x.png")
	runCmd("iconutil", "-c", "icns", "-o", "icon/PureFreecell.icns", iconDir)
	os.RemoveAll(iconDir)
}

// =============================================================================
// Build for releasing a windows app to steam and the windows store:
//  1. update the version number
//     1.1 update version in ./deploy.go
//     1.2 update version in ./win/win_manifest.xml
//     1.3 update version in ./win/AppxManifest.xml
//
// FUTURE: ship on windows.... currently PureFreecell name is not available.
//
// NOTE: embed windows resources like icons using:
// https://github.com/akavel/rsrc
//   - go install github.com/akavel/rsrc@latest
//   - $GOPATH/bin/rsrc -arch amd64 -ico icon.ico
//
// NOTE: the windows application manifest (win_manifest.xml) is based on:
// https://github.com/microsoft/dotnet-samples/blob/master/WinForms-HDPI/SystemAware/app.manifest
//
// NOTE: download the windows SDK from
// https://developer.microsoft.com/en-us/windows/downloads/windows-sdk/
// The SDK installs to C:\Program Files (x86)\Windows Kits\10\bin\10.0.26100.0\x64
//
// NOTE: makeappx is part of the windows SDK. See:
// https://learn.microsoft.com/en-us/windows/msix/package/create-app-package-with-makeappx-tool
// This generates the .msix package
//
// NOTE: run the "Windows App Cert Kit" (part of the windows SDK). See:
// https://learn.microsoft.com/en-us/windows/win32/win_cert/using-the-windows-app-certification-kit
// This validates the .msix package
func packageWINDOWS() {
	println("FUTURE: packaging windows...")

	// -----------------------------------------------------------------------------
	// clean output directory.
	platform := "builds/win"
	os.RemoveAll(platform) // remove any existing output
	os.MkdirAll(platform, dirMode)

	// -----------------------------------------------------------------------------
	// generate the windows syso file using https://github.com/akavel/rsrc
	// The output win_amd64.syso must be in the build directory
	// and is included automatically in the binary by "go build"
	icon := "icon/freecellIcon512.ico"
	runCmd("rsrc", "-arch", "amd64", "-ico", icon, "-manifest", "win/win_manifest.xml")
	runCmd("mv", "rsrc_windows_amd64.syso", "../win_amd64.syso")

	// -----------------------------------------------------------------------------
	println("...building windows executable")
	// build a windows exe called ./freecell.exe
	// Get verbose build output by adding -x before -ldflags.
	// To get a debug version add "--tags", "debug"
	//
	// NOTE: https://github.com/golang/go/issues/71242 discusses asyncpreemptoff and freezes w. steam.
	runCmd("go", "build", "-C", "..",
		"-ldflags=-H=windowsgui -X runtime.godebugDefault=asyncpreemptoff=1 -X main.Version="+appVer)

	// -----------------------------------------------------------------------------
	// Create the steam zip file.
	// Upload this to steam as a new build from the steamworks app webpage.
	//
	// NOTE: use the windows tar.exe, not the git bash tar.
	// It seems the bash /usr/bin/tar output is not recognized by windows.
	runCmd("cp", "../freecell.exe", "./PureFreecell.exe")
	env := []string{"PATH=/c/WINDOWS/system32:$PATH"}
	runCmdEnv(env, "tar", "-a", "-cf", platform+"/steam_PureFreecell.zip", "PureFreecell.exe", "OpenAL32.dll")
	runCmd("rm", "./PureFreecell.exe")

	// -----------------------------------------------------------------------------
	// Create the windows store package.
	// Validate using "Windows App Cert Kit" before uploading.
	// Upload from the microsoft product center website on the application overview page - packages section.
	pkg := platform + "/PureFreecell"
	dll := pkg + "/VFS/SystemX64"
	os.MkdirAll(pkg, dirMode)
	os.MkdirAll(dll, dirMode)
	runCmd("mv", "../freecell.exe", pkg+"/PureFreecell.exe")
	runCmd("cp", "OpenAL32.dll", pkg)
	runCmd("cp", "OpenAL32.dll", dll)
	runCmd("cp", "win/AppxManifest.xml", pkg)
	runCmd("cp", "win/gl50x50.png", pkg)
	runCmd("cp", "win/logo150x150.png", pkg)
	runCmd("cp", "win/logo310x150.png", pkg)
	runCmd("cp", "win/logo310x310.png", pkg)
	runCmd("cp", "win/logo44x44.png", pkg)
	runCmd("makeappx.exe", "pack", "-d", pkg, "-p", platform+"/win_PureFreecell.msix")
}
