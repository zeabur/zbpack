module.exports = {
	extends: ["@commitlint/config-conventional"],
	rules: {
		"scope-enum": [2, "always", ["cli", "lib", "planner", "utils", "lint"]],
		"subject-case": [1, "always", ["sentence-case", "start-case"]],
	}
};
