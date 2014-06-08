/*jshint unused:false*/
var dojoConfig = {
	async: true,
	baseUrl: 'http://localhost:8888',
	tlmSiblingOfDojo: false,
	isDebug: true,
	packages: [
		'dojo',
		'dijit',
		'dojox',
		'wstest',
		'org'
	],
	deps: [ 'wstest' ],
	callback: function (wstest) {
		wstest.init();
	}
};
