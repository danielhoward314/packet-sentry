import type { RuntimeEnv } from "./env";

declare global {
  interface Window {
    __ENV__?: RuntimeEnv;
  }
}

export {};
