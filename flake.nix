{
  description = "alertmanager-signal";

  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" ];
    in {

      packages = nixpkgs.lib.genAttrs systems (system:
        let pkgs = import nixpkgs { inherit system; };
        in {

          # Development environment `nix develop`
          devShells.${system}.default = pkgs.mkShell {
            name = "alertmanager-signal";
            nativeBuildInputs = with pkgs; [
              go gotools gopls go-outline gopkgs gocode-gomod godef golint delve # go tooling
              just vscodium bumpver
            ];
          };

          # Build `nix build .#alertmanager-signal`
          alertmanager-signal = pkgs.buildGoModule {
            pname = "alertmanager-signal";
            version = "1.1.1";
            vendorHash = null;

            src = pkgs.lib.cleanSource ./.;

            buildFlags = [
              "-trimpath"
            ];

            ldflags = [
              "-w" # omit DWARF symbol table
              "-s" # omit symbol table and debug information
              "-extldflags=-static"
            ];

            env = {
              CGO_ENABLED = 0;
            };
          };


          # Container Image `nix build .#image`
          image = pkgs.dockerTools.buildImage {
            name = "docker.io/schlauerlauer/alertmanager-webhook-signal";
            tag = "1.1.1";
            created = "now";
            copyToRoot = pkgs.buildEnv {
              name = "image-root";
              paths = [
                self.packages.${system}.alertmanager-signal
                pkgs.cacert
              ];
            };
            config = {
              Cmd = [ "/bin/alertmanager-signal" ];
              Labels = {
                "org.opencontainers.image.title" = "alertmanager-webhook-signal";
                "org.opencontainers.image.description" = "A simple program to forward Alertmanager and Grafana Alerts to Signal";
                "org.opencontainers.image.source" = "https://github.com/schlauerlauer/alertmanager-webhook-signal";
                "org.opencontainers.image.revision" = "1.1.1";
              };
            };
          };

        }
      );

    };
}
