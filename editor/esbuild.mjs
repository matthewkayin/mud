// Builds the extension (node) and the map view webview (browser) bundles into dist/.
// Flags: --watch rebuilds on change, --minify minifies for packaging
import * as esbuild from "esbuild";

const isWatch = process.argv.includes("--watch");
const isMinify = process.argv.includes("--minify");

const sharedOptions = {
    bundle: true,
    sourcemap: true,
    minify: isMinify,
    logLevel: "info",
};

const builds = [
    {
        ...sharedOptions,
        entryPoints: ["src/extension.ts"],
        outfile: "dist/extension.js",
        platform: "node",
        format: "cjs",
        external: ["vscode"],
    },
    {
        ...sharedOptions,
        entryPoints: ["webview/src/main.ts"],
        outfile: "dist/webview.js",
        platform: "browser",
        format: "iife",
    },
];

if (isWatch) {
    for (const options of builds) {
        const context = await esbuild.context(options);
        await context.watch();
    }
} else {
    await Promise.all(builds.map((options) => esbuild.build(options)));
}
