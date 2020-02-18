package pb

//go:generate protoc --go_out=plugins=grpc:. pb.proto
// errorcode.proto error.proto timestamp.proto biguint.proto

//go:generate go tool fix pb.pb.go
// errorcode.pb.go error.pb.go timestamp.pb.go biguint.pb.go
