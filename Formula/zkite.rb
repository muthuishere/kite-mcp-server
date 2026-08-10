# typed: false
# frozen_string_literal: true

class Zkite < Formula
  desc "CLI and MCP server for the Kite Connect trading API"
  homepage "https://github.com/zerodha/kite-mcp-server"
  version "0.1.0"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/zerodha/kite-mcp-server/releases/download/v#{version}/zkite-darwin-arm64.tar.gz"
      sha256 "REPLACE_WITH_SHA256_FOR_DARWIN_ARM64"
    end
    on_intel do
      url "https://github.com/zerodha/kite-mcp-server/releases/download/v#{version}/zkite-darwin-amd64.tar.gz"
      sha256 "REPLACE_WITH_SHA256_FOR_DARWIN_AMD64"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/zerodha/kite-mcp-server/releases/download/v#{version}/zkite-linux-arm64.tar.gz"
      sha256 "REPLACE_WITH_SHA256_FOR_LINUX_ARM64"
    end
    on_intel do
      url "https://github.com/zerodha/kite-mcp-server/releases/download/v#{version}/zkite-linux-amd64.tar.gz"
      sha256 "REPLACE_WITH_SHA256_FOR_LINUX_AMD64"
    end
  end

  def install
    bin.install "zkite"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/zkite --version")
  end
end
