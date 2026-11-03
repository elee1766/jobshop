# golangci-lint built with Go 1.27: nixpkgs builds it with Go 1.26, and a
# linter built with an older Go refuses to load go 1.27 modules. Keep the
# version equal to the one in .github/workflows/ci.yml.
{
  buildGo127Module,
  fetchFromGitHub,
}:
buildGo127Module (finalAttrs: {
  pname = "golangci-lint";
  version = "2.15.0";
  src = fetchFromGitHub {
    owner = "golangci";
    repo = "golangci-lint";
    tag = "v${finalAttrs.version}";
    hash = "sha256-m9PoJHescRuT9DiKsb6LL+lQlc7RKAWDiRyf8pNMffA=";
  };
  vendorHash = "sha256-gWbhZc/BJwuiwyTMaR1ShEBI1lo0bGLCXsdj5n91baw=";
  subPackages = [ "cmd/golangci-lint" ];
  ldflags = [
    "-s"
    "-w"
    "-X main.version=${finalAttrs.version}"
    "-X main.commit=v${finalAttrs.version}"
    "-X main.date=1970-01-01T00:00:00Z"
  ];
  meta.mainProgram = "golangci-lint";
})
