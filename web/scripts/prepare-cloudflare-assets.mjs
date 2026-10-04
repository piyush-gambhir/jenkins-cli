import { cp, mkdir, readFile, rm, writeFile } from "node:fs/promises";
import path from "node:path";

const sitePath = process.argv[2];

if (!sitePath || !/^[a-z0-9-]+$/.test(sitePath)) {
  throw new Error("Expected a URL-safe site path argument.");
}

const source = path.resolve("out");
const destinationRoot = path.resolve(".cloudflare/assets");
const destination = path.join(destinationRoot, sitePath);

await rm(destinationRoot, { recursive: true, force: true });
await mkdir(destination, { recursive: true });
await cp(source, destination, { recursive: true });

// Wrangler only reads _redirects from the assets root, and requests arrive with
// the /<site> prefix, so move the file up and prefix every local path.
const nestedRedirects = path.join(destination, "_redirects");
const redirects = await readFile(nestedRedirects, "utf8").catch((error) => {
  if (error.code === "ENOENT") return null;
  throw error;
});
if (redirects !== null) {
  const prefix = (url) => (url.startsWith("/") ? `/${sitePath}${url}` : url);
  const rules = redirects.split("\n").map((line) => {
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith("#")) return line;
    const [from, to, ...rest] = trimmed.split(/\s+/);
    if (!to) throw new Error(`Invalid _redirects rule: ${line}`);
    return [prefix(from), prefix(to), ...rest].join(" ");
  });
  await writeFile(path.join(destinationRoot, "_redirects"), rules.join("\n"));
  await rm(nestedRedirects);
}
