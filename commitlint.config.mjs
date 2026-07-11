export default {
  extends: ["@commitlint/config-conventional"],
  // Relaxed ruleset (matches Akka.Streams.Http): keep the type discipline that
  // release-please needs, drop the stylistic nits (subject casing, line
  // lengths) that reject otherwise fine commits like "ci(go): Go pipeline".
  rules: {
    "type-enum": [
      2,
      "always",
      [
        "feat",
        "fix",
        "perf",
        "docs",
        "chore",
        "refactor",
        "style",
        "test",
        "ci",
        "build",
        "deps",
      ],
    ],
    "header-max-length": [1, "always", 120],
    "body-max-line-length": [0, "always"],
    "footer-max-line-length": [0, "always"],
    "subject-case": [0, "always"],
  },
  ignores: [(message) => message.includes("dependabot[bot]")],
};
