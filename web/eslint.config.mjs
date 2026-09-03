import coreWebVitals from "eslint-config-next/core-web-vitals";
import nextTypescript from "eslint-config-next/typescript";

/**
 * ESLint, flat config.
 *
 * `next lint` was removed in Next 16, and the script that called it had been
 * quietly reading "lint" as a directory name and exiting zero -- so this
 * project has had a lint script and no linting for as long as it has been on
 * Next 16.
 *
 * The set is Next's own recommended rules plus its TypeScript layer, and
 * nothing else. Formatting is not linted: there is no Prettier here, and a
 * linter arguing about commas trains everyone to skim its output, which is how
 * the rules that catch real defects get skimmed too.
 */
const config = [
  {
    // Build output and coverage are generated. Linting them reports other
    // people's code and buries our own.
    ignores: [".next/**", "out/**", "coverage/**", "node_modules/**", "next-env.d.ts"],
  },

  ...coreWebVitals,
  ...nextTypescript,

  {
    rules: {
      // An unused binding is usually a leftover, but a deliberately ignored
      // one is a real pattern -- destructuring a field off an object to drop
      // it, or a catch that does not care why. The underscore prefix is how
      // this codebase already says so.
      "@typescript-eslint/no-unused-vars": [
        "error",
        {
          argsIgnorePattern: "^_",
          varsIgnorePattern: "^_",
          caughtErrorsIgnorePattern: "^_",
          destructuredArrayIgnorePattern: "^_",
          ignoreRestSiblings: true,
        },
      ],
    },
  },

  {
    // Tests reach into internals and build shapes the app never builds by
    // hand; a cast there is a fixture, not a hole in the types.
    files: ["**/*.test.ts", "**/*.test.tsx", "src/test/**"],
    rules: {
      "@typescript-eslint/no-explicit-any": "off",
      "@typescript-eslint/no-non-null-assertion": "off",
      // A test may render a bare anchor to check what a primitive does with
      // one. The rule is about navigation in the app, and there is no
      // navigation in jsdom.
      "@next/next/no-html-link-for-pages": "off",
    },
  },
];

export default config;
