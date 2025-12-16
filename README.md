<!-- Copyright © 2025 Galvanized Logic Inc. -->

# Pure Freecell

Pure Freecell is the classic freecell game written in golang
and used as an example of how to build and ship a game using
the golang [vu](https://github.com/gazed/vu) game engine.

Pure Freecell was written to test the vu engine and to create
a freecell with the minimal feature set enjoyed by the author.

To demonstrate the deploy scripts,
[Pure Freecell](https://apps.apple.com/us/app/pure-freecell/id1659399857)
is available for sale on the IOS App store (as long as the author pays
the yearly apple developer fee).

### Notes

- Generate `./asset/shaders` using `go generate` before compiling using `go build`.
- `go build` adds all assets to the output binary, so recompile after
   changing any asset. 
