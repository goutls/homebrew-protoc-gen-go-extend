class ProtocGenGoExtendAT0016 < Formula
  desc "Protoc plugin that generates extend standard enums and message code"
  homepage "https://github.com/goutls/protoc-gen-go-extend"
  url "https://api.github.com/repos/goutls/protoc-gen-go-extend/tarball/v0.0.16"
  sha256 "8b3c949a02d1e5438ef02bbc7f3da32a003260ff4731b0e415313d16c566cd6b"
  license "Apache-2.0"
  revision 2

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