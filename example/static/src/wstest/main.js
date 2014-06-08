define([
    'exports',
    'dojox/cometd'
], function(wstest, cometd) {
    var config = {
        contextPath: '/ws'
    };
    wstest.init = function() {
        console.log("we are going");
        cometd.configure({
            url: location.protocol + '//' + location.host + config.contextPath + '/cometd',
            logLevel: 'info'
        });

        cometd.handshake();
    };
});