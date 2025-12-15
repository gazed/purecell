
# package structure

IOS apps are packaged as .ipa files. This are zip compressed archives used
to deliver and install IOS apps.

An example XCode ios game "ipa" package structure that does not use storyboards.
```
├── Payload
│   └── simple-ios.app                 <- app directory
│       ├── _CodeSignature             <- added when signing
│       │   └── CodeResources
│       ├── default.metallib           <- library
│       ├── simple-ios                 <- executable
│       ├── Assets.car                 <- assets catalog compressed archive.
│       ├── embedded.mobileprovision   <- needed for ad-hoc testing on a developer device.
│       ├── Info.plist                 <- see contents using: plutil -p Info.plist
│       └── PkgInfo                    <- 4 byte pacakge type + 4 byte app signature.
```
