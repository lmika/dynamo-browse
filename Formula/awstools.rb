# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class Awstools < Formula
  desc "TUI tools for AWS administration"
  homepage "https://audax.tools"
  version "0.0.3"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/lmika/audax/releases/download/v0.0.3/audax_0.0.3_darwin_arm64.tar.gz"
      sha256 "90e38259ef7a092fa111e7c630eef868012ad883f586d12dc3209ffddb642eda"

      def install
        bin.install "dynamo-browse"
      end
    end
    if Hardware::CPU.intel?
      url "https://github.com/lmika/audax/releases/download/v0.0.3/audax_0.0.3_darwin_amd64.tar.gz"
      sha256 "3c6a07e92ca0a0af5a405d8c42daee61d13b8ec215ac991e02f374fc09414c81"

      def install
        bin.install "dynamo-browse"
      end
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      url "https://github.com/lmika/audax/releases/download/v0.0.3/audax_0.0.3_linux_amd64.tar.gz"
      sha256 "1b1a1ca9f5baa93cda863d7f1834a9e482339f5ba40321e80093f94cbe04f020"

      def install
        bin.install "dynamo-browse"
      end
    end
  end
end
