self: super: rec {
  go_1_24 = super.go_1_24.overrideAttrs (old: rec {
    version = "1.24.4";
    src = super.fetchurl {
      url = "https://go.dev/dl/go${version}.src.tar.gz";
      hash = "sha256-WoaoOjH5+oFJC4xUIKw4T9PZWj5x+6Zlx7P5XR3+8rQ=";
    };
  });
  go = go_1_24;
}
