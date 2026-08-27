import { cp, mkdir } from "node:fs/promises";

await mkdir("dist", { recursive: true });
for (const file of ["index.html", "app.js", "styles.css"]) {
  await cp(file, `dist/${file}`);
}
