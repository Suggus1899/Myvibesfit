import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // El Dockerfile copia .next/standalone: sin esto esa carpeta no se genera y
  // la imagen final arranca sin server.js.
  output: "standalone",
};

export default nextConfig;
