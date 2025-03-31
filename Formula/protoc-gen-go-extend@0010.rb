class ProtocGenGoExtendAT0010 < Formula
  desc "Protoc plugin that generates extend standard enums and message code"
  homepage "https://github.com/goutls/protoc-gen-go-extend"
  url "https://api.github.com/repos/goutls/protoc-gen-go-extend/tarball/v0.0.10"
  sha256 "b5091a354a85e9f1f6071b07810a297610a8ce72fc65ee6f8fbed2693cf4dc03"
  license "Apache-2.0"
  revision 3

  keg_only :versioned_formula

  livecheck do
    url :stable
    regex(%r{cmd/protoc-gen-go-extend/v?(\d+(?:\.\d+)+)}i)
  end

  depends_on "go" => :build
  depends_on "protobuf"

  def install
      system "go", "build", *std_go_args(ldflags: "-s -w", output: bin/"protoc-gen-go-extend"), "./protoc-gen-go-extend/"
  end

  test do
    (testpath/"service.proto").write <<~PROTO
      syntax = "proto3";

      option go_package = ".;proto";

      service Greeter {
        rpc Hello(HelloRequest) returns (HelloResponse);
      }

      message HelloRequest {}
      message HelloResponse {}
    PROTO

    system "protoc", "--plugin=#{bin}/protoc-gen-go-extend", "--go-extend_out=.", "service.proto"

    assert_path_exists testpath/"service_grpc.pb.go"
  end
end