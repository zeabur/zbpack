module.exports = {
	extends: ["@commitlint/config-conventional"],
	rules: {
		"scope-enum-with-subscope": [2, "always"],
		"subject-case": [1, "always", ["sentence-case", "start-case"]],
	},
	plugins: [
		{
			rules: {
				"scope-enum-with-subscope": ({ scope }) => {
					if (!scope || typeof scope !== "string") return [true];

					const scopeSegment = scope.split("/");
					const availableScopes = ["cli", "lib", "planner", "utils", "lint"];

					return [
						availableScopes.includes(scopeSegment[0]),
						`Your scope should be one of the following: ${availableScopes.join(", ")}`,
					];
				},
			},
		},
	],
};
