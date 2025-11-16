{
  description = "A development environment for Go projects.";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };
  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs {
          inherit system;
        };
      in
      {
        devShells.default = pkgs.mkShell {
          hardeningDisable = [ "fortify" ]; # Needed to debug Golang code
          packages = with pkgs; [
            # --- Go Backend Dependencies ---
            go # The Go compiler and toolchain
            gopls # The official Go language server for IDE integration
            golangci-lint # Fast linters Runner for Go
            delve # Debugger for go
            
            # --- GTK4 Dependencies ---
            gtk4 # GTK4 library
            pkg-config # Required for CGo to find libraries
            gobject-introspection # GObject introspection
            glib # GLib library
            cairo # Cairo graphics library
            pango # Text rendering library
            gdk-pixbuf # Image loading library
            graphene # Graphics library used by GTK4
          ];

          shellHook = ''
            # Set up Go environment
            export GOROOT="$(go env GOROOT)"
            export GOPATH="$HOME/go"
            export GOPROXY="https://proxy.golang.org,direct"
            export GOSUMDB="sum.golang.org"
            export PATH="$GOPATH/bin:$GOROOT/bin:$PATH"
            
            # Create GOPATH directories if they don't exist
            mkdir -p "$GOPATH"/{bin,src,pkg}
            
            # GTK4 environment variables
            export PKG_CONFIG_PATH="${pkgs.gtk4.dev}/lib/pkgconfig:${pkgs.glib.dev}/lib/pkgconfig:$PKG_CONFIG_PATH"
            export LD_LIBRARY_PATH="${pkgs.gtk4}/lib:${pkgs.glib}/lib:$LD_LIBRARY_PATH"
            export GI_TYPELIB_PATH="${pkgs.gtk4}/lib/girepository-1.0:${pkgs.gobject-introspection}/lib/girepository-1.0"
            
            echo "--------------------------------------------------"
            echo "  Entering multi-project development environment  "
            echo "--------------------------------------------------"
            echo "Available tools:"
            echo "- Go: $(go version)"
            echo "- GTK: $(pkg-config --modversion gtk4)"
            echo "--------------------------------------------------"
          '';
        };
      }
    );
}