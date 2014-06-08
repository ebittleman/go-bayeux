/*jshint unused:false */
var profile = {
	basePath: '../src/',

	// Builds a new release.
	action: 'release',

	// Strips all comments and whitespace from CSS files and inlines @imports where possible.
	cssOptimize: 'comments',

	// Excludes tests, demos, and original template files from being included in the built version.
	mini: true,

	// Uses Closure Compiler as the JavaScript minifier. This can also be set to "shrinksafe" to use ShrinkSafe,
	// though ShrinkSafe is deprecated and not recommended.
	// This option defaults to "" (no compression) if not provided.
	optimize: '',

	// We're building layers, so we need to set the minifier to use for those, too.
	// This defaults to "shrinksafe" if not provided.
	layerOptimize: 'uglify',

	// Strips all calls to console functions within the code. You can also set this to "warn" to strip everything
	// but console.error, and any other truthy value to strip everything but console.warn and console.error.
	// This defaults to "normal" (strip all but warn and error) if not provided.
	stripConsole: 'all',

	// The default selector engine is not included by default in a dojo.js build in order to make mobile builds
	// smaller. We add it back here to avoid that extra HTTP request. There is also a "lite" selector available; if
	// you use that, you will need to set the `selectorEngine` property in `app/run.js`, too. (The "lite" engine is
	// only suitable if you are not supporting IE7 and earlier.)
	selectorEngine: 'acme',

	// Since we're using dojoConfig.map to patch dojo/_base/declare, we must build anonymous modules
	insertAbsMids: 0,

	// Builds can be split into multiple different JavaScript files called "layers". This allows applications to
	// defer loading large sections of code until they are actually required while still allowing multiple modules to
	// be compiled into a single file.
	layers: {
		'dojo/dojo': {
			include: [
				'dojo/i18n'
			],

			// By default, the build system will try to include `dojo/main` in the built `dojo/dojo` layer, which adds
			// a bunch of stuff we do not want or need. We want the initial script load to be as small and quick to
			// load as possible, so we configure it as a custom, bootable base.
			boot: true,
			customBase: true
		},

		'wstest/main': {
			include: []
		}
	},

	// Providing hints to the build system allows code to be conditionally removed on a more granular level than
	// simple module dependencies can allow. This is especially useful for creating tiny mobile builds.
	// Keep in mind that dead code removal only happens in minifiers that support it! Currently, only Closure Compiler
	// to the Dojo build system with dead code removal.
	// A documented list of has-flags in use within the toolkit can be found at
	// <http://dojotoolkit.org/reference-guide/dojo/has.html>.
	staticHasFeatures: {
		'host-browser': 1,
		'host-node': 0,
		'host-rhino': 0,
		'dojo-firebug': 0,

		// The trace & log APIs are used for debugging the loader, so we do not need them in the build.
		'dojo-trace-api': 0,
		'dojo-log-api': 0,

		// This causes normally private loader data to be exposed for debugging. In a release build, we do not need
		// that either.
		'dojo-publish-privates': 0,

		// This application is pure AMD, so get rid of the legacy loader.
		'dojo-sync-loader': 0,

		// `dojo-xhr-factory` relies on `dojo-sync-loader`, which we have removed.
		'dojo-xhr-factory': 0,

		// We are not loading tests in production, so we can get rid of some test sniffing code.
		'dojo-test-sniff': 0
	}
};
if (typeof module !== 'undefined') { module.exports = profile; }
