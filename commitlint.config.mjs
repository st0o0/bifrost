export default {
  extends: ["@commitlint/config-conventional"],
  // Dependabot commits use a non-conventional "deps" type and long changelog
  // bodies (URLs), so they trip type-enum and body-max-line-length. Exempt them;
  // human commits keep the full ruleset.
  ignores: [(message) => message.includes("dependabot[bot]")],
};
