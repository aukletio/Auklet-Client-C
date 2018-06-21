// Package generate is not intended to be compiled. It exists so that protoc,
// called by go generate, will run in the project root and thus resolve import
// paths in .proto files correctly.
//
// go generate must be run whenever any .proto files are changed. The generated
// code should be checked in.
package generate

//go:generate protoc ./device/device.proto --go_out=../../../
//go:generate protoc ./schema/schema.proto --go_out=../../../
