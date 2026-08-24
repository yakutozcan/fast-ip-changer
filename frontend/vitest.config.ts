import {defineConfig} from 'vitest/config'

/*
 * A dedicated Vitest config rather than a `test` block in vite.config.ts: the
 * unit suites exercise plain modules under src/lib, so they need neither the
 * React nor the Tailwind plugin and start faster without them.
 *
 * The default environment is `node`. Only the suite that talks to
 * localStorage needs a DOM, and it opts in per file with a
 * `@vitest-environment happy-dom` docblock — cheaper than paying for a
 * document in every suite.
 */
export default defineConfig({
  test: {
    environment: 'node',
    include: ['src/**/*.test.ts'],
    // Restoring stubbed globals between tests keeps navigator/localStorage
    // tampering in one test from leaking into the next.
    unstubGlobals: true,
    restoreMocks: true
  }
})
