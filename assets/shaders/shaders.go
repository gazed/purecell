// Copyright © 2024 Galvanized Logic Inc.

// Package shaders contains the builtin engine shaders.
// Run "go generate" to create the shader .spv byte code files.
//
// Shaders are linked directly to the engine code in a few ways:
//
//   - vu/load/shd.go expects specific names for attributes and
//     uniform types. The shd.go code may need to be updated for new shaders.
//   - The order of attributes and uniforms in the shader description
//     files (*.shd) matter and relate directly to the layout values
//     in the shader code.
//   - The shader push constants block only guarantees upto 128 bytes.
//
// PBR shaders are based on the youtube tutorial43 from:
//
//	https://github.com/emeiri/ogldev/
//	https://github.com/emeiri/ogldev/blob/master/Common/Shaders/lighting_new.fs
//	https://github.com/emeiri/ogldev/blob/master/Common/Shaders/lighting_new.vs
package shaders

// =============================================================================
// run "go generate" to create or update the shader byte code.

// 3D shaders
//go:generate glslc board.vert -o board.vert.spv
//go:generate glslc board.frag -o board.frag.spv
//go:generate glslc card.vert -o card.vert.spv
//go:generate glslc card.frag -o card.frag.spv
//go:generate glslc tex3D.vert -o tex3D.vert.spv
//go:generate glslc tex3D.frag -o tex3D.frag.spv

// 2D shaders
//go:generate glslc icon.vert -o icon.vert.spv
//go:generate glslc icon.frag -o icon.frag.spv
//go:generate glslc tint.vert -o tint.vert.spv
//go:generate glslc tint.frag -o tint.frag.spv
