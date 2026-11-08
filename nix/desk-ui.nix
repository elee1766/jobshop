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
      ../ui/vite.config.js
      ../ui/svelte.config.js
      ../ui/jsconfig.json
      ../ui/shell
      ../ui/kit
      ../ui/public
      ../packages/jobshop-client
    ];
  };
  sourceRoot = "source/ui";
  npmDepsHash = "sha256-3fDXmh4FkdU1CvJ7eQ5yze1MFn+kT5ulNDL76H65Jhs=";
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
