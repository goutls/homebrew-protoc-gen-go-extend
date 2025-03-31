class ProtocGenGoExtend < Formula
  desc "Protoc plugin that generates extend standard enums and message code"
  homepage "https://github.com/goutls/protoc-gen-go-extend"
  url "https://api.github.com/repos/goutls/protoc-gen-go-extend/tarball/v0.0.17"
  sha256 "66d2b43ad740a7227b07f3deee6d1f50914d04651952a2ec247004b81558fa79"
  license "Apache-2.0"
  revision 1

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