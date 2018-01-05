#!/bin/bash
protoc -I pb/ pb/pb.proto --go_out=plugins=grpc:pb
