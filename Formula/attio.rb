class Attio < Formula
  desc "Command-line interface for Attio API"
  homepage "https://github.com/exhibiton/attio-cli"
  version "0.1.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/exhibiton/attio-cli/releases/download/v0.1.0/attio_0.1.0_darwin_arm64.tar.gz"
      sha256 "2b6937f905bd6710d8a3177fe521c9aba7dcebbf4bff94f7ee518071b82b3658"
    else
      url "https://github.com/exhibiton/attio-cli/releases/download/v0.1.0/attio_0.1.0_darwin_amd64.tar.gz"
      sha256 "ae6205ca9fc0d370fd170be753355679de42cb509d4e2443297247120b8dddd0"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/exhibiton/attio-cli/releases/download/v0.1.0/attio_0.1.0_linux_arm64.tar.gz"
      sha256 "5659aaa04aac83e5bfa6eaca3a955f35119b9a2fc0e411046b8a5c91ec05eec9"
    else
      url "https://github.com/exhibiton/attio-cli/releases/download/v0.1.0/attio_0.1.0_linux_amd64.tar.gz"
      sha256 "74ddad596677d267989cab284307ff667a7d4316c744cf549832601138b34ca0"
    end
  end

  def install
    bin.install "attio"
  end

  test do
    output = shell_output("#{bin}/attio version")
    assert_match "version", output
  end
end
