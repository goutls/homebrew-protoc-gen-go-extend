class ProtocGenGoExtendAT001 < Formula
  desc "Protoc plugin that generates extend standard enums and message code"
  homepage "https://github.com/goutls/protoc-gen-go-extend"
  url "https://api.github.com/repos/goutls/protoc-gen-go-extend/tarball/v0.0.1"
  sha256 "83ce672530c5fac35aa36cf05aab5113661eba349f59e36acc8a0a31acb2244c"
  license "Apache-2.0"
  revision 5

  keg_only :versioned_formula

  livecheck do
    url :stable
    regex(%r{protoc-gen-go-extend/v?(\d+(?:\.\d+)+)}i)
  end

  depends_on "go" => :build
  depends_on "protobuf"

  def install
      cd "protoc-gen-go-extend" do
        system "go", "build", *std_go_args(ldflags: "-s -w")
      end
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