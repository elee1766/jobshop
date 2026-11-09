{
  lib,
  buildNpmPackage,
  nodejs_22,
}:

buildNpmPackage {
  pname = "jobshop-desk";
  version = "0.0.0";
  nodejs = nodejs_22;

  src = lib.fileset.toSource {
    root = ../.;
    fileset = lib.fileset.unions [
      ../ui/package.json
      ../ui/package-lock.json
      ../ui/index.html
      ../ui/vite.config.ts
      ../ui/tsconfig.json
      ../ui/src
      ../ui/kit
      ../ui/public
      ../packages/jobshop-client
    ];
  };
  sourceRoot = "source/ui";
  npmDepsHash = "sha256-S1xv+UoTjJFSrmHSiN8HYWL+U77neUzh7SK9L8Ei/zs=";
  npmBuildFlags = [
    "--"
    "--outDir"
    "dist"
  ];

  installPhase = ''
    runHook preInstall
    cp -r dist "$out"
    runHook postInstall
  '';
}
