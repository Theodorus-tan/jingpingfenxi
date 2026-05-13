declare module 'vite-plugin-eslint' {
  import type { Plugin } from 'vite';
  const eslint: (options?: Record<string, unknown>) => Plugin;
  export default eslint;
}
